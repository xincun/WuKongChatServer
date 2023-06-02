package label

import (
	"github.com/WuKongIM/WuKongChatServer/pkg/db"
	"github.com/WuKongIM/WuKongChatServer/pkg/util"
	"github.com/gocraft/dbr/v2"
)

// DB DB
type DB struct {
	session *dbr.Session
}

// newDB New
func newDB(session *dbr.Session) *DB {
	return &DB{
		session: session,
	}
}

// insert 添加标签
func (d *DB) insert(m *model) (int64, error) {
	result, err := d.session.InsertInto("label").Columns(util.AttrToUnderscore(m)...).Record(m).Exec()
	id, err := result.LastInsertId()
	return id, err
}

// update 修改标签
func (d *DB) update(name, memberUIDs, uid string, id int64) error {
	_, err := d.session.Update("label").SetMap(map[string]interface{}{
		"name":        name,
		"member_uids": memberUIDs,
	}).Where("uid=? and id=?", uid, id).Exec()
	return err
}

// delete 删除
func (d *DB) delete(id int64, uid string) error {
	_, err := d.session.DeleteFrom("label").Where("id=? and uid=?", id, uid).Exec()
	return err
}

// query 标签列表
func (d *DB) query(uid string) ([]*model, error) {
	var labels []*model
	_, err := d.session.Select("*").From("label").Where("uid=?", uid).OrderDir("updated_at", false).Load(&labels)
	return labels, err
}

// queryDetail 查询标签详情
func (d *DB) queryDetail(id int64) (*model, error) {
	var label *model
	_, err := d.session.Select("*").From("label").Where("id=?", id).Load(&label)
	return label, err
}

// model 标签对象
type model struct {
	UID        string //标签所属者
	Name       string //标签名字
	MemberUids string //标签成员
	db.BaseModel
}
