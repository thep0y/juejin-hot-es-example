/*
 * @Author: thepoy
 * @Email: thepoy@163.com
 * @File Name: search.go
 * @Created:  2021-05-03 08:26:49
 * @Modified: 2021-05-03 19:26:35
 */

package commands

import (
	"elasticsearch/juejinhot/search"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search juejin hot recommended articles",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(os.Stdout, "\x1b[2m%s\x1b[0m\n", strings.Repeat("━", tWidth))
		fmt.Fprintf(os.Stdout, "\x1b[2m?q=\x1b[0m\x1b[1m%s\x1b[0m\n", strings.Join(args, " "))
		fmt.Fprintf(os.Stdout, "\x1b[2m%s\x1b[0m\n", strings.Repeat("━", tWidth))

		es, err := elasticsearch.NewDefaultClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "\x1b[1;107;41mERROR: %s\x1b[0m\n", err)
		}

		config := search.StoreConfig{
			Client:    es,
			IndexName: IndexName,
		}
		store, err := search.NewStore(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\x1b[1;107;41mERROR: %s\x1b[0m\n", err)
			os.Exit(1)
		}
		s := Search{
			store:        store,
			reHightlight: regexp.MustCompile("<em>(.+?)</em>"),
		}

		results, err := s.getResults(strings.Join(args, " "))
		if err != nil {
			fmt.Fprintf(os.Stderr, "\x1b[1;107;41mERROR: %s\x1b[0m\n", err)
			os.Exit(1)
		}

		if results.Total < 1 {
			fmt.Fprintln(os.Stdout, "⨯ No results")
			fmt.Fprintf(os.Stdout, "\x1b[2m%s\x1b[0m\n", strings.Repeat("─", tWidth))
			os.Exit(0)
		}

		for _, result := range results.Hits {
			s.displayResult(os.Stdout, result)
		}
	},
}

// Search 获取并显示查询到的结果
type Search struct {
	store        *search.Store
	reHightlight *regexp.Regexp
}

func (s *Search) getResults(query string) (*search.SearchResults, error) {
	return s.store.Search(query)
}

func (s *Search) displayResult(w io.Writer, hit *search.Hit) {
	var title string
	if len(hit.Highlights.Title) > 0 {
		title = hit.Highlights.Title[0]
	} else {
		title = hit.Title
	}

	var brief string
	if len(hit.Highlights.Brief) > 0 {
		brief = hit.Highlights.Brief[0]
	} else {
		brief = hit.Brief
	}

	var category string
	if len(hit.Highlights.Category) > 0 {
		category = hit.Highlights.Category[0]
	} else {
		category = hit.Category
	}

	var company string
	if len(hit.Highlights.Company) > 0 {
		company = hit.Highlights.Company[0]
	} else {
		company = hit.Author.Company
	}

	fmt.Fprintf(w, "\x1b[2m• title: \x1b[0m \x1b[1m%s\x1b[0m\n", s.highlightString(title))
	fmt.Fprintf(w, "  \x1b[2mbrief_content:\x1b[0m \x1b[1m%s\n", s.highlightString(brief))
	fmt.Fprintf(w, "  \x1b[2mcategory:\x1b[0m \x1b[1m%s \n", s.highlightString(category))
	fmt.Fprintf(w, "  \x1b[2mcompany:\x1b[0m \x1b[1m%s\n", s.highlightString(company))
}

func (s *Search) highlightString(input string) string {
	return s.reHightlight.ReplaceAllString(input, "\x1b[30;47m$1\x1b[0m")
}
