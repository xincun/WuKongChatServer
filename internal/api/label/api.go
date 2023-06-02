package label

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/WuKongIM/WuKongChatServer/internal/api/user"
	"github.com/WuKongIM/WuKongChatServer/internal/config"
	"github.com/WuKongIM/WuKongChatServer/pkg/log"
	"github.com/WuKongIM/WuKongChatServer/pkg/wkhttp"
	"go.uber.org/zap"
)

// Label 标签
type Label struct {
	ctx *config.Context
	log.Log
	db          *DB
	userService user.IService
}

// New  New
func New(ctx *config.Context) *Label {
	return &Label{
		ctx:         ctx,
		db:          newDB(ctx.DB()),
		Log:         log.NewTLog("label"),
		userService: user.NewService(ctx),
	}
}

// Route 路由配置
func (l *Label) Route(r *wkhttp.WKHttp) {
	labels := r.Group("/v1/label", r.AuthMiddleware(l.ctx.Cache(), l.ctx.GetConfig().TokenCachePrefix))
	{
		labels.POST("", l.add)          // 添加标签
		labels.DELETE("/:id", l.delete) //删除标签
		labels.PUT("/:id", l.update)    //修改标签
		labels.GET("", l.list)          //标签列表
		labels.GET("/:id", l.detail)    //标签详情
	}
}

// add 添加标签
func (l *Label) add(c *wkhttp.Context) {
	loginUID := c.MustGet("uid").(string)
	var req labelReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}

	if len(strings.TrimSpace(req.Name)) == 0 {
		c.ResponseError(errors.New("标签名字不能为空"))
		return
	}

	if len(req.MemberUIDs) == 0 {
		c.ResponseError(errors.New("标签成员不能为空"))
		return
	}

	label := &model{
		UID:        loginUID,
		MemberUids: strings.Join(req.MemberUIDs, ","),
		Name:       req.Name,
	}
	labelID, err := l.db.insert(label)
	if err != nil {
		l.Error("添加标签失败", zap.Error(err))
		c.ResponseError(errors.New("添加标签失败"))
		return
	}
	c.JSON(http.StatusOK, labelSimpleResp{
		ID:      labelID,
		Name:    label.Name,
		Members: req.MemberUIDs,
	})
}

// delete 删除标签
func (l *Label) delete(c *wkhttp.Context) {
	loginUID := c.MustGet("uid").(string)
	id := c.Param("id")
	labelID, _ := strconv.ParseInt(id, 10, 64)
	err := l.db.delete(labelID, loginUID)
	if err != nil {
		l.Error("删除标签失败", zap.Error(err))
		c.ResponseError(errors.New("删除标签失败"))
		return
	}
	c.ResponseOK()
}

// 查询我的所有标签
func (l *Label) list(c *wkhttp.Context) {
	loginUID := c.MustGet("uid").(string)
	labels, err := l.db.query(loginUID)
	if err != nil {
		l.Error("查询用户标签列表失败", zap.Error(err))
		c.ResponseError(errors.New("查询用户标签列表失败"))
		return
	}
	list := make([]*labelResp, 0)
	for _, label := range labels {
		uids := strings.Split(label.MemberUids, ",")
		friends, err := l.userService.GetFriendsWithToUIDs(loginUID, uids)
		if err != nil {
			l.Error("查询标签成员列表失败", zap.Error(err))
			c.ResponseError(errors.New("查询标签成员列表失败"))
			return
		}
		members := make([]*labelMember, 0)
		for _, friend := range friends {
			members = append(members, &labelMember{
				UID:  friend.UID,
				Name: friend.Name,
			})
		}
		list = append(list, &labelResp{
			ID:      label.Id,
			Name:    label.Name,
			Count:   len(uids),
			Members: members,
		})

	}
	c.Response(list)
}

// detail 标签详情
func (l *Label) detail(c *wkhttp.Context) {
	id := c.Param("id")
	loginUID := c.MustGet("uid").(string)
	if strings.TrimSpace(id) == "" {
		c.ResponseError(errors.New("标签ID不能为空"))
		return
	}
	labelID, _ := strconv.ParseInt(id, 10, 64)
	label, err := l.db.queryDetail(labelID)
	if err != nil {
		c.ResponseError(errors.New("查询标签详情失败"))
		return
	}
	uids := strings.Split(label.MemberUids, ",")
	friends, err := l.userService.GetFriendsWithToUIDs(loginUID, uids)
	if err != nil {
		l.Error("查询标签成员列表失败", zap.Error(err))
		c.ResponseError(errors.New("查询标签成员列表失败"))
		return
	}
	members := make([]*labelMember, 0)
	for _, friend := range friends {
		members = append(members, &labelMember{
			UID:  friend.UID,
			Name: friend.Name,
		})
	}
	resp := &labelResp{
		ID:      label.Id,
		Name:    label.Name,
		Count:   len(uids),
		Members: members,
	}
	c.Response(resp)
}

// update 修改标签
func (l *Label) update(c *wkhttp.Context) {
	id := c.Param("id")
	loginUID := c.MustGet("uid").(string)
	var req labelReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}

	if len(strings.TrimSpace(req.Name)) == 0 {
		c.ResponseError(errors.New("标签名字不能为空"))
		return
	}

	if len(req.MemberUIDs) == 0 {
		c.ResponseError(errors.New("标签成员不能为空"))
		return
	}
	var uids strings.Builder
	for i := 0; i < len(req.MemberUIDs); i++ {
		if i != 0 {
			uids.WriteString(",")
		}
		uids.WriteString(req.MemberUIDs[i])
	}
	labelID, _ := strconv.ParseInt(id, 10, 64)
	err := l.db.update(req.Name, uids.String(), loginUID, labelID)
	if err != nil {
		l.Error("修改标签失败", zap.Error(err))
		c.ResponseError(errors.New("修改标签失败"))
		return
	}
	c.ResponseOK()
}

// labelReq 添加标签请求
type labelReq struct {
	Name       string   `json:"name"`        //标签名称
	MemberUIDs []string `json:"member_uids"` //成员用户ID
}

// labelResp 标签返回
type labelResp struct {
	ID      int64          `json:"id"`
	Name    string         `json:"name"`    //标签名称
	Count   int            `json:"count"`   //成员数量
	Members []*labelMember `json:"members"` //成员
}
type labelMember struct {
	UID  string `json:"uid"`
	Name string `json:"name"` //用户名称
}

type labelSimpleResp struct {
	ID      int64    `json:"id"`
	Name    string   `json:"name"`    //标签名称
	Members []string `json:"members"` //成员
}
