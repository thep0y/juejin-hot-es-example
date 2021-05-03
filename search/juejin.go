/*
 * @Author: thepoy
 * @Email: thepoy@163.com
 * @File Name: juejin.go
 * @Created: 2021-04-24 16:14:08
 * @Modified: 2021-05-03 20:32:09
 */

package search

import "encoding/json"

// Document 掘金热门推荐中部分信息构成的文档
type Document struct {
	ID         string `json:"article_id"`
	Title      string `json:"title"`
	Brief      string `json:"brief_content"`
	Category   string `json:"category"`
	CreateTime string `json:"create_time"`
	Author     Author `json:"author_info"`
}

// Author 每篇文章的作者信息
type Author struct {
	UserName    string `json:"user_name"`
	Company     string `json:"company"`
	JobTitle    string `json:"job_title"`
	Description string `json:"description"`
}

// String 将文档以序列化为字符串
func (d Document) String() string {
	j, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}

	return string(j)
}
