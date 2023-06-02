package message

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/WuKongIM/WuKongChatServer/internal/api/base/event"
	"github.com/WuKongIM/WuKongChatServer/internal/config"
	"github.com/WuKongIM/WuKongChatServer/internal/server"
	"github.com/WuKongIM/WuKongChatServer/pkg/util"
	"github.com/stretchr/testify/assert"
)

var uid = "10000"

// var friendUID = "10001"
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

func TestMessageSync(t *testing.T) {
	s, ctx := newTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())

	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/v1/message/sync", bytes.NewReader([]byte(util.ToJson(map[string]interface{}{
		"uid":             uid,
		"max_message_seq": 100,
		"limit":           100,
	}))))
	req.Header.Set("token", token)
	assert.NoError(t, err)
	s.GetRoute().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	time.Sleep(time.Millisecond * 200)

}

func TestMessageSyncack(t *testing.T) {
	s, ctx := newTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())

	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/v1/message/syncack/111", bytes.NewReader([]byte(util.ToJson(map[string]interface{}{
		"uid":              uid,
		"last_message_seq": 100,
	}))))
	req.Header.Set("token", token)
	assert.NoError(t, err)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	time.Sleep(time.Millisecond * 200)

}
