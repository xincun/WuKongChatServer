package sticker

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/WuKongIM/WuKongChatServer/internal/api/base/event"
	"github.com/WuKongIM/WuKongChatServer/internal/config"
	"github.com/WuKongIM/WuKongChatServer/pkg/log"
	"github.com/WuKongIM/WuKongChatServer/pkg/wkhttp"
	"go.uber.org/zap"
)

// Sticker Sticker
type Sticker struct {
	ctx *config.Context
	log.Log
	db *DB
}

// New 创建
func New(ctx *config.Context) *Sticker {
	s := &Sticker{ctx: ctx, db: newDB(ctx.DB()), Log: log.NewTLog("Sticker")}

	return s
}

// Route 路由配置
func (s *Sticker) Route(r *wkhttp.WKHttp) {
	v := r.Group("/v1/sticker", r.AuthMiddleware(s.ctx.Cache(), s.ctx.GetConfig().TokenCachePrefix))
	{
		v.GET("", s.search)
		//用户添加表情
		v.POST("/user", s.userAdd)
		//用户删除表情
		v.DELETE("/user", s.userDelete)
		//用户自定义表情
		v.GET("/user", s.userCustomSticker)
		//用户移除表情分类
		v.DELETE("/remove", s.userDeleteByCategory)
		//通过category添加一批表情
		v.POST("/user/:category", s.userAddByCategory)
		//获取用户分类列表
		v.GET("/user/category", s.getCategorys)
		//通过分类获取表情
		v.GET("/user/sticker", s.getStickerWithCategory)
		//获取表情商店
		v.GET("/store", s.list)
		//将自定义表情移到最前
		v.PUT("/user/front", s.moveToFront)
		//将用户表情分类排序
		v.PUT("/user/category/reorder", s.reorderUserCategory)
	}

	if !s.ctx.GetConfig().StickerAddOffOfRegister {
		s.ctx.AddEventListener(event.EventUserRegister, s.handleRegisterUserEvent)
	}

}

// 用户添加表情
func (s *Sticker) userAdd(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	var req struct {
		Path        string `json:"path"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
		Format      string `json:"format"`
		Placeholder string `json:"placeholder"`
		Category    string `json:"category"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}
	if req.Path == "" {
		c.ResponseError(errors.New("文件地址不能为空"))
		return
	}
	if req.Width == 0 || req.Height == 0 {
		c.ResponseError(errors.New("表情高宽不能为空"))
		return
	}
	tempSticker, err := s.db.queryUserCustomStickerWithPath(loginUID, req.Path)
	if err != nil {
		s.Error("查询用户自定义表情错误", zap.Error(err))
		c.ResponseError(errors.New("查询用户自定义表情错误"))
		return
	}
	if tempSticker != nil && tempSticker.Path != "" {
		c.ResponseOK()
		return
	}
	cmodel, err := s.db.queryUserStickerWithMaxSortNum(loginUID)
	if err != nil {
		s.Error("删除表情错误", zap.Error(err))
		c.ResponseError(errors.New("查询用户表情最大编号错误"))
		return
	}
	var sortNum int = 1
	if cmodel != nil {
		sortNum = cmodel.SortNum + 1
	}
	tx, _ := s.ctx.DB().Begin()
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			panic(err)
		}
	}()
	//将表情添加到用户表
	_, err = s.db.insertUserStickerTx(&customModel{
		Path:        req.Path,
		UID:         loginUID,
		SortNum:     sortNum,
		Width:       req.Width,
		Height:      req.Height,
		Format:      req.Format,
		Placeholder: req.Placeholder,
		Category:    req.Category,
	}, tx)
	if err != nil {
		tx.Rollback()
		s.Error("添加用户表情错误", zap.Error(err))
		c.ResponseError(errors.New("添加用户表情错误"))
		return
	}
	//将表情添加到所有表情中
	_, err = s.db.insertStickerTx(&model{
		Path:           req.Path,
		Category:       loginUID,
		UserCustom:     1,
		Width:          req.Width,
		Height:         req.Height,
		SearchableWord: "",
		Format:         req.Format,
		Title:          "",
		Placeholder:    req.Placeholder,
	}, tx)
	if err != nil {
		tx.Rollback()
		s.Error("将用户自定义表情添加到表情中错误", zap.Error(err))
		c.ResponseError(errors.New("将用户自定义表情添加到表情中错误"))
		return
	}

	err = tx.Commit()
	if err != nil {
		s.Error("数据库事物提交失败", zap.Error(err))
		c.ResponseError(errors.New("数据库事物提交失败"))
		tx.Rollback()
		return
	}
	c.ResponseOK()
}

// 用户删除表情
func (s *Sticker) userDelete(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	var req struct {
		Paths []string `json:"paths"` //删除表情ID集合
	}
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}
	if len(req.Paths) == 0 {
		c.ResponseError(errors.New("数据不能为空"))
		return
	}
	err := s.db.deleteUserStickerWithPaths(req.Paths, loginUID)
	if err != nil {
		s.Error("删除表情错误", zap.Error(err))
		c.ResponseError(errors.New("删除表情错误"))
		return
	}
	c.ResponseOK()
}

// 获取用户自定义表情
func (s *Sticker) userCustomSticker(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	list, err := s.db.queryUserCustomSticker(loginUID)
	if err != nil {
		s.Error("查询用户自定义表情错误", zap.Error(err))
		c.ResponseError(errors.New("查询用户自定义表情错误"))
		return
	}
	resps := make([]*categoryStickerResp, 0)
	if list != nil && len(list) > 0 {
		for _, model := range list {
			resps = append(resps, &categoryStickerResp{
				Path:        model.Path,
				Category:    model.Category,
				Width:       model.Width,
				Height:      model.Height,
				SortNum:     model.SortNum,
				Format:      model.Format,
				Placeholder: model.Placeholder,
			})
		}
	}
	c.Response(resps)
}

// 通过category删除用户表情
func (s *Sticker) userDeleteByCategory(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	category := c.Query("category")
	if category == "" {
		c.ResponseError(errors.New("分类不能为空"))
		return
	}
	err := s.db.deleteUserStickerWithCategory(category, loginUID)
	if err != nil {
		s.Error("移除表情分类错误", zap.Error(err))
		c.ResponseError(errors.New("移除表情分类错误"))
		return
	}
	c.ResponseOK()
}

// 通过分类添加表情
func (s *Sticker) userAddByCategory(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	category := c.Param("category")
	if category == "" {
		c.ResponseError(errors.New("分类名不能为空"))
		return
	}
	list, err := s.db.queryStickersByCategory(category)
	if err != nil {
		s.Error("通过category查询表情错误", zap.Error(err))
		c.ResponseError(errors.New("通过category查询表情错误"))
		return
	}
	if len(list) == 0 {
		c.ResponseError(errors.New("该分类下表情为空"))
		return
	}
	model, err := s.db.queryUserCategoryWithCategory(loginUID, category)
	if err != nil {
		s.Error("查询用户分类表情错误", zap.Error(err))
		c.ResponseError(errors.New("查询用户分类表情错误"))
		return
	}
	if model != nil {
		c.ResponseOK()
		return
	}
	cmodel, err := s.db.queryUserCategoryWithMaxSortNum(loginUID)
	if err != nil {
		s.Error("查询最大用户表情分类错误", zap.Error(err))
		c.ResponseError(errors.New("查询最大用户表情分类错误"))
		return
	}
	var sortNum int = 1
	if cmodel != nil {
		sortNum = cmodel.SortNum + 1
	}
	_, err = s.db.insertUserCategory(&categoryModel{
		UID:      loginUID,
		Category: category,
		SortNum:  sortNum,
	})
	if err != nil {
		s.Error("添加表情分类错误", zap.Error(err))
		c.ResponseError(errors.New("添加表情分类错误"))
		return
	}
	c.ResponseOK()
}

// 获取用户表情分类
func (s *Sticker) getCategorys(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	list, err := s.db.queryUserCategorys(loginUID)
	if err != nil {
		s.Error("查询用户表情分类错误", zap.Error(err))
		c.ResponseError(errors.New("查询用户表情分类错误"))
		return
	}
	result := make([]*stickerCategoryResp, 0)
	if list != nil && len(list) > 0 {
		for _, model := range list {
			if model.Category != loginUID {
				result = append(result, &stickerCategoryResp{
					Category: model.Category,
					Cover:    model.Cover,
					CoverLim: model.CoverLim,
					SortNum:  model.SortNum,
					Title:    model.Title,
					Desc:     model.Desc,
				})
			}
		}
	}
	c.Response(result)
}

// 获取分类下的表情
func (s *Sticker) getStickerWithCategory(c *wkhttp.Context) {
	category := c.Query("category")
	if category == "" {
		c.ResponseError(errors.New("参数错误"))
		return
	}
	// 获取表情详情
	m, err := s.db.queryStoreWithCategory(category)
	if err != nil {
		s.Error("查询分类表情错误", zap.Error(err))
		c.ResponseError(errors.New("查询分类表情错误"))
		return
	}
	if m == nil {
		c.ResponseError(errors.New("该分类无表情"))
		return
	}
	isAdded, err := s.db.isAddedCategory(category, c.GetLoginUID())
	if err != nil {
		s.Error("查询登录用户是否添加该分类表情错误", zap.Error(err))
		c.ResponseError(errors.New("查询登录用户是否添加该分类表情错误"))
		return
	}
	list, err := s.db.queryStickersByCategory(category)
	if err != nil {
		s.Error("查询分类下的表情错误", zap.Error(err))
		c.ResponseError(errors.New("查询分类下的表情错误"))
		return
	}
	stickerList := make([]*categoryStickerResp, 0)
	if list != nil && len(list) > 0 {
		for _, model := range list {
			stickerList = append(stickerList, &categoryStickerResp{
				Path:           model.Path,
				Category:       model.Category,
				Title:          model.Title,
				Width:          model.Width,
				Height:         model.Height,
				Placeholder:    model.Placeholder,
				Format:         model.Format,
				SearchableWord: model.SearchableWord,
			})
		}
	}
	var resp = &stickerDetialResp{
		List:     stickerList,
		Title:    m.Title,
		Cover:    m.Cover,
		CoverLim: m.CoverLim,
		Category: category,
		Desc:     m.Desc,
		Added:    isAdded,
	}
	c.Response(resp)
}

// 表情商店
func (s *Sticker) list(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	pageIndex, pageSize := c.GetPage()
	list, err := s.db.queryStickerStroeWithPage(uint64(pageIndex), uint64(pageSize), loginUID)
	if err != nil {
		s.Error("查询表情商店错误", zap.Error(err))
		c.ResponseError(errors.New("查询表情商店错误"))
		return
	}
	resps := make([]*stickerStoreResp, 0)
	if list != nil && len(list) > 0 {
		for _, model := range list {
			resps = append(resps, &stickerStoreResp{
				Status:   model.Status,
				Title:    model.Title,
				Desc:     model.Desc,
				Cover:    model.Cover,
				CoverLim: model.CoverLim,
				Category: model.Category,
			})
		}
	}
	c.Response(resps)
}

// 将某些表情移到最前
func (s *Sticker) moveToFront(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	var req struct {
		Paths []string `json:"paths"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}
	if len(req.Paths) == 0 {
		c.ResponseError(errors.New("数据不能为空"))
		return
	}
	cmodel, err := s.db.queryUserStickerWithMaxSortNum(loginUID)
	if err != nil {
		s.Error("查询用户表情最大编号错误", zap.Error(err))
		c.ResponseError(errors.New("查询用户表情最大编号错误"))
		return
	}
	var maxNum = 0
	if cmodel != nil {
		maxNum = cmodel.SortNum
	}
	tx, _ := s.ctx.DB().Begin()
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			panic(err)
		}
	}()
	var tempSortNum = len(req.Paths)
	for _, path := range req.Paths {
		err = s.db.updateCustomStickerSortNumTx(path, loginUID, (tempSortNum + maxNum), tx)
		if err != nil {
			tx.Rollback()
			s.Error("修改用户自定义表情顺序错误", zap.Error(err))
			c.ResponseError(errors.New("修改用户自定义表情顺序错误"))
			return
		}
		tempSortNum--
	}
	err = tx.Commit()
	if err != nil {
		s.Error("数据库事物提交失败", zap.Error(err))
		c.ResponseError(errors.New("数据库事物提交失败"))
		tx.Rollback()
		return
	}
	c.ResponseOK()
}

// 将用户表情分类排序
func (s *Sticker) reorderUserCategory(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	var req struct {
		Categorys []string `json:"categorys"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}
	if len(req.Categorys) == 0 {
		c.ResponseError(errors.New("数据不能为空"))
		return
	}
	tx, _ := s.ctx.DB().Begin()
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			panic(err)
		}
	}()
	var sortNum = 0
	for _, category := range req.Categorys {
		err := s.db.updateCategorySortNumTx(category, loginUID, (len(req.Categorys) - sortNum), tx)
		if err != nil {
			tx.Rollback()
			s.Error("修改用户表情分类顺序错误", zap.Error(err))
			c.ResponseError(errors.New("修改用户表情分类顺序错误"))
			return
		}
		sortNum++
	}
	err := tx.Commit()
	if err != nil {
		s.Error("数据库事物提交失败", zap.Error(err))
		c.ResponseError(errors.New("数据库事物提交失败"))
		tx.Rollback()
		return
	}
	c.ResponseOK()
}

// 搜索表情
func (s *Sticker) search(c *wkhttp.Context) {
	keyword := c.Query("keyword")
	page := c.Query("page")
	pageSize := c.Query("page_size")
	if page == "" {
		page = "1"
	}
	if pageSize == "" {
		pageSize = "20"
	}
	pageI, _ := strconv.ParseInt(page, 10, 64)
	pageSizeI, _ := strconv.ParseInt(pageSize, 10, 64)
	list, err := s.db.search(keyword, uint64(pageI), uint64(pageSizeI))
	if err != nil {
		s.Error("查询表情失败！", zap.Error(err))
		c.ResponseError(errors.New("查询表情失败！"))
		return
	}
	resps := make([]*stickerResp, 0)
	if list == nil || len(list) == 0 {
		c.JSON(http.StatusOK, resps)
		return
	}
	for _, m := range list {
		resps = append(resps, &stickerResp{
			Title:          m.Title,
			Category:       m.Category,
			Height:         m.Height,
			Width:          m.Width,
			Format:         m.Format,
			Path:           m.Path,
			Placeholder:    m.Placeholder,
			SearchableWord: m.SearchableWord,
		})
	}
	// resp, err := network.Get("https://www.soogif.com/material/query", map[string]string{
	// 	"query": keyword,
	// 	"start": fmt.Sprintf("%d", (pageI-1)*pageSizeI),
	// 	"size":  pageSize,
	// }, nil)
	// if err != nil {
	// 	s.Error("查询表情失败！", zap.Error(err))
	// 	c.ResponseError(errors.New("查询表情失败！"))
	// 	return
	// }
	// var soogifResp soogifResp
	// err = util.ReadJsonByByte([]byte(resp.Body), &soogifResp)
	// if err != nil {
	// 	s.Error("服务器发送错误！", zap.Error(err))
	// 	c.ResponseError(errors.New("服务器发送错误！"))
	// 	return
	// }
	// stickerResps := make([]stickerResp, 0)
	// if len(soogifResp.Data.List) > 0 {
	// 	for _, soogifItem := range soogifResp.Data.List {
	// 		stickerResp := stickerResp{}
	// 		stickerResp.Title = soogifItem.Title
	// 		stickerResp.Size = soogifItem.Size
	// 		stickerResp.Width = soogifItem.Width
	// 		stickerResp.Height = soogifItem.Height
	// 		stickerResp.URL = soogifItem.URL
	// 		stickerResp.StickerType = "gif"
	// 		stickerResps = append(stickerResps, stickerResp)
	// 	}
	// }
	c.JSON(http.StatusOK, resps)
}

// stickerResp stickerResp
type stickerResp struct {
	Path           string `json:"path"` // 表情地址
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	Title          string `json:"title"`           // 表情名字
	Category       string `json:"category"`        // 分类
	Placeholder    string `json:"placeholder"`     //占位图
	Format         string `json:"format"`          // 表情类型 gif|gzip
	SearchableWord string `json:"searchable_word"` // 搜索关键字
}

type soogifResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Total int `json:"total"`
		List  []struct {
			Flag     int    `json:"flag"`
			Size     int64  `json:"size"`
			Width    int    `json:"width"`
			Text     string `json:"text"`
			Title    string `json:"title"`
			Category int    `json:"category"`
			SubText  string `json:"subText"`
			URL      string `json:"url"`
			Tags     string `json:"tags"`
			Sid      string `json:"sid"`
			Height   int    `json:"height"`
		} `json:"list"`
		UniqueSubTextsJSON []struct {
			Size          int    `json:"size"`
			UniqueSubText string `json:"uniqueSubText"`
			URL           string `json:"url"`
		} `json:"uniqueSubTextsJson"`
		ResuLtTotal int `json:"resuLtTotal"`
	} `json:"data"`
}

// 表情分类
type stickerCategoryResp struct {
	Category string `json:"category"` // 分类
	Cover    string `json:"cover"`    // 封面
	CoverLim string `json:"cover_lim"`
	SortNum  int    `json:"sort_num"` //排序号
	Title    string `json:"title"`    // 标题
	Desc     string `json:"desc"`     // 描述
}

// 表情商店
type stickerStoreResp struct {
	Status   int    `json:"status"`    // 1:用户已添加
	Category string `json:"category"`  // 分类
	Cover    string `json:"cover"`     // 封面
	CoverLim string `json:"cover_lim"` // json格式封面
	Title    string `json:"title"`     // 标题
	Desc     string `json:"desc"`      // 封面
}

type categoryStickerResp struct {
	Path           string `json:"path"` // 表情地址
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	Title          string `json:"title"`           // 表情名字
	Category       string `json:"category"`        // 分类
	Placeholder    string `json:"placeholder"`     //占位图
	Format         string `json:"format"`          // 表情类型 gif|gzip
	SortNum        int    `json:"sort_num"`        //排序编号
	SearchableWord string `json:"searchable_word"` // 搜索关键字
}

// 表情详情
type stickerDetialResp struct {
	List     []*categoryStickerResp `json:"list"`  // 表情列表
	Desc     string                 `json:"desc"`  //说明
	Cover    string                 `json:"cover"` //封面
	CoverLim string                 `json:"cover_lim"`
	Title    string                 `json:"title"`    //标题
	Category string                 `json:"category"` //分类
	Added    bool                   `json:"added"`    // 是否添加
}
