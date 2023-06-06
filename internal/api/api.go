package api

import (
	"github.com/WuKongIM/WuKongChatServer/internal/api/base/event"
	"github.com/WuKongIM/WuKongChatServer/internal/api/channel"
	"github.com/WuKongIM/WuKongChatServer/internal/api/common"
	"github.com/WuKongIM/WuKongChatServer/internal/api/file"
	"github.com/WuKongIM/WuKongChatServer/internal/api/group"
	"github.com/WuKongIM/WuKongChatServer/internal/api/message"
	"github.com/WuKongIM/WuKongChatServer/internal/api/qrcode"
	"github.com/WuKongIM/WuKongChatServer/internal/api/report"
	"github.com/WuKongIM/WuKongChatServer/internal/api/robot"
	"github.com/WuKongIM/WuKongChatServer/internal/api/statistics"
	"github.com/WuKongIM/WuKongChatServer/internal/api/user"
	"github.com/WuKongIM/WuKongChatServer/internal/api/webhook"
	"github.com/WuKongIM/WuKongChatServer/internal/config"
	"github.com/WuKongIM/WuKongChatServer/pkg/register"
	"github.com/WuKongIM/WuKongChatServer/pkg/wkhttp"
	"github.com/robfig/cron"
)

// Route 路由
func Route(r *wkhttp.WKHttp) {
	routes := register.GetRoutes()
	for _, route := range routes {
		route.Route(r)
	}
}

// Init 注册所有api
func Init(ctx *config.Context) {
	// 用户api
	register.Add(user.New(ctx))
	// 用户好友
	register.Add(user.NewFriend(ctx))
	// 消息api
	register.Add(message.New(ctx))
	// 群api
	register.Add(group.New(ctx))
	// 最近会话
	register.Add(message.NewConversation(ctx))
	// webhook
	register.Add(webhook.New(ctx))
	// file
	register.Add(file.New(ctx))
	// qrcode
	register.Add(qrcode.New(ctx))
	// 举报
	register.Add(report.New(ctx))
	// 用户后台管理
	register.Add(user.NewManager(ctx))
	// 举报后台管理
	register.Add(report.NewManager(ctx))
	// 群后台管理
	register.Add(group.NewManager(ctx))
	// 统计管理
	register.Add(statistics.NewStatistics(ctx))
	// 消息管理
	register.Add(message.NewManager(ctx))
	// 通用管理
	register.Add(common.NewManager(ctx))
	// 通用管理
	register.Add(common.New(ctx))
	// 频道相关
	register.Add(channel.New(ctx))
	// 机器人
	register.Add(robot.New(ctx))
	// 机器人管理
	register.Add(robot.NewManager(ctx))

	//开始定时处理事件
	cn := cron.New()
	//定时发布事件 每59秒执行一次
	cn.AddFunc("0/59 * * * * ?", ctx.Event.(*event.Event).EventTimerPush)
	cn.Start()

}
