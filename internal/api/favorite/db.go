package favorite

import (
	"github.com/WuKongIM/WuKongChatServer/pkg/db"
	"github.com/WuKongIM/WuKongChatServer/pkg/util"
	"github.com/gocraft/dbr/v2"
)

// DB DB
type DB struct {
	session *dbr.Session
}

// NewDB NewDB
func NewDB(session *dbr.Session) *DB {
	return &DB{
		session: session,
	}
}

// InsertFavorite 添加收藏
func (d *DB) InsertFavorite(m *Model) error {
	_, err := d.session.InsertInto("favorite").Columns(util.AttrToUnderscore(m)...).Record(m).Exec()
	return err
}

// DeleteFavorite 删除收藏
func (d *DB) DeleteFavorite(uniqueKey string, uid string) error {
	_, err := d.session.DeleteFrom("favorite").Where("unique_key=? and uid=?", uniqueKey, uid).Exec()
	return err
}

// QueryCollectByUniqueKey 通过唯一标识查询收藏
func (d *DB) QueryCollectByUniqueKey(uniqueKey string, uid string) (*Model, error) {
	var favorite *Model
	_, err := d.session.Select("*").From("favorite").Where("unique_key=? and uid=?", uniqueKey, uid).Load(&favorite)
	return favorite, err
}

// QueryFavoriteList 查询收藏
func (d *DB) QueryFavoriteList(uid string, pageSize, page uint64) ([]*Model, error) {
	var favorites []*Model
	_, err := d.session.Select("*").From("favorite").Where("uid=?", uid).Offset((page-1)*pageSize).Limit(pageSize).OrderDir("updated_at", false).Load(&favorites)
	return favorites, err
}

// Model 收藏对象
type Model struct {
	UID        string //收藏者
	Type       int    // 收藏类型 1. 纯文本 2. 图片
	UniqueKey  string // 唯一key
	AuthorUID  string // 作者uid
	AuthorName string // 作者名称
	Payload    string // 负载内容
	db.BaseModel
}
