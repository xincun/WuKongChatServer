package channel

import (
	"github.com/WuKongIM/WuKongChatServer/internal/config"
	"github.com/WuKongIM/WuKongChatServer/pkg/db"
	"github.com/gocraft/dbr/v2"
)

type channelSettingDB struct {
	session *dbr.Session
	ctx     *config.Context
}

func newChannelSettingDB(ctx *config.Context) *channelSettingDB {
	return &channelSettingDB{
		session: ctx.DB(),
		ctx:     ctx,
	}
}

func (c *channelSettingDB) queryWithChannel(channelID string, channelType uint8) (*channelSettingModel, error) {
	var m *channelSettingModel
	_, err := c.session.Select("*").From("channel_setting").Where("channel_id=? and channel_type=?", channelID, channelType).Load(&m)
	return m, err
}

func (c *channelSettingDB) queryWithChannelIDs(channelIDs []string) ([]*channelSettingModel, error) {
	var models []*channelSettingModel
	_, err := c.session.Select("*").From("channel_setting").Where("channel_id in ?", channelIDs).Load(&models)
	return models, err
}

type channelSettingModel struct {
	ChannelID         string
	ChannelType       uint8
	ParentChannelID   string
	ParentChannelType uint8
	db.BaseModel
}
