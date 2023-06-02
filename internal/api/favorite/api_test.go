package favorite

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/WuKongIM/WuKongChatServer/internal/testutil"
	"github.com/WuKongIM/WuKongChatServer/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestFavorite_Add(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	favorite := &favoriteReq{
		Type:       1,
		UniqueKey:  "0011",
		AuthorName: "sl",
		AuthorUID:  "sluid",
		Payload: map[string]interface{}{
			"content": "这是文本",
		},
	}
	req, _ := http.NewRequest("POST", "/v1/favorites", bytes.NewReader([]byte(util.ToJson(favorite))))
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFavorite_List(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	content := map[string]interface{}{
		"content": "这是文本信息",
	}
	b, err := json.Marshal(content)
	err = f.db.InsertFavorite(&Model{
		Type:       1,
		AuthorName: "sl",
		AuthorUID:  "sluid",
		Payload:    string(b[:]),
		UID:        "10000",
		UniqueKey:  "fkey",
	})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/favorite/my?page_size=1&page_index=1", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"unique_key":"fkey"`))
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"author_name":"sl"`))
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"author_uid":"sluid"`))
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"type":1`))
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"content":"这是文本信息"`))
}

func TestFavorite_Delete(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)

	err = f.db.InsertFavorite(&Model{
		Type:       1,
		AuthorName: "sl",
		AuthorUID:  "sluid",
		Payload:    "{'content':'这是文本'}",
		UID:        "10000",
		UniqueKey:  "fkey",
	})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/favorites/10000", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
