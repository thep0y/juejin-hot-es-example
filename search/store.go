/*
 * @Author: thepoy
 * @Email: thepoy@163.com
 * @File Name: store.go
 * @Created: 2021-04-24 16:14:08
 * @Modified: 2021-05-04 10:15:01
 */

package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// SearchResults es 查询响应
type SearchResults struct {
	Total int    `json:"total"`
	Hits  []*Hit `json:"hits"`
}

// Hit 查询响应中的单条结果
type Hit struct {
	Document
	Sort       []interface{} `json:"sort"`
	Highlights *struct {
		Title    []string `json:"title"`
		Brief    []string `json:"brief_content"`
		Category []string `json:"category"`
		Company  []string `json:"author_info.company"`
	} `json:"highlights,omitempty"`
}

// StoreConfig 存储配置
type StoreConfig struct {
	Client    *elasticsearch.Client
	IndexName string
}

// Store 存储结构体，用来索引和查询文档
type Store struct {
	es        *elasticsearch.Client
	indexName string
}

// NewStore 创建一个新的 Store 实例
func NewStore(c StoreConfig) (*Store, error) {
	indexName := c.IndexName
	if indexName == "" {
		indexName = "juejin-hot"
	}

	s := Store{
		es:        c.Client,
		indexName: indexName,
	}

	return &s, nil
}

// CreateIndex 用给定的格式创建索引
func (s *Store) CreateIndex(mapping string) error {
	res, err := s.es.Indices.Create(
		s.indexName,
		s.es.Indices.Create.WithBody(strings.NewReader(mapping)),
	)
	if err != nil {
		return err
	}

	if res.IsError() {
		return fmt.Errorf("Error: %s", res)
	}

	return nil
}

// Create 创建新文档到 Store 中
func (s *Store) Create(item *Document) error {
	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}

	ctx := context.Background()
	res, err := esapi.CreateRequest{
		Index:      s.indexName,
		DocumentID: item.ID,
		Body:       bytes.NewReader(payload),
	}.Do(ctx, s.es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return err
		}
		err := e["error"].(map[string]interface{})
		return fmt.Errorf("[%s] %s: %s", res.Status(), err["type"], err["reason"])
	}

	return nil
}

// Exists 当 id 对应的文档在 Store 中存在时，返回 true
func (s *Store) Exists(id string) (bool, error) {
	res, err := s.es.Exists(s.indexName, id)
	if err != nil {
		return false, err
	}

	switch res.StatusCode {
	case 200:
		return true, nil
	case 404:
		return false, nil
	default:
		return false, fmt.Errorf("[%s]", res.Status())
	}
}

// Delete 删除指定 id 的文档
func (s *Store) Delete(id string) (bool, error) {
	res, err := s.es.Delete(s.indexName, id)
	if err != nil {
		return false, err
	}

	switch res.StatusCode {
	case 200:
		return true, nil
	case 404:
		return false, nil
	default:
		return false, fmt.Errorf("[%s]", res.Status())
	}
}

// Search 根据给定的查询词 query 查询，after 为可选参数
func (s *Store) Search(query string, after ...string) (*SearchResults, error) {
	var results SearchResults

	res, err := s.es.Search(
		s.es.Search.WithIndex(s.indexName),
		s.es.Search.WithBody(s.buildQuery(query, after...)),
	)
	if err != nil {
		return &results, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return &results, err
		}
		return &results, fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
	}

	type envelopeResponse struct {
		Took int
		Hits struct {
			Total struct {
				Value int
			}
			Hits []struct {
				ID         string          `json:"_id"`
				Source     json.RawMessage `json:"_source"`
				Highlights json.RawMessage `json:"highlight"`
				Sort       []interface{}   `json:"sort"`
			}
		}
	}

	var r envelopeResponse
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return &results, err
	}

	results.Total = r.Hits.Total.Value

	if len(r.Hits.Hits) < 1 {
		results.Hits = []*Hit{}
		return &results, nil
	}

	for _, hit := range r.Hits.Hits {
		var h Hit
		h.ID = hit.ID
		h.Sort = hit.Sort

		if err := json.Unmarshal(hit.Source, &h); err != nil {
			return &results, err
		}

		if len(hit.Highlights) > 0 {
			if err := json.Unmarshal(hit.Highlights, &h.Highlights); err != nil {
				return &results, err
			}
		}
		results.Hits = append(results.Hits, &h)
	}

	return &results, nil
}

func (s *Store) buildQuery(query string, after ...string) io.Reader {
	var b strings.Builder

	b.WriteString("{\n")

	if query == "" {
		b.WriteString(searchAll)
	} else {
		b.WriteString(fmt.Sprintf(searchMatch, query))
	}

	if len(after) > 0 && after[0] != "" && after[0] != "null" {
		b.WriteString(",\n")
		b.WriteString(fmt.Sprintf(`	"search_after": %s`, after))
	}

	b.WriteString("\n}")

	return strings.NewReader(b.String())
}

const searchAll = `
	"query" : { "match_all" : {} },
	"size" : 25,
	"sort" : { "create_time" : "desc", "_doc" : "asc" }`

const searchMatch = `
	"query" : {
		"multi_match" : {
			"query" : %q,
			"fields" : ["title^100", "brief_content^100", "category", "author_info.company"],
			"operator" : "and",
			"type":"phrase"
		}
	},
	"highlight" : {
		"fields" : {
			"title" : { "number_of_fragments" : 0 },
			"brief_content" : { "number_of_fragments" : 3, "fragment_size" : 25 },
			"category" : { "number_of_fragments" : 0 },
			"author_info.company" : { "number_of_fragments" : 0 }
		}
	},
	"size" : 25,
	"sort" : [ { "_score" : "desc" }, { "_doc" : "asc" } ]`
