package favorite

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/WuKongIM/WuKongChatServer/internal/config"
	"github.com/WuKongIM/WuKongChatServer/pkg/log"
	"github.com/WuKongIM/WuKongChatServer/pkg/wkhttp"
	"go.uber.org/zap"
)

// Favorite 收藏
type Favorite struct {
	ctx *config.Context
	log.Log
	db *DB
}

// New New
func New(ctx *config.Context) *Favorite {
	return &Favorite{ctx: ctx, db: NewDB(ctx.DB()), Log: log.NewTLog("Favorite")}
}

// Route 路由配置
func (f *Favorite) Route(r *wkhttp.WKHttp) {
	favorites := r.Group("/v1/favorites", r.AuthMiddleware(f.ctx.Cache(), f.ctx.GetConfig().TokenCachePrefix))
	{
		favorites.POST("", f.add)          // 添加收藏
		favorites.DELETE("/:id", f.delete) //删除收藏
	}
	favorite := r.Group("/v1/favorite", r.AuthMiddleware(f.ctx.Cache(), f.ctx.GetConfig().TokenCachePrefix))
	{
		favorite.GET("/my", f.page)
	}
}

// 添加收藏
func (f *Favorite) add(c *wkhttp.Context) {
	loginUID := c.MustGet("uid").(string)
	var req favoriteReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}
	if err := req.check(); err != nil {
		c.ResponseError(err)
		return
	}

	favorite, err := f.db.QueryCollectByUniqueKey(req.UniqueKey, loginUID)
	if err != nil {
		f.Error("通过unique_key查询收藏失败", zap.Error(err))
		c.ResponseError(errors.New("通过unique_key查询收藏失败"))
		return
	}
	if favorite != nil {
		c.ResponseOK()
		return
	}

	favoriteModel := &Model{}
	favoriteModel.UID = loginUID
	favoriteModel.AuthorName = req.AuthorName
	favoriteModel.AuthorUID = req.AuthorUID
	favoriteModel.Type = req.Type
	favoriteModel.UniqueKey = req.UniqueKey
	//	favoriteModel.Payload = json
	b, err := json.Marshal(req.Payload)
	favoriteModel.Payload = string(b[:])
	err = f.db.InsertFavorite(favoriteModel)
	if err != nil {
		f.Error("添加收藏失败", zap.Error(err))
		c.ResponseError(errors.New("添加收藏失败"))
		return
	}
	c.ResponseOK()
}

// 删除收藏
func (f *Favorite) delete(c *wkhttp.Context) {
	id := c.Param("id")
	loginUID := c.GetLoginUID()
	err := f.db.DeleteFavorite(id, loginUID)
	if err != nil {
		f.Error("删除收藏失败", zap.Error(err))
		c.ResponseError(errors.New("删除收藏失败"))
		return
	}
	c.ResponseOK()
}

// 获取我的收藏列表
func (f *Favorite) page(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	page := c.Query("page_index")
	size := c.Query("page_size")
	pageIndex, _ := strconv.Atoi(page)
	pageSize, _ := strconv.Atoi(size)
	list, err := f.db.QueryFavoriteList(loginUID, uint64(pageSize), uint64(pageIndex))
	if err != nil {
		f.Error("查询收藏列表失败", zap.Error(err))
		c.ResponseError(errors.New("查询收藏列表失败"))
		return
	}
	favoriteResps := []*favoriteRes{}
	for _, result := range list {
		var dat map[string]interface{}
		if err = json.Unmarshal([]byte(result.Payload), &dat); err != nil {
			f.Error("收藏内容转换失败", zap.Error(err))
			continue
		}
		favoriteResps = append(favoriteResps, &favoriteRes{
			UniqueKey:  result.UniqueKey,
			Type:       result.Type,
			AuthorName: result.AuthorName,
			AuthorUID:  result.AuthorUID,
			No:         fmt.Sprintf("%d", result.Id),
			CreatedAt:  result.CreatedAt.String(),
			Payload:    dat,
		})
	}
	c.Response(favoriteResps)
}

func (f *favoriteReq) check() error {
	if strings.TrimSpace(f.AuthorUID) == "" {
		return errors.New("作者uid不能为空！")
	}
	if strings.TrimSpace(f.AuthorName) == "" {
		return errors.New("作者名称不能为空！")
	}
	if strings.TrimSpace(f.UniqueKey) == "" {
		return errors.New("收藏唯一key不能为空！")
	}

	if f.Payload == nil || len(f.Payload) == 0 {
		return errors.New("收藏内容不能为空！")
	}
	return nil
}

// ------------ vo -----------
type favoriteReq struct {
	Type       int                    `json:"type"`        // 收藏类型 1. 纯文本 2. 图片
	UniqueKey  string                 `json:"unique_key"`  //唯一key 如果是收藏消息，此字段一般为message_id
	AuthorName string                 `json:"author_name"` //作者名称
	AuthorUID  string                 `json:"author_uid"`  //作者uid份
	Payload    map[string]interface{} `json:"payload"`     //收藏内容
}

type favoriteRes struct {
	No         string                 `json:"no"`
	Type       int                    `json:"type"`        // 收藏类型 1. 纯文本 2. 图片
	UniqueKey  string                 `json:"unique_key"`  //唯一key 如果是收藏消息，此字段一般为message_id
	AuthorName string                 `json:"author_name"` //作者名称
	AuthorUID  string                 `json:"author_uid"`  //作者uid份
	Payload    map[string]interface{} `json:"payload"`     //收藏内容
	CreatedAt  string                 `json:"created_at"`  // 创建时间
}
