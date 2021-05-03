/*
 * @Author: thepoy
 * @Email: thepoy@163.com
 * @File Name: root.go
 * @Created:  2021-05-02 20:23:41
 * @Modified: 2021-05-03 19:33:02
 */

package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	// IndexName 索引名
	IndexName string
	tWidth    int
)

var rootCmd = &cobra.Command{
	Use:   "juejin",
	Short: "juejin allows you to index and search hot-recommended article's titles",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&IndexName, "index", "i", "juejin", "Index name")
	tWidth, _, _ = terminal.GetSize(int(os.Stderr.Fd()))
}

// Execute 启动命令行程序
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
