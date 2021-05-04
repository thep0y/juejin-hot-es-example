/*
 * @Author: thepoy
 * @Email: thepoy@163.com
 * @File Name: delete.go
 * @Created: 2021-04-24 16:14:08
 * @Modified: 2021-05-04 10:24:15
 */

package commands

import (
	"elasticsearch/juejinhot/search"
	"fmt"
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete item with id",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(os.Stdout, "\x1b[2m%s\x1b[0m\n", strings.Repeat("━", tWidth))
		fmt.Fprintf(os.Stdout, "\x1b[2mid=\x1b[0m\x1b[1m%s\x1b[0m\n", strings.Join(args, " "))
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

		del := Delete{store}

		ok, err := del.delete(strings.Join(args, " "))
		if !ok {
			if err != nil {
				fmt.Fprintf(os.Stderr, "\x1b[1;107;41mERROR: %s\x1b[0m\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "\x1b[1;107;41mERROR: Not found: id=%s\x1b[0m\n", strings.Join(args, " "))
			}
		} else {
			fmt.Fprintf(os.Stdout, "\x1b[2mDeleted: id=\x1b[0m\x1b[1m%s\x1b[0m\n", strings.Join(args, " "))
		}
	},
}

// Delete 封装删除文档方法的结构体
type Delete struct {
	store *search.Store
}

func (d *Delete) delete(id string) (bool, error) {
	return d.store.Delete(id)
}
