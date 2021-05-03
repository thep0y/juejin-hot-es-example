/*
 * @Author: thepoy
 * @Email: thepoy@163.com
 * @File Name: index_test.go
 * @Created: 2021-04-24 16:14:08
 * @Modified: 2021-05-02 20:41:54
 */

package commands

import (
	"testing"
)

func TestCreateCursor(t *testing.T) {
	body := newRequestBody(0)
	body.createCursor()

	b, e := body.marshal()
	if e != nil {
		t.Errorf("Error marshalling request body: %s", e)
	}

	t.Logf("Cursor: %s", b)
}
