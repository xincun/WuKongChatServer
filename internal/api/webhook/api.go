package webhook

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/WuKongIM/WuKongChatServer/internal/api/group"
	"github.com/WuKongIM/WuKongChatServer/internal/api/user"
	"github.com/WuKongIM/WuKongChatServer/internal/common"
	"github.com/WuKongIM/WuKongChatServer/internal/config"
	"github.com/WuKongIM/WuKongChatServer/pkg/log"
	"github.com/WuKongIM/WuKongChatServer/pkg/pool"
	"github.com/WuKongIM/WuKongChatServer/pkg/util"
	"github.com/WuKongIM/WuKongChatServer/pkg/wkhook"
	"github.com/WuKongIM/WuKongChatServer/pkg/wkhttp"
	"go.uber.org/zap"
)

// Webhook Webhook
type Webhook struct {
	log.Log
	tokenCacheExpire time.Duration // token缓存失效时间
	ctx              *config.Context
	supportTypes     []common.ContentType
	db               *DB
	messageDB        *messageDB
	pushMap          map[common.DeviceType]map[string]Push
	groupService     group.IService
	userService      user.IService
	wkhook.UnimplementedWebhookServiceServer
}

// New New
func New(ctx *config.Context) *Webhook {
	pushCertDir := "configs/push/"
	pushCertName := "push_dev.p12"
	if !ctx.GetConfig().APNSDev {
		pushCertName = "push.p12"
	}
	supportTypes := getSupportTypes() // 支持推送的消息类型
	// pushMap := map[common.DeviceType]Push{
	// 	common.DeviceTypeHMS: NewHMSPush(ctx.GetConfig().HMSAppID, ctx.GetConfig().HMSAppSecret, ctx.GetConfig().AppPackage),
	// 	common.DeviceTypeIOS: NewIOSPush(ctx.GetConfig().APNSTopic, ctx.GetConfig().APNSDev, fmt.Sprintf("%s%s", pushCertDir, pushCertName), ctx.GetConfig().APNSPassword),
	// 	common.DeviceTypeMI:  NewMIPush(ctx.GetConfig().MIAppID, ctx.GetConfig().MIAppSecret, ctx.GetConfig().AppPackage),
	// }

	pushMap := map[common.DeviceType]map[string]Push{
		common.DeviceTypeHMS: {
			"com.xinbida.wukongchat": NewHMSPush(ctx.GetConfig().PushParams["com.xinbida.wukongchat.HMSAppID"], ctx.GetConfig().PushParams["com.xinbida.wukongchat.HMSAppSecret"], "com.xinbida.wukongchat"),
		},
		common.DeviceTypeMI: {
			"com.xinbida.wukongchat": NewMIPush(ctx.GetConfig().PushParams["com.xinbida.wukongchat.MIAppID"], ctx.GetConfig().PushParams["com.xinbida.wukongchat.MIAppSecret"], "com.xinbida.wukongchat", ctx.GetConfig().PushParams["com.xinbida.wukongchat.ChannelID"]),
		},
		common.DeviceTypeOPPO: {
			"com.xinbida.wukongchat": NewOPPOPush(ctx.GetConfig().PushParams["com.xinbida.wukongchat.OPPOAppID"], ctx.GetConfig().PushParams["com.xinbida.wukongchat.OPPOAppKey"], ctx.GetConfig().PushParams["com.xinbida.wukongchat.OPPOAppSecret"], ctx.GetConfig().PushParams["com.xinbida.wukongchat.OPPOMasterSecret"], ctx),
		},
		common.DeviceTypeVIVO: {
			"com.xinbida.wukongchat": NewVIVOPush(ctx.GetConfig().PushParams["com.xinbida.wukongchat.VIVOAppID"], ctx.GetConfig().PushParams["com.xinbida.wukongchat.VIVOAppKey"], ctx.GetConfig().PushParams["com.xinbida.wukongchat.VIVOAppSecret"], ctx),
		},
		common.DeviceTypeIOS: {
			"com.xinbida.wukongchat": NewIOSPush(ctx.GetConfig().APNSTopic, ctx.GetConfig().APNSDev, fmt.Sprintf("%s%s", pushCertDir, pushCertName), ctx.GetConfig().APNSPassword),
		},
	}
	return &Webhook{
		db:           NewDB(ctx.DB()),
		supportTypes: supportTypes,
		ctx:          ctx,
		Log:          log.NewTLog("Webhook"),
		pushMap:      pushMap,
		messageDB:    newMessageDB(ctx),
		groupService: group.NewService(ctx),
		userService:  user.NewService(ctx),
	}
}
func getSupportTypes() []common.ContentType {
	return []common.ContentType{common.Text, common.Image, common.GIF, common.Voice, common.Video, common.File, common.Location, common.Card, common.RedPacket, common.MultipleForward, common.VectorSticker, common.EmojiSticker}
}

// Route 路由配置
func (w *Webhook) Route(r *wkhttp.WKHttp) {
	r.POST("/v1/webhook", w.webhook)

	r.POST("/v2/webhook", w.webhook)

	r.POST("/v1/datasource", w.datasource)

	r.POST("/v1/webhook/message/notify", w.messageNotify) // 接受IM的消息通知,(TODO: 此接口需要与IM做安全认证)

	// 注册grpc服务
	wkhook.RegisterWebhookServiceServer(w.ctx.Server.GrpcServer, w)

}

func (w *Webhook) SendWebhook(ctx context.Context, req *wkhook.EventReq) (*wkhook.EventResp, error) {
	w.Debug("收到webhook grpc事件", zap.String("event", req.Event), zap.String("data", string(req.Data)))
	_, err := w.handleEvent(req.Event, req.Data)
	if err != nil {
		w.Error("处理webhook事件失败！", zap.Error(err))
		return nil, err
	}
	return &wkhook.EventResp{
		Status: wkhook.EventStatus_Success,
	}, nil
}

func (w *Webhook) messageNotify(c *wkhttp.Context) {
	var messages []MsgResp
	if err := c.BindJSON(&messages); err != nil {
		w.Error("数据格式有误！", zap.Error(err))
		c.ResponseError(errors.New("数据格式有误！"))
		return
	}
	messageIDs, err := w.handleMessageNotify(messages)
	if err != nil {
		c.ResponseError(err)
		return
	}
	c.Response(messageIDs)

}

func (w *Webhook) handleMessageNotify(messages []MsgResp) ([]string, error) {
	messageIDs := make([]string, 0, len(messages))
	if len(messages) <= 0 {
		return messageIDs, nil
	}

	confMessages := make([]*config.MessageResp, 0, len(messages))

	tx, _ := w.ctx.DB().Begin()
	defer func() {
		if err := recover(); err != nil {
			tx.RollbackUnlessCommitted()
			panic(err)
		}
	}()
	for _, message := range messages {
		messageIDs = append(messageIDs, fmt.Sprintf("%d", message.MessageID))

		if message.Header.SyncOnce == 1 || message.Header.NoPersist == 1 { // 只同步一次或有标记为不存储的消息，不进行存储
			continue
		}
		fakeChannelID := message.ChannelID
		if message.ChannelType == common.ChannelTypePerson.Uint8() {
			fakeChannelID = common.GetFakeChannelIDWith(message.FromUID, message.ChannelID)
		}
		messageM := message.toModel()
		messageM.ChannelID = fakeChannelID
		err := w.messageDB.insertOrUpdateTx(messageM, tx)
		if err != nil {
			tx.Rollback()
			w.Error("插入消息失败！", zap.Error(err))
			return nil, err
		}
		confMessages = append(confMessages, message.toConfigMessageResp())

	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		w.Error("提交事务失败！", zap.Error(err))
		return nil, err
	}

	// 通知消息监听者
	if len(confMessages) > 0 {
		w.ctx.NotifyMessagesListeners(confMessages)
	}
	return messageIDs, nil
}

func (w *Webhook) webhook(c *wkhttp.Context) {

	event := c.Query("event")

	data, err := c.GetRawData()
	if err != nil {
		w.Error("读取数据失败！", zap.Error(err))
		c.ResponseError(err)
		return
	}
	result, err := w.handleEvent(event, data)
	if err != nil {
		w.Error("事件处理失败！", zap.Error(err), zap.String("event", event), zap.String("data", string(data)))
		c.ResponseError(err)
		return
	}
	if result != nil {
		c.Response(result)
	} else {
		c.ResponseOK()
	}

}

func (w *Webhook) handleEvent(event string, data []byte) (interface{}, error) {
	if event == EventMsgOffline {
		return nil, w.handleMsgOffline(data)
	} else if event == EventOnlineStatus {
		return nil, w.handleOnlineStatus(data)
	} else if event == EventMsgNotify {
		var messages []MsgResp
		err := util.ReadJsonByByte(data, &messages)
		if err != nil {
			return nil, err
		}
		return w.handleMessageNotify(messages)
	}
	return nil, nil
}

func (w *Webhook) handleOnlineStatus(data []byte) error {
	var onlineStatusList []string
	if err := util.ReadJsonByByte(data, &onlineStatusList); err != nil {
		return err
	}
	if len(onlineStatusList) == 0 {
		return nil
	}
	onlineStatusArray := make([]config.OnlineStatus, 0)
	for _, onlineStatus := range onlineStatusList {
		onlineStatusSplits := strings.Split(onlineStatus, "-")
		if len(onlineStatusSplits) < 3 {
			continue
		}
		uid := onlineStatusSplits[0]                                         // uid
		deviceFlagI64, _ := strconv.ParseUint(onlineStatusSplits[1], 10, 64) // 设备标记
		statusI64, _ := strconv.ParseUint(onlineStatusSplits[2], 10, 64)     // 在线状态 0.离线 1.在线
		var socketID int64
		var onlineCount int
		var totalOnlineCount int
		if len(onlineStatusSplits) >= 6 {
			socketID, _ = strconv.ParseInt(onlineStatusSplits[3], 10, 64)             // socketID
			onlineCountI64, _ := strconv.ParseInt(onlineStatusSplits[4], 10, 64)      // 在线数量 当前DeviceFlag下的在线设备数量
			totalOnlineCountI64, _ := strconv.ParseInt(onlineStatusSplits[5], 10, 64) // 在线数量 用户下所有设备的在线数量
			onlineCount = int(onlineCountI64)
			totalOnlineCount = int(totalOnlineCountI64)
		}

		status := int(statusI64)
		deviceFlag := uint8(deviceFlagI64)

		onlineStatusArray = append(onlineStatusArray, config.OnlineStatus{
			UID:              uid,
			DeviceFlag:       deviceFlag,
			Online:           status == 1,
			SocketID:         socketID,
			OnlineCount:      onlineCount,
			TotalOnlineCount: totalOnlineCount,
		})

	}
	listeners := w.ctx.GetAllOnlineStatusListeners()
	if len(listeners) > 0 {
		for _, listener := range listeners {
			listener(onlineStatusArray)
		}
	}

	return nil
}

func (w *Webhook) handleMsgOffline(data []byte) error {
	var msgResp msgOfflineNotify
	err := util.ReadJsonByByte(data, &msgResp)
	if err != nil {
		return err
	}
	w.Debug("收到离线消息->", zap.Any("msg", msgResp))

	var toUids []string
	if msgResp.Compress == "gzip" {
		if len(msgResp.CompresssToUIDs) > 0 {
			gReader, err := gzip.NewReader(bytes.NewReader(msgResp.CompresssToUIDs))
			if err != nil {
				w.Error("解码gzip失败！", zap.String("compresssToUIDs", string(msgResp.CompresssToUIDs)))
				return err
			}
			defer gReader.Close()
			compresssToUIDBytes, err := ioutil.ReadAll(gReader)
			if err != nil {
				w.Error("读取gzip压缩数据失败！", zap.Error(err))
				return err
			}
			err = util.ReadJsonByByte(compresssToUIDBytes, &toUids)
			if err != nil {
				w.Error("")
				return err
			}
		}

	} else {
		toUids = msgResp.ToUIDS
	}

	if len(toUids) == 0 {
		return nil
	}

	return w.pushTo(msgResp, toUids)
}

func (w *Webhook) pushTo(msgResp msgOfflineNotify, toUids []string) error {
	setting := config.SettingFromUint8(msgResp.Setting)
	isVideoCall := false
	if !setting.Signal { // 只解析未加密的消息
		contentMap, err := util.JsonToMap(string(msgResp.Payload))
		if err != nil {
			w.Error("消息payload格式有误！", zap.Error(err), zap.String("payload", string(msgResp.Payload)))
			return err
		}
		msgResp.PayloadMap = contentMap
		if contentMap["type"] == nil {
			return errors.New("type为空！")
		}
		contentTypeInt64, _ := contentMap["type"].(json.Number).Int64()
		contentType := common.ContentType(contentTypeInt64)
		msgResp.ContentType = int(contentType)
	}

	var err error
	var users []*user.Resp
	userSettings := make([]*user.SettingResp, 0)
	groupSettings := make([]*group.SettingResp, 0)
	if !isVideoCall { // 音视频消息不检查设置，直接推送
		// 查询免打扰
		// 查询用户总设置
		users, err = w.userService.GetUsers(toUids)
		if err != nil {
			w.Error("查询推送用户信息错误", zap.Error(err))
			return nil
		}
		if msgResp.ChannelType == common.ChannelTypePerson.Uint8() {
			// 查询用户对某人设置
			if msgResp.FromUID != "" && len(toUids) > 0 {
				uids := make([]string, 0)
				uids = append(uids, msgResp.FromUID)
				userSettings, err = w.userService.GetUserSettings(uids, toUids[0])
				if err != nil {
					w.Error("查询用户对某人设置错误", zap.Error(err))
					return nil
				}
			}
		} else {
			// 查询一批用户对某个群的设置
			groupSettings, err = w.groupService.GetSettingsWithUIDs(msgResp.ChannelID, toUids)
			if err != nil {
				w.Error("查询一批用户对某群设置错误", zap.Error(err))
				return nil
			}
		}
	}

	for _, toUID := range toUids {
		if !isVideoCall {
			if !w.allowPush(users, userSettings, groupSettings, toUID) {
				continue
			}
		} else {
			w.Info("开始音视频推送...")
		}

		w.ctx.PushPool.Work <- &pool.Job{
			Data: map[string]interface{}{
				"toUID": toUID,
				"msg":   msgResp,
			},
			JobFunc: func(id int64, data interface{}) {
				dataMap := data.(map[string]interface{})
				toUID := dataMap["toUID"].(string)
				msgResp := dataMap["msg"].(msgOfflineNotify)
				result, err := w.push(toUID, msgResp)
				if err != nil {
					w.Debug("推送失败！", zap.String("uid", toUID), zap.String("deviceType", result.deviceType), zap.String("deviceToken", result.deviceToken), zap.Error(err))
				} else {
					w.Debug("推送成功！", zap.String("uid", toUID), zap.String("deviceType", result.deviceType), zap.String("deviceToken", result.deviceToken))
				}
			},
		}

	}
	return nil
}

// 是否允许推送
func (w *Webhook) allowPush(users []*user.Resp, userSettings []*user.SettingResp, groupSettings []*group.SettingResp, toUID string) bool {
	isPush := true
	if len(users) > 0 {
		for _, user := range users {
			if user.UID == toUID {
				if user.NewMsgNotice == 0 {
					isPush = false
				}
				break
			}
		}
	}
	if isPush && userSettings != nil && len(userSettings) > 0 {
		for _, userSetting := range userSettings {
			if userSetting.UID == toUID {
				if userSetting.Mute == 1 {
					isPush = false
				}
				break
			}

		}
	}
	if isPush && groupSettings != nil && len(groupSettings) > 0 {
		for _, groupSetting := range groupSettings {
			if groupSetting.UID == toUID {
				if groupSetting.Mute == 1 {
					isPush = false
				}
				break
			}
		}
	}
	return isPush
}

func (w *Webhook) push(toUID string, msgResp msgOfflineNotify) (pushResp, error) {

	var deviceMap map[string]string
	deviceMap, err := w.ctx.GetRedisConn().Hgetall(fmt.Sprintf("%s%s", common.UserDeviceTokenPrefix, toUID))
	if err != nil {
		return pushResp{}, err
	}
	if len(deviceMap) <= 0 {
		return pushResp{}, errors.New("用户设备信息不存在！")
	}
	deviceToken := deviceMap["device_token"]
	deviceType := deviceMap["device_type"]
	bundleID := deviceMap["bundle_id"]

	w.Debug("开始推送", zap.String("uid", toUID), zap.String("deviceType", deviceType), zap.String("deviceToken", deviceToken))

	if w.pushMap[common.DeviceType(deviceType)] == nil {
		return pushResp{
			deviceType:  deviceType,
			deviceToken: deviceToken,
		}, errors.New("不支持的推送设备！")
	}
	pusher := w.pushMap[common.DeviceType(deviceType)][bundleID]
	if pusher == nil {
		w.Warn("不支持的推送设备！", zap.String("deviceType", deviceType), zap.String("uid", toUID))
		return pushResp{
			deviceType:  deviceType,
			deviceToken: deviceToken,
		}, errors.New("不支持的推送设备！")
	}
	payload, err := pusher.GetPayload(msgResp, w.ctx, toUID)
	if err != nil {
		return pushResp{
			deviceType:  deviceType,
			deviceToken: deviceToken,
		}, err
	}
	err = pusher.Push(deviceToken, payload)
	if err != nil {
		return pushResp{
			deviceType:  deviceType,
			deviceToken: deviceToken,
		}, err
	}
	return pushResp{
		deviceType:  deviceType,
		deviceToken: deviceToken,
	}, nil
}

func (w *Webhook) containSupportType(contentType common.ContentType) bool {
	for _, t := range w.supportTypes {
		if t == contentType {
			return true
		}
	}
	return false
}

// Event Event
type Event struct {
	Event string      `json:"event"` // 事件标示
	Data  interface{} `json:"data"`  // 事件数据
}

type messageHeader struct {
	NoPersist int `json:"no_persist"` // 是否不持久化
	RedDot    int `json:"red_dot"`    // 是否显示红点
	SyncOnce  int `json:"sync_once"`  // 此消息只被同步或被消费一次
}

// MsgResp MsgResp
type MsgResp struct {
	Header      messageHeader `json:"header"`  // 消息头部
	Setting     uint8         `json:"setting"` // setting
	ClientMsgNo string        `json:"client_msg_no"`
	MessageID   int64         `json:"message_id"`   // 服务端的消息ID(全局唯一)
	MessageSeq  uint32        `json:"message_seq"`  // 消息序列号 （用户唯一，有序递增）
	FromUID     string        `json:"from_uid"`     // 发送者UID
	ToUID       string        `json:"to_uid"`       // 接受者uid
	ChannelID   string        `json:"channel_id"`   // 频道ID
	ChannelType uint8         `json:"channel_type"` // 频道类型
	Timestamp   int32         `json:"timestamp"`    // 服务器消息时间戳(10位，到秒)
	Payload     []byte        `json:"payload"`      // 消息内容
	ContentType int           // 消息正文类型
	PayloadMap  map[string]interface{}
}

func (m *MsgResp) toModel() *messageModel {

	setting := config.SettingFromUint8(m.Setting)

	var signal uint8 = 0
	if setting.Signal {
		signal = 1
	}
	return &messageModel{
		MessageID:   fmt.Sprintf("%d", m.MessageID),
		MessageSeq:  int64(m.MessageSeq),
		ClientMsgNo: m.ClientMsgNo,
		Header:      util.ToJson(m.Header),
		Setting:     m.Setting,
		Signal:      signal,
		FromUID:     m.FromUID,
		ChannelID:   m.ChannelID,
		ChannelType: m.ChannelType,
		Timestamp:   m.Timestamp,
		Payload:     string(m.Payload),
		IsDeleted:   0,
	}
}

func (m *MsgResp) toConfigMessageResp() *config.MessageResp {
	return &config.MessageResp{
		MessageID:   m.MessageID,
		MessageSeq:  m.MessageSeq,
		ClientMsgNo: m.ClientMsgNo,
		Header: config.MsgHeader{
			NoPersist: m.Header.NoPersist,
			RedDot:    m.Header.RedDot,
			SyncOnce:  m.Header.SyncOnce,
		},
		FromUID:     m.FromUID,
		ToUID:       m.ToUID,
		ChannelID:   m.ChannelID,
		ChannelType: m.ChannelType,
		Timestamp:   m.Timestamp,
		Payload:     m.Payload,
	}
}

type msgOfflineNotify struct {
	MsgResp
	ToUIDS          []string `json:"to_uids"`                    // im服务推离线的时候接受uid是一个集合
	Compress        string   `json:"compress,omitempty"`         // 压缩ToUIDs 如果为空 表示不压缩 为gzip则采用gzip压缩
	CompresssToUIDs []byte   `json:"compress_to_uids,omitempty"` // 已压缩的to_uids
	SourceID        int64    `json:"source_id,omitempty"`        // 来源节点ID
}

type pushResp struct {
	deviceToken string
	deviceType  string
}
