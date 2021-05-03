/*
 * @Author: thepoy
 * @Email: thepoy@163.com
 * @File Name: index.go
 * @Created: 2021-04-24 16:14:08
 * @Modified: 2021-05-03 19:40:37
 */

package commands

import (
	"elasticsearch/juejinhot/search"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	indexSetup bool
	endPage    int

	url = "https://api.juejin.cn/recommend_api/v1/article/recommend_all_feed"
)

func init() {
	rootCmd.AddCommand(indexCmd)
	indexCmd.Flags().BoolVar(&indexSetup, "setup", false, "Create Elasticsearch index")
	indexCmd.Flags().IntVar(&endPage, "pages", 5, "The count of pages you want to crawl")
}

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index juejin hot-recommended articles into Elasticsearch",
	Run: func(cmd *cobra.Command, args []string) {
		crawler := Crawler{
			log: zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
				Level(func() zerolog.Level {
					if os.Getenv("DEBUG") != "" {
						return zerolog.DebugLevel
					} else {
						return zerolog.InfoLevel
					}
				}()).
				With().
				Timestamp().
				Logger(),
		}

		es, err := elasticsearch.NewDefaultClient()
		if err != nil {
			crawler.log.Fatal().Err(err).Msg("Error creating Elasticsearch client")
		}

		config := search.StoreConfig{Client: es, IndexName: IndexName}
		store, err := search.NewStore(config)
		if err != nil {
			crawler.log.Fatal().Err(err).Msg("Cannot create store")
		}
		crawler.store = store

		if indexSetup {
			crawler.log.Info().Msg("Creating index with mapping")
			if err := crawler.setupIndex(); err != nil {
				crawler.log.Fatal().Err(err).Msg("Cannot create Elasticsearch index")
			}
		}

		crawler.log.Info().Msgf("Starting the crawl with %d workers at 0 offset", crawler.workers)
		crawler.Run()
	},
}

// requestBody 请求体
type requestBody struct {
	ClientType int    `json:"client_type"`
	Cursor     string `json:"cursor"`
	IDType     int    `json:"id_type"`
	Limit      int    `json:"limit"`
	SortType   int    `json:"sort_type"`
	offset     int
}

// newRequestBody 用 offset 创建新的请求体
func newRequestBody(offset int) *requestBody {
	body := new(requestBody)
	body.Limit = 20
	body.ClientType = 2608
	body.IDType = 2
	body.SortType = 200
	body.offset = offset
	return body
}

// marshal 序列化请求体
func (r *requestBody) marshal() ([]byte, error) {
	if r.Cursor == "" {
		return nil, errors.New("cursor cannot be null, you should execute createCursor(offset int) first")
	}
	return json.Marshal(r)
}

// createCursor 使用请求体内的 offset 构建请求体的 Cursor
func (r *requestBody) createCursor() {
	var cursor string
	if r.offset == 0 {
		r.Cursor = "0"
	} else {
		cursor = fmt.Sprintf(`{"v":"6956728664562073630","i":%d}`, r.offset)
		r.Cursor = base64.StdEncoding.EncodeToString([]byte(cursor))
	}
}

// Crawler 爬虫结构体
type Crawler struct {
	store *search.Store
	log   zerolog.Logger

	workers int
	queue   chan int

	Offset int
}

// Run 启动爬虫
func (c *Crawler) Run() {
	var wg sync.WaitGroup

	rand.Seed(time.Now().Unix())

	// 每一页开启一个协程
	for i := 0; i < endPage; i++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()

			c.ProcessPage(page * 20)
		}(i)
	}

	wg.Wait()
}

// ProcessPage 掘金热门，每页 20 条，用 post 方式发送请求
func (c *Crawler) ProcessPage(offset int) []search.Document {
	c.log.Debug().Int("Page", offset/20+1).Msg("Processing offset")

	body := newRequestBody(offset)
	body.createCursor()

	bodyBytes, err := body.marshal()
	if err != nil {
		c.log.Error().Err(err).Msg("Error marshalling request body")
		os.Exit(1)
	}

	res, err := c.post(string(bodyBytes))
	if err != nil {
		c.log.Error().Err(err).Msg("Error posting")
		os.Exit(1)
	}

	docs, err := c.processResponse(res)
	if err != nil {
		c.log.Error().Err(err).Int("Offset", offset).Msg("Error processing response")
	} else {
		for _, doc := range docs {
			if ok, err := c.existsDocument(doc.ID); ok {
				if err != nil {
					c.log.Fatal().Err(err).Str("Article ID", doc.ID).Msg("Error skipping existing doc")
				} else {
					c.log.Info().Str("ID", doc.ID).Msg("Skipping existing doc")
				}
				continue
			}
			err = c.storeDocument(&doc)
			if err != nil {
				c.log.Error().Err(err).Int("Offset", offset).Msg("Error storing doc")
			} else {
				c.log.Info().Str("Article ID", doc.ID).Str("title", doc.Title).Msg("Stored doc")
			}
		}
	}

	time.Sleep(time.Duration(rand.Intn(100)+100) * time.Millisecond)

	return docs
}

// post 请亚前设置 Content-Type 为 application/json，然后返回响应
func (c *Crawler) post(body string) (*http.Response, error) {
	client := http.DefaultClient

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		c.log.Error().Err(err).Msg("Error creating new request")
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")

	return client.Do(req)
}

type jsonResponseItem struct {
	ItemInfo struct {
		ArticleInfo struct {
			ArticleID  string `json:"article_id"`
			Title      string `json:"title"`
			Brief      string `json:"brief_content"`
			CreateTime string `json:"ctime"`
		} `json:"article_info"`
		AuthorInfo struct {
			UserName    string `json:"user_name"`
			Company     string `json:"company"`
			JobTitle    string `json:"job_title"`
			Description string `json:"description"`
		} `json:"author_user_info"`
		Category struct {
			CategoryName string `json:"category_name"`
		} `json:"category"`
	} `json:"item_info"`
}

type jsonResponse struct {
	Data []jsonResponseItem `json:"data"`
}

func (c *Crawler) processResponse(res *http.Response) ([]search.Document, error) {
	defer res.Body.Close()

	var j jsonResponse
	if err := json.NewDecoder(res.Body).Decode(&j); err != nil {
		return nil, err
	}

	docs := make([]search.Document, 20)
	for i := 0; i < 20; i++ {
		docs[i].ID = j.Data[i].ItemInfo.ArticleInfo.ArticleID
		docs[i].Author = j.Data[i].ItemInfo.AuthorInfo
		docs[i].Brief = j.Data[i].ItemInfo.ArticleInfo.Brief
		docs[i].Category = j.Data[i].ItemInfo.Category.CategoryName
		docs[i].Title = j.Data[i].ItemInfo.ArticleInfo.Title
		docs[i].CreateTime = j.Data[i].ItemInfo.ArticleInfo.CreateTime
	}

	return docs, nil
}

func (c *Crawler) storeDocument(doc *search.Document) error {
	return c.store.Create(doc)
}

func (c *Crawler) existsDocument(id string) (bool, error) {
	ok, err := c.store.Exists(id)
	if err != nil {
		return false, fmt.Errorf("store: %s", err)
	}
	return ok, nil
}

func (c *Crawler) setupIndex() error {
	mapping := `{
		"mappings": {
			"_doc": {
				"properties": {
					"id":         { "type": "keyword" },
					"title":      { "type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart" },
					"brief_content":        { "type": "text", "analyzer": "ik_max_word", "search_analyzer": "ik_smart" },
					"category": { "type": "keyword", "analyzer": "ik_max_word", "search_analyzer": "ik_smart" },
					"author_info": {
						"properties": {
							"user_name": { "type": "string" },
							"company": { "type": "string", "analyzer": "ik_max_word", "search_analyzer": "ik_smart" },
							"job_title": { "type": "string" },
							"description": { "type": "text" }
						}
					}
				}
			}
		}
	}`
	return c.store.CreateIndex(mapping)
}
