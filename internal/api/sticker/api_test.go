package sticker

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/WuKongIM/WuKongChatServer/internal/api/base/event"
	"github.com/WuKongIM/WuKongChatServer/internal/config"
	"github.com/WuKongIM/WuKongChatServer/internal/server"
	"github.com/WuKongIM/WuKongChatServer/internal/testutil"
	"github.com/WuKongIM/WuKongChatServer/pkg/util"
	"github.com/stretchr/testify/assert"
)

var uid = "10000"
var token = "token122323"

func newTestServer() (*server.Server, *config.Context) {
	os.Remove("test.db")
	cfg := config.New()
	cfg.Test = true
	cfg.SQLDir = "../../../configs/sql"
	ctx := config.NewContext(cfg)
	ctx.Event = event.New(ctx)
	err := ctx.Cache().Set(cfg.TokenCachePrefix+token, uid+"@test")
	if err != nil {
		panic(err)
	}
	// 创建server
	s := server.New(ctx.GetConfig().Addr, ctx.GetConfig().SSLAddr, ctx.GetConfig().GRPCAddr)
	return s, ctx

}

func TestSearch(t *testing.T) {
	s, ctx := newTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/v1/sticker?keyword=哈哈&page=1&page_size=20", nil)
	assert.NoError(t, err)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddUserCustomSticker(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/v1/sticker/user", bytes.NewReader([]byte(util.ToJson(map[string]interface{}{
		"width":  230,
		"height": 300,
		"path":   "xx111",
	}))))
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserCustomSticker(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	_, err = f.db.insertUserSticker(&customModel{
		Width:   200,
		Height:  300,
		Path:    "sdd",
		UID:     testutil.UID,
		SortNum: 1,
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/v1/sticker/user", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"path":"sdd"`))
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"height":300`))
}

func TestDeleteCustomSticker(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	_, err = f.db.insertUserSticker(&customModel{
		Width:   200,
		Height:  300,
		Path:    "sdd",
		UID:     testutil.UID,
		SortNum: 1,
	})
	assert.NoError(t, err)
	list := make([]string, 0)
	list = append(list, "sdd")
	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", "/v1/sticker/user", bytes.NewReader([]byte(util.ToJson(map[string]interface{}{
		"paths": list,
	}))))
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserDeleteByCategory(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	_, err = f.db.insertUserCategory(&categoryModel{
		UID:      testutil.UID,
		SortNum:  1,
		Category: "111",
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", "/v1/sticker/remove", bytes.NewReader([]byte(util.ToJson(map[string]interface{}{
		"category": "111",
	}))))
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserAddByCategory(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	_, err = f.db.insertSticker(&model{
		Category:       "111",
		Width:          200,
		Height:         200,
		Title:          "t11",
		Path:           "2233",
		SearchableWord: "t11",
		UserCustom:     0,
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/v1/sticker/user/111", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserCategory(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	_, err = f.db.insertUserCategory(&categoryModel{
		Category: "111",
		UID:      testutil.UID,
		SortNum:  3,
	})
	assert.NoError(t, err)
	_, err = f.db.insertUserCategory(&categoryModel{
		Category: "2232",
		UID:      testutil.UID,
		SortNum:  1,
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/v1/sticker/user/category", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"category":"2232"`))
}

func TestGetStickerWithCategory(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	_, err = f.db.insertSticker(&model{
		Category: "111",
		Path:     "xxx",
		Title:    "t11",
		Width:    200,
		Height:   300,
	})
	assert.NoError(t, err)
	_, err = f.db.insertSticker(&model{
		Category: "111",
		Path:     "sss",
		Title:    "tss",
		Width:    200,
		Height:   300,
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/v1/sticker/user/sticker?category=111", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"category":"2232"`))
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"category":"sss"`))
}

func TestStoreList(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	_, err = f.db.insertStore(&storeModel{
		Title:    "tss",
		Category: "css",
		Cover:    "coverss",
		Desc:     "dss",
	})
	assert.NoError(t, err)
	_, err = f.db.insertStore(&storeModel{
		Title:    "tss1",
		Category: "css1",
		Cover:    "coverss1",
		Desc:     "dss1",
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/v1/sticker/store?page_index=2&page_size=10", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"cover":"css1"`))
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"category":"dss1"`))
}
