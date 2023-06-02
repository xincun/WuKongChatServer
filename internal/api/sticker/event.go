package sticker

import (
	"errors"

	"github.com/WuKongIM/WuKongChatServer/internal/config"
	"github.com/WuKongIM/WuKongChatServer/pkg/util"
	"go.uber.org/zap"
)

// 新增注册用户时默认添加一个表情
func (s *Sticker) handleRegisterUserEvent(data []byte, commit config.EventCommit) {
	var req map[string]interface{}
	err := util.ReadJsonByByte(data, &req)
	if err != nil {
		s.Error("表情处理用户注册加入群聊参数有误")
		commit(err)
		return
	}
	uid := req["uid"].(string)
	if uid == "" {
		s.Error("表情处理用户注册加入群聊UID不能为空")
		commit(errors.New("表情处理用户注册加入群聊UID不能为空"))
		return
	}
	model, err := s.db.queryUserCategoryWithCategory(uid, "duck")
	if err != nil {
		s.Error("查询用户分类表情错误", zap.Error(err))
		commit(errors.New("查询用户分类表情错误"))
		return
	}
	if model != nil {
		commit(nil)
		return
	}
	cmodel, err := s.db.queryUserCategoryWithMaxSortNum(uid)
	if err != nil {
		s.Error("查询最大用户表情分类错误", zap.Error(err))
		commit(errors.New("查询最大用户表情分类错误"))
		return
	}
	var sortNum int = 1
	if cmodel != nil {
		sortNum = cmodel.SortNum + 1
	}
	_, err = s.db.insertUserCategory(&categoryModel{
		UID:      uid,
		SortNum:  sortNum,
		Category: "duck",
	})
	if err != nil {
		s.Error("注册用户添加表情分类错误", zap.Error(err))
		commit(err)
		return
	}
	commit(nil)
}
