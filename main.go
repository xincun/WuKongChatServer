package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/WuKongIM/WuKongChatServer/internal/api"
	"github.com/WuKongIM/WuKongChatServer/internal/api/base/event"
	"github.com/WuKongIM/WuKongChatServer/internal/config"
	"github.com/WuKongIM/WuKongChatServer/internal/server"
	"github.com/WuKongIM/WuKongChatServer/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/judwhite/go-svc"
)

// go ldflags
var Version string    // version
var Commit string     // git commit id
var CommitDate string // git commit date
var TreeState string  // git tree state

func main() {

	gin.SetMode(gin.ReleaseMode)

	cfg := config.New()
	cfg.Version = Version
	// 初始化context
	ctx := config.NewContext(cfg)
	ctx.Event = event.New(ctx)

	logOpts := log.NewOptions()
	logOpts.Level = cfg.Logger.Level
	logOpts.LineNum = cfg.Logger.LineNum
	log.Configure(logOpts)

	var serverType string
	if len(os.Args) > 1 {
		serverType = strings.TrimSpace(os.Args[1])
	}

	if serverType == "api" || serverType == "" { // api服务启动
		runAPI(ctx)
	}

}

func runAPI(ctx *config.Context) {
	// 创建server
	s := server.New(ctx.GetConfig().Addr, ctx.GetConfig().SSLAddr, ctx.GetConfig().GRPCAddr)
	ctx.Server = s
	// 替换web下的配置文件
	replaceWebConfig(ctx.GetConfig())
	// 初始化api
	api.Init(ctx)
	s.GetRoute().UseGin(ctx.Tracer().GinMiddle()) // 需要放在 api.Route(s.GetRoute())的前面
	s.GetRoute().UseGin(func(c *gin.Context) {
		ingorePaths := ingorePaths()
		for _, ingorePath := range ingorePaths {
			if ingorePath == c.FullPath() {
				return
			}
		}
		gin.Logger()(c)
	})
	// 开始route
	api.Route(s.GetRoute())

	// 打印服务器信息
	printServerInfo(ctx)

	// 运行
	err := svc.Run(s)
	if err != nil {
		panic(err)
	}
}

func printServerInfo(ctx *config.Context) {
	infoStr := `
[?25l[?7lffffffffffffttttttffffffffffff
fffffffftffLLCCCCLLfftffffffff
fffffftLG0@@@@@@@@@@0GLtffffff
fffftL0@@@@@@@@@@@@@@@@0Ltffff
ffftC@88@@@@@@88@@@@@@88@Ctfff
fffL@0tt8@@@@8tt0@@@@8tt0@Lfff
fft0@@8GG@@@@G00G@@@@0G8@@0tff
fft0@@@@GG@@GG@@0G@@0G@@@@0tff
fftC@@@@@G00G@@@@GG0G@@@@@Ctff
ffftG@@@@0tt8@@@@8tt0@@@@Gtfff
fffftL8@@@88@@@@@@88@@@8Ltffff
fffffftG@@@@@@@@@@@@8GLftfffff
fffffftC@8GCCGGGGCLffttfffffff
fffffffLLftttttttttfffffffffff
ffffffffffffffffffffffffffffff[0m
[15A[9999999D[33C[0m[1m[32mWuKongChat is running[0m 
[33C[0m---------------------[0m 
[33C[0m[1m[33mMode[0m[0m:[0m #mode#[0m 
[33C[0m[1m[33mApp name[0m[0m:[0m #appname#[0m 
[33C[0m[1m[33mVersion[0m[0m:[0m #version#[0m 
[33C[0m[1m[33mGit[0m[0m:[0m #git#[0m 
[33C[0m[1m[33mGo build[0m[0m:[0m #gobuild#[0m 
[33C[0m[1m[33mIM URL[0m[0m:[0m #imurl#[0m 
[33C[0m[1m[33mFile Upload URL[0m[0m:[0m #uploadURL#[0m 
[33C[0m[1m[33mFile Download URL[0m[0m:[0m #downloadURL#[0m 
[33C[0m[1m[33mThe API is listening at[0m[0m:[0m #apiAddr#[0m 

[33C[30m[40m   [31m[41m   [32m[42m   [33m[43m   [34m[44m   [35m[45m   [36m[46m   [37m[47m   [m
[33C[38;5;8m[48;5;8m   [38;5;9m[48;5;9m   [38;5;10m[48;5;10m   [38;5;11m[48;5;11m   [38;5;12m[48;5;12m   [38;5;13m[48;5;13m   [38;5;14m[48;5;14m   [38;5;15m[48;5;15m   [m


[?25h[?7h
	`
	infoStr = strings.Replace(infoStr, "#mode#", string(ctx.GetConfig().Mode), -1)
	infoStr = strings.Replace(infoStr, "#appname#", ctx.GetConfig().AppName, -1)
	infoStr = strings.Replace(infoStr, "#version#", ctx.GetConfig().Version, -1)
	infoStr = strings.Replace(infoStr, "#git#", fmt.Sprintf("%s-%s", CommitDate, Commit), -1)
	infoStr = strings.Replace(infoStr, "#gobuild#", runtime.Version(), -1)
	infoStr = strings.Replace(infoStr, "#imurl#", ctx.GetConfig().IMURL, -1)
	infoStr = strings.Replace(infoStr, "#uploadURL#", ctx.GetConfig().UploadURL, -1)
	infoStr = strings.Replace(infoStr, "#downloadURL#", ctx.GetConfig().FileDownloadURL, -1)
	infoStr = strings.Replace(infoStr, "#apiAddr#", ctx.GetConfig().Addr, -1)
	fmt.Println(infoStr)
}

func ingorePaths() []string {

	return []string{
		"/v1/robots/:robot_id/:app_key/getEvents",
		"/v1/ping",
	}
}

func replaceWebConfig(cfg *config.Config) {
	path := "./configs/web/js/config.js"
	newConfigContent := fmt.Sprintf(`const apiURL = "%s/"`, cfg.APIBaseURL)
	ioutil.WriteFile(path, []byte(newConfigContent), 0)

}
