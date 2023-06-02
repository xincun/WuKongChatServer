package user

import (
	"errors"

	"github.com/WuKongIM/WuKongChatServer/internal/common"
	"github.com/WuKongIM/WuKongChatServer/pkg/wkhttp"
	"go.uber.org/zap"
)

func (u *User) signalKeyGet(c *wkhttp.Context) {
	var req signalKeyGetReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("数据格式有误！"))
		u.Error("数据格式有误！", zap.Error(err))
		return
	}
	if err := req.check(); err != nil {
		c.ResponseError(err)
		return
	}
	if req.ChannelType != common.ChannelTypePerson.Uint8() {
		c.ResponseError(errors.New("暂只支持个人聊天加密！"))
		return
	}
	identityM, err := u.identitieDB.queryWithUID(req.ChannelID)
	if err != nil {
		c.ResponseErrorf("查询用户身份失败！", err)
		return
	}
	if identityM == nil {
		c.ResponseError(errors.New("用户未上传身份信息！"))
		return
	}
	onetimePrekeyM, err := u.onetimePrekeysDB.queryMinWithUID(req.ChannelID)
	if err != nil {
		c.ResponseError(errors.New("查询一次性key失败！"))
		u.Error("查询一次性key失败！", zap.Error(err))
		return
	}
	if onetimePrekeyM != nil {
		err = u.onetimePrekeysDB.delete(onetimePrekeyM.UID, onetimePrekeyM.KeyID)
		if err != nil {
			c.ResponseError(errors.New("删除一次性密钥失败！"))
			u.Error("删除一次性密钥失败！", zap.Error(err))
			return
		}
	}

	c.Response(toSignalKeyGetResp(identityM, onetimePrekeyM))
}

// 添加用户signal相关key
func (u *User) signalKeyAdd(c *wkhttp.Context) {
	var req signalKeyReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("数据格式有误！"))
		u.Error("数据格式有误！", zap.Error(err))
		return
	}
	if err := req.check(); err != nil {
		c.ResponseError(err)
		return
	}
	loginUID := c.GetLoginUID()

	err := u.identitieDB.deleteWithUID(loginUID)
	if err != nil {
		c.ResponseErrorf("删除历史身份key失败！", err)
		return
	}
	err = u.onetimePrekeysDB.deleteWithUID(loginUID)
	if err != nil {
		c.ResponseErrorf("删除历史一次性key失败！", err)
		return
	}

	tx, _ := u.db.session.Begin()
	defer func() {
		if err := recover(); err != nil {
			tx.RollbackUnlessCommitted()
			panic(err)
		}
	}()
	identitiesM := req.toIdentitiesModel(loginUID)

	err = u.identitieDB.saveOrUpdateTx(identitiesM, tx)
	if err != nil {
		tx.Rollback()
		u.Error("保存或更新singal key信息失败！", zap.Error(err))
		c.ResponseError(errors.New("保存或更新singal key信息失败！"))
		return
	}
	onetimePrekeyMs := req.toOneTimePreKeyModels(loginUID)
	if len(onetimePrekeyMs) > 0 {
		for _, onetimePrekeyM := range onetimePrekeyMs {
			err = u.onetimePrekeysDB.insertTx(onetimePrekeyM, tx)
			if err != nil {
				tx.Rollback()
				u.Error("添加一次性密钥失败！", zap.Error(err))
				c.ResponseError(errors.New("添加一次性密钥失败！"))
				return
			}
		}
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		u.Error("提交事务失败！", zap.Error(err))
		c.ResponseError(errors.New("提交事务失败！"))
		return
	}
	c.ResponseOK()
}

type signalKeyReq struct {
	RegistrationID  uint32          `json:"registration_id"`  // 身份ID
	IdentityKey     string          `json:"identity_key"`     // 身份公钥(长期公钥)
	SignedPrekeyID  int             `json:"signed_prekey_id"` //
	SignedPubkey    string          `json:"signed_pubkey"`    // 签名公钥 (中期公钥)
	SignedSignature string          `json:"signed_signature"` // 签名后的中期公钥
	OnetimePreKeys  []oneTimePreKey `json:"onetime_prekeys"`
}

func (s signalKeyReq) toIdentitiesModel(uid string) *identitiesModel {
	return &identitiesModel{
		UID:             uid,
		RegistrationID:  s.RegistrationID,
		IdentityKey:     s.IdentityKey,
		SignedPrekeyID:  s.SignedPrekeyID,
		SignedPubkey:    s.SignedPubkey,
		SignedSignature: s.SignedSignature,
	}
}

func (s signalKeyReq) toOneTimePreKeyModels(uid string) []*onetimePrekeysModel {
	models := make([]*onetimePrekeysModel, 0, len(s.OnetimePreKeys))
	if len(s.OnetimePreKeys) > 0 {
		for _, onetimePreKey := range s.OnetimePreKeys {
			models = append(models, onetimePreKey.toOnetimePrekeysModel(uid))
		}
	}
	return models
}

func (s signalKeyReq) check() error {
	if s.RegistrationID == 0 {
		return errors.New("registration_id不能为空！")
	}
	if s.IdentityKey == "" {
		return errors.New("identity_key不能为空！")
	}
	if s.SignedPrekeyID <= 0 {
		return errors.New("signed_prekey_id不能小于或等于0")
	}
	if s.SignedPubkey == "" {
		return errors.New("signed_pubkey不能为空！")
	}
	if s.SignedSignature == "" {
		return errors.New("signed_signature不能为空！")
	}
	return nil
}

// 一次性公钥
type oneTimePreKey struct {
	KeyID  int    `json:"key_id"`
	Pubkey string `json:"pubkey"`
}

func (o oneTimePreKey) toOnetimePrekeysModel(uid string) *onetimePrekeysModel {
	return &onetimePrekeysModel{
		UID:    uid,
		KeyID:  o.KeyID,
		Pubkey: o.Pubkey,
	}
}

// 获取signal key的请求
type signalKeyGetReq struct {
	ChannelID   string `json:"channel_id"`
	ChannelType uint8  `json:"channel_type"`
}

func (s signalKeyGetReq) check() error {
	if s.ChannelID == "" {
		return errors.New("频道ID不能为空！")
	}
	return nil
}

type signalKeyGetResp struct {
	UID             string        `json:"uid"`              // 用户uid
	RegistrationID  uint32        `json:"registration_id"`  // 身份ID
	IdentityKey     string        `json:"identity_key"`     // 身份公钥(长期公钥)
	SignedPrekeyID  int           `json:"signed_prekey_id"` //
	SignedPubkey    string        `json:"signed_pubkey"`    // 签名公钥 (中期公钥)
	SignedSignature string        `json:"signed_signature"` // 签名后的中期公钥
	OnetimePreKey   oneTimePreKey `json:"onetime_prekey"`
}

func toSignalKeyGetResp(identitiesM *identitiesModel, onetimePrekeysM *onetimePrekeysModel) *signalKeyGetResp {
	s := &signalKeyGetResp{
		UID:             identitiesM.UID,
		RegistrationID:  identitiesM.RegistrationID,
		IdentityKey:     identitiesM.IdentityKey,
		SignedPrekeyID:  identitiesM.SignedPrekeyID,
		SignedPubkey:    identitiesM.SignedPubkey,
		SignedSignature: identitiesM.SignedSignature,
	}
	if onetimePrekeysM != nil {
		s.OnetimePreKey = oneTimePreKey{
			KeyID:  onetimePrekeysM.KeyID,
			Pubkey: onetimePrekeysM.Pubkey,
		}
	}

	return s
}
