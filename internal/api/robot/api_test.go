package robot

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/WuKongIM/WuKongChatServer/internal/api/base/event"
	"github.com/WuKongIM/WuKongChatServer/internal/config"
	"github.com/WuKongIM/WuKongChatServer/internal/server"
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
func TestSyncRobot(t *testing.T) {
	s, ctx := newTestServer()
	f := New(ctx)
	f.Route(s.GetRoute())

	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/v1/robot/sync", bytes.NewReader([]byte(util.ToJson([]map[string]interface{}{
		{
			"robot_id": ctx.GetConfig().SystemUID,
			"version":  0,
		},
	}))))
	assert.NoError(t, err)
	req.Header.Set("token", token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMention(t *testing.T) {

	reg := regexp.MustCompile(`@\S+`)

	fmt.Println(reg.FindAllString("dsds @增加啊每个萨摩 你好", -1))
}
