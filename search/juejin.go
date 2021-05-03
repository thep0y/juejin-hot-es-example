/*
 * @Author: thepoy
 * @Email: thepoy@163.com
 * @File Name: juejin.go
 * @Created: 2021-04-24 16:14:08
 * @Modified: 2021-05-03 09:14:17
 */

package search

import "encoding/json"

type Document struct {
	ID         string `json:"article_id"`
	Title      string `json:"title"`
	Brief      string `json:"brief_content"`
	Category   string `json:"category"`
	CreateTime string `json:"create_time"`
	Author     Author `json:"author_info"`
}

type Author struct {
	UserName    string `json:"user_name"`
	Company     string `json:"company"`
	JobTitle    string `json:"job_title"`
	Description string `json:"description"`
}

func (d Document) String() string {
	j, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}

	return string(j)
}
