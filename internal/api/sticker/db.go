package sticker

import (
	"fmt"

	"github.com/WuKongIM/WuKongChatServer/pkg/db"
	"github.com/WuKongIM/WuKongChatServer/pkg/util"
	"github.com/gocraft/dbr/v2"
)

// DB db
type DB struct {
	session *dbr.Session
}

// newDB New
func newDB(session *dbr.Session) *DB {
	return &DB{
		session: session,
	}
}

// 添加表情
func (d *DB) insertStickerTx(m *model, tx *dbr.Tx) (int64, error) {
	result, err := tx.InsertInto("sticker").Columns(util.AttrToUnderscore(m)...).Record(m).Exec()
	id, err := result.LastInsertId()
	return id, err
}

// 添加表情
func (d *DB) insertSticker(m *model) (int64, error) {
	result, err := d.session.InsertInto("sticker").Columns(util.AttrToUnderscore(m)...).Record(m).Exec()
	id, err := result.LastInsertId()
	return id, err
}

// 修改用户自定义表情顺序
func (d *DB) updateCustomStickerSortNumTx(path string, uid string, sortNum int, tx *dbr.Tx) error {
	_, err := tx.Update("sticker_custom").SetMap(map[string]interface{}{
		"sort_num": sortNum,
	}).Where("uid=? and path=?", uid, path).Exec()
	return err
}

// 修改用户表情分类顺序
func (d *DB) updateCategorySortNumTx(category string, uid string, sortNum int, tx *dbr.Tx) error {
	_, err := tx.Update("sticker_user_category").SetMap(map[string]interface{}{
		"sort_num": sortNum,
	}).Where("uid=? and category=?", uid, category).Exec()
	return err
}

// 通过Category查询表情列表
func (d *DB) queryStickersByCategory(category string) ([]*model, error) {
	var stickers []*model
	_, err := d.session.Select("*").From("sticker").Where("category=?", category).Load(&stickers)
	return stickers, err
}

// 添加用户表情
func (d *DB) insertUserStickerTx(m *customModel, tx *dbr.Tx) (int64, error) {
	result, err := tx.InsertInto("sticker_custom").Columns(util.AttrToUnderscore(m)...).Record(m).Exec()
	id, err := result.LastInsertId()
	return id, err
}

// 添加用户表情
func (d *DB) insertUserSticker(m *customModel) (int64, error) {
	result, err := d.session.InsertInto("sticker_custom").Columns(util.AttrToUnderscore(m)...).Record(m).Exec()
	id, err := result.LastInsertId()
	return id, err
}

// 查询编号最大的表情
func (d *DB) queryUserStickerWithMaxSortNum(uid string) (*customModel, error) {
	var m *customModel
	_, err := d.session.Select("*").From("sticker_custom").Where("uid=?", uid).OrderDir("sort_num", false).Limit(1).Load(&m)
	return m, err
}

// 查询某个自定义表情是否添加
func (d *DB) queryUserCustomStickerWithPath(uid string, path string) (*customModel, error) {
	var m *customModel
	_, err := d.session.Select("*").From("sticker_custom").Where("uid=? and path=?", uid, path).Load(&m)
	return m, err
}

// 查询用户自定义表情
func (d *DB) queryUserCustomSticker(uid string) ([]*customModel, error) {
	var stickers []*customModel
	_, err := d.session.Select("*").From("sticker_custom").Where("uid=?", uid).OrderDir("sort_num", false).Load(&stickers)
	return stickers, err
}

// 通过paths批量删除用户表情
func (d *DB) deleteUserStickerWithPaths(paths []string, uid string) error {
	_, err := d.session.DeleteFrom("sticker_custom").Where("path in ? and uid=?", paths, uid).Exec()
	return err
}

// 查询用户保存的表情分类
func (d *DB) queryUserCategorys(uid string) ([]*categoryDetailModel, error) {
	var stickers []*categoryDetailModel
	_, err := d.session.Select("sticker_user_category.*,IFNULL(sticker_store.cover,'') cover,IFNULL(sticker_store.title,'') title,IFNULL(sticker_store.cover_lim,'') cover_lim,IFNULL(sticker_store.`desc`,'') `desc`").From("sticker_user_category").LeftJoin("sticker_store", "sticker_user_category.category=sticker_store.category").Where("sticker_user_category.uid=?", uid).OrderDir("sticker_user_category.sort_num", false).Load(&stickers)
	return stickers, err
}

// 通过category删除用户表情
func (d *DB) deleteUserStickerWithCategory(category string, uid string) error {
	_, err := d.session.DeleteFrom("sticker_user_category").Where("category=? and uid=?", category, uid).Exec()
	return err
}

// 添加用户表情分类
func (d *DB) insertUserCategory(m *categoryModel) (int64, error) {
	result, err := d.session.InsertInto("sticker_user_category").Columns(util.AttrToUnderscore(m)...).Record(m).Exec()
	id, err := result.LastInsertId()
	return id, err
}

// 查询sortnum最大的用户表情分类
func (d *DB) queryUserCategoryWithMaxSortNum(uid string) (*categoryModel, error) {
	var m *categoryModel
	_, err := d.session.Select("*").From("sticker_user_category").Where("uid=?", uid).OrderDir("sort_num", false).Limit(1).Load(&m)
	return m, err
}

// 通过分类查询某个用户是否添加过该分类表情
func (d *DB) queryUserCategoryWithCategory(uid string, category string) (*categoryModel, error) {
	var m *categoryModel
	_, err := d.session.Select("*").From("sticker_user_category").Where("uid=? and category=?", uid, category).Load(&m)
	return m, err
}

// 查询表情商店数据
func (d *DB) queryStickerStroeWithPage(pageIndex, pageSize uint64, uid string) ([]*stickerStoreModel, error) {
	var stickers []*stickerStoreModel
	str := fmt.Sprintf("sticker_store.category=sticker_user_category.category and sticker_user_category.uid='%s'", uid)
	_, err := d.session.Select("sticker_store.*,IF(sticker_user_category.id>0,1,0) status").From("sticker_store").Where("sticker_store.is_gone=0").LeftJoin("sticker_user_category", str).OrderDir("sticker_store.created_at", false).Offset((pageIndex - 1) * pageSize).Limit(pageSize).Load(&stickers)
	return stickers, err
}

// 添加表情商店
func (d *DB) insertStore(m *storeModel) (int64, error) {
	result, err := d.session.InsertInto("sticker_store").Columns(util.AttrToUnderscore(m)...).Record(m).Exec()
	id, err := result.LastInsertId()
	return id, err
}

// 查询某个表情商店
func (d *DB) queryStoreWithCategory(category string) (*storeModel, error) {
	var m *storeModel
	_, err := d.session.Select("*").From("sticker_store").Where("category=?", category).Load(&m)
	return m, err
}

// 搜索表情
func (d *DB) search(keyword string, page, pageSize uint64) ([]*model, error) {
	var list []*model
	_, err := d.session.Select("*").From("`sticker`").Where("searchable_word like ? and category <> 'emoji'", "%"+keyword+"%").Offset((page-1)*pageSize).Limit(pageSize).OrderDir("created_at", false).Load(&list)
	return list, err
}

// 判断某个用户是否添加过某个分类的表情
func (d *DB) isAddedCategory(category, uid string) (bool, error) {
	var m *categoryModel
	_, err := d.session.Select("*").From("sticker_user_category").Where("uid=? and category=?", uid, category).Load(&m)
	return m != nil, err
}

// 表情model
type model struct {
	Title          string //标题
	Path           string //表情地址
	Category       string // 分类
	SearchableWord string //搜索关键字
	UserCustom     int    //1：用户自定义
	Placeholder    string //占位图
	Format         string // 表情格式 gif｜gzip
	Width          int
	Height         int
	db.BaseModel
}

// 用户自定义表情model
type customModel struct {
	Path        string // 表情地址
	UID         string // uid
	SortNum     int    //排序编号
	Format      string // 表情格式
	Placeholder string // 占位图
	Category    string // 表情分类
	Width       int
	Height      int
	db.BaseModel
}

// 表情分类model
type categoryModel struct {
	UID      string // uid
	SortNum  int    //排序编号
	Category string // 分类
	db.BaseModel
}

// 分类详情
type categoryDetailModel struct {
	categoryModel
	Cover    string //封面
	CoverLim string
	Title    string //标题
	Desc     string //说明
}

// 表情商店
type stickerStoreModel struct {
	Category string //分类
	Title    string //标题
	Desc     string //说明
	Cover    string //封面
	CoverLim string
	Status   int //1:已经添加0:未添加
	db.BaseModel
}

// 表情商店
type storeModel struct {
	Desc     string //说明
	Cover    string //封面
	Title    string //标题
	CoverLim string
	Category string //分类
}
