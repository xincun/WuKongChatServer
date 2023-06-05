package config

import (
	"fmt"
	"hash/crc32"
	"os"
	"strconv"
	"strings"
	"time"

	limlog "github.com/WuKongIM/WuKongChatServer/pkg/log"
	"github.com/WuKongIM/WuKongChatServer/pkg/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Mode string

const (
	//debug 模式
	DebugMode Mode = "debug"
	// 正式模式
	ReleaseMode Mode = "release"
	// 压力测试模式
	BenchMode Mode = "bench"
)

// UploadService UploadService
type UploadService string

const (
	// UploadServiceAliyunOSS 阿里云oss上传服务
	UploadServiceAliyunOSS UploadService = "aliyunOSS"
	// UploadServiceSeaweedFS seaweedfs(https://github.com/chrislusf/seaweedfs)
	UploadServiceSeaweedFS UploadService = "seaweedFS"
	// UploadServiceMinio minio
	UploadServiceMinio UploadService = "minio"
)

func (u UploadService) String() string {
	return string(u)
}

type TablePartitionConfig struct {
	MessageTableCount         int // 消息表数量
	MessageUserEditTableCount int // 用户消息编辑表
	ChannelOffsetTableCount   int // 频道偏移表
}

func newTablePartitionConfig() TablePartitionConfig {

	return TablePartitionConfig{
		MessageTableCount:         5,
		MessageUserEditTableCount: 3,
		ChannelOffsetTableCount:   3,
	}
}

// Config 配置信息
type Config struct {
	AppID      string // APP ID
	AppName    string // APP名称
	Version    string // 版本
	Addr       string // 服务监听地址 x.x.x.x:8080
	SSLAddr    string // ssl 监听地址
	BaseURL    string // 本服务的对外的基础地址
	H5BaseURL  string // h5页面的基地址 如果没有配置默认未 BaseURL + /web
	APIBaseURL string // api的基地址 如果没有配置默认未 BaseURL + /v1
	Mode       Mode   // 模式 debug 测试 release 正式 bench 压力测试
	Logger     struct {
		Dir     string // 日志存储目录
		Level   zapcore.Level
		LineNum bool // 是否显示代码行数
	}

	MySQLAddr            string        // mysql的连接信息
	SQLDir               string        // 数据库脚本路径
	Migration            bool          // 是否合并数据库
	RedisAddr            string        // redis地址
	AsynctaskRedisAddr   string        // 异步任务的redis地址 不写默认为RedisAddr的地址
	NodeID               int           //  节点ID 节点ID需要小于1024
	Test                 bool          // 是否是测试模式
	IMURL                string        // im基地址
	MessagePoolSize      int64         // 发消息任务池大小
	PushPoolSize         int64         // 推送任务池大小
	UploadService        UploadService // 上传服务
	UploadURL            string        // 上传地址
	FileDownloadURL      string        // 文件下载url
	MinioAccessKeyID     string        //minio accessKeyID
	MinioSecretAccessKey string        //minio secretAccessKey
	DefaultAvatar        string        // 默认头像
	QRCodeInfoURL        string        // 获取二维码信息的URL
	VisitorUIDPrefix     string        // 访客uid的前缀
	OnlineStatusOn       bool          // 是否开启在线状态显示
	TimingWheelTick      duration      // The time-round training interval must be 1ms or more
	TimingWheelSize      int64         // Time wheel size
	// ---------- 缓存设置 ----------
	TokenCachePrefix            string // token缓存前缀
	LoginDeviceCachePrefix      string // 登录设备缓存前缀
	LoginDeviceCacheExpire      time.Duration
	UIDTokenCachePrefix         string        // uidtoken缓存前缀
	FriendApplyTokenCachePrefix string        // 申请好友的token的前缀
	FriendApplyExpire           time.Duration // 好友申请过期时间
	TokenExpire                 time.Duration // token失效时间
	PayTokenExpire              time.Duration // 支付token失效时间
	NameCacheExpire             time.Duration // 名字缓存过期时间
	// -------- 推送 ---------
	APNSDev      bool   // apns是否是开发模式
	APNSPassword string // apns的密码
	APNSTopic    string
	// 华为推送
	HMSAppID     string // 华为app id
	HMSAppSecret string // 华为app Secret
	// 小米推送
	MIAppID     string // 小米app id
	MIAppSecret string // 小米app Secret

	// 推送保存的key信息
	PushParams map[string]string

	ElasticsearchURL string // elasticsearch 地址

	FileHelperName      string // 文件上传助手的名称
	FileHelperAvatar    string // 文件上传助手的头像
	PushContentDetailOn bool   // 推送是否显示正文详情(如果为false，则只显示“您有一条新的消息” 默认为true)

	// ---------- 短信运营商 ----------
	SMSCode                string // 模拟的短信验证码
	SMSProvider            SMSProvider
	UniSMS                 UnismsConfig
	AliyunSMS              AliyunSMSConfig
	AliyunInternationalSMS AliyunInternationalSMSConfig // 阿里云国际短信

	SystemUID      string //系统账号uid
	SystemFileUID  string // 系统文件uid
	SystemGroupID  string //系统群ID
	SystemAdminUID string //系统管理员账号
	WelcomeMessage string //登录注册欢迎语
	// ---------- tracing ----------
	TracingOn  bool   // 是否开启tracing
	TracerAddr string // tracer的地址

	SystemGroupName string // 系统群的名字

	TablePartitionConfig TablePartitionConfig

	SupportEmail     string // 技术支持的邮箱地址
	SupportEmailSmtp string // 技术支持的邮箱的smtp
	SupportEmailPwd  string // 邮箱密码

	WebLoginURL string // web登录地址

	// ---------- 系统配置  由系统生成,无需用户配置 ----------
	AppRSAPrivateKey string
	AppRSAPubKey     string

	// ---------- robot ----------
	RobotMessageExpire     time.Duration // 机器人消息过期时间
	RobotInlineQueryExpire time.Duration // 机器人inlineQuery事件过期时间
	RobotEventPoolSize     int64         // 机器人事件池大小

	// ----------  微信 ----------
	WXAppID  string // 微信appid 在开放平台内
	WXSecret string

	// ---------- 其他 ----------
	MessageSaveAcrossDevice bool   // 消息是否跨设备保存（换设备登录消息是否还能同步到老消息）
	RegisterOnlyChina       bool   // 是否仅仅中国手机号可以注册
	EndToEndEncryptionOn    bool   // 是否开启端对端加密 默认开启
	GRPCAddr                string // grpc的通信地址 （建议内网通信）
	RegisterOff             bool   // 是否关闭注册
	StickerAddOffOfRegister bool   // 是否关闭注册添加表情

	GroupUpgradeWhenMemberCount int  // 当成员数量大于此配置时 自动升级为超级群 默认为 1000
	ShortnoNumOn                bool // 是否开启数字短编号
	ShortnoNumLen               int  // 数字短编号长度
	ShortnoEditOff              bool // 是否关闭短编号编辑
	PhoneSearchOff              bool // 是否关闭手机号搜索

	GithubAPI string // github api地址
}

// New New
func New() *Config {
	extIP, err := util.GetExternalIP()
	if err != nil {
		limlog.Warn("获取外网IP失败！", zap.Error(err))
	}

	cfg := &Config{
		AppID:                       "wukongchat",
		AppName:                     GetEnv("AppName", "悟空聊天"),
		Addr:                        GetEnv("Addr", ":8080"),
		Version:                     "2.0.0",
		SSLAddr:                     GetEnv("SSLAddr", ""),
		Migration:                   true,
		BaseURL:                     strings.ReplaceAll(GetEnv("BaseURL", "http://127.0.0.1:8080"), "{EXT_IP}", extIP),
		H5BaseURL:                   GetEnv("H5BaseURL", ""),
		APIBaseURL:                  GetEnv("APIBaseURL", ""),
		QRCodeInfoURL:               "v1/qrcode/:code",
		NodeID:                      1,
		TokenExpire:                 time.Hour * 24 * 30,
		PayTokenExpire:              time.Minute * 5,
		MySQLAddr:                   GetEnv("MySQLAddr", "root:demo@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=true"),
		SQLDir:                      "configs/sql",
		RedisAddr:                   GetEnv("RedisAddr", "127.0.0.1:6379"),
		Test:                        GetEnvBool("Test", false),
		IMURL:                       GetEnv("IMURL", "http://127.0.0.1:5000"),
		TokenCachePrefix:            "token:",
		LoginDeviceCachePrefix:      "login_device:",
		LoginDeviceCacheExpire:      time.Minute * 5,
		UIDTokenCachePrefix:         "uidtoken:",
		FriendApplyTokenCachePrefix: "friend_token:",
		FriendApplyExpire:           time.Hour * 24 * 15,
		MessagePoolSize:             100,
		PushPoolSize:                GetEnvInt64("PushPoolSize", 100),
		APNSDev:                     GetEnvBool("APNSDev", true),
		APNSPassword:                GetEnv("APNSPassword", "123456"),
		APNSTopic:                   GetEnv("APNSTopic", "com.xinbida.wukongchat"),
		NameCacheExpire:             time.Hour * 24 * 7,
		SMSProvider:                 SMSProvider(GetEnv("SMSProvider", string(SMSProviderAliyun))),
		VisitorUIDPrefix:            "_vt_",
		OnlineStatusOn:              GetEnvBool("OnlineStatusOn", true),
		SystemUID:                   "u_10000",
		SystemFileUID:               "fileHelper",
		SystemGroupID:               "g_10000",
		SystemAdminUID:              "admin",
		Mode:                        Mode(GetEnv("Mode", string(ReleaseMode))),
		Logger: struct {
			Dir     string
			Level   zapcore.Level
			LineNum bool
		}{
			Dir:     GetEnv("Logger.Dir", ""),
			Level:   zapcore.Level(GetEnvInt("Logger.Level", int(zapcore.InfoLevel))),
			LineNum: GetEnvBool("Logger.LineNum", false),
		},
		TimingWheelTick: duration{
			Duration: time.Millisecond * 10,
		},
		TimingWheelSize:      100,
		MinioAccessKeyID:     GetEnv("MinioAccessKeyID", ""),
		MinioSecretAccessKey: GetEnv("MinioSecretAccessKey", ""),
		SMSCode:              GetEnv("SMSCode", ""),
		UniSMS: UnismsConfig{
			AccessKeyID: GetEnv("UniSMS.AccessKeyID", ""),
			Signature:   GetEnv("UniSMS.Signature", ""),
		},
		AliyunSMS: AliyunSMSConfig{
			AccessKeyID:  GetEnv("AliyunSMS.AccessKeyID", ""),
			AccessSecret: GetEnv("AliyunSMS.AccessSecret", ""),
			TemplateCode: GetEnv("AliyunSMS.TemplateCode", ""),
			SignName:     GetEnv("AliyunSMS.SignName", ""),

			// AccessKeyID:  GetEnv("AliyunSMS.AccessKeyID", "LTAI5t8oNMBiXGnZfbiSZEPD"),
			// AccessSecret: GetEnv("AliyunSMS.AccessSecret", "3PurLYINDcr27N30ljlIoSeHgpa2YP"),
			// TemplateCode: GetEnv("AliyunSMS.TemplateCode", "SMS_11460360"),
			// SignName:     GetEnv("AliyunSMS.SignName", "Kchat"),
		},
		AliyunInternationalSMS: AliyunInternationalSMSConfig{
			AccessKeyID:  GetEnv("AliyunSMS.AccessKeyID", ""),
			AccessSecret: GetEnv("AliyunSMS.AccessSecret", ""),
			SignName:     GetEnv("AliyunSMS.SignName", ""),
		},
		PushContentDetailOn:     GetEnvBool("PushContentDetailOn", true),
		UploadService:           UploadService(GetEnv("UploadService", UploadServiceMinio.String())),
		UploadURL:               GetEnv("UploadURL", "http://127.0.0.1:9000"),
		FileDownloadURL:         GetEnv("FileDownloadURL", "http://127.0.0.1:9000"),
		DefaultAvatar:           "configs/assets/avatar.png",
		ElasticsearchURL:        "http://elasticsearch:9200",
		FileHelperName:          "文件传输助手",
		FileHelperAvatar:        "https://timgsa.baidu.com/timg?image&quality=80&size=b9999_10000&sec=1589347370173&di=5c94e79d1cfe1e493460144d7f525abb&imgtype=0&src=http%3A%2F%2Fr.sinaimg.cn%2Flarge%2Farticle%2Fa26f4cee14bb0b95d05a2e0b20b26c19.png",
		TracingOn:               GetEnvBool("TracingOn", false),
		TracerAddr:              GetEnv("TracerAddr", ""),
		SystemGroupName:         GetEnv("SystemGroupName", "悟空聊天压力测试群"),
		SupportEmail:            GetEnv("SupportEmail", "support@githubim.com"),
		SupportEmailSmtp:        GetEnv("SupportEmailSmtp", "smtp.exmail.qq.com:25"),
		SupportEmailPwd:         GetEnv("SupportEmailPwd", ""),
		WebLoginURL:             GetEnv("WebLoginURL", "http://localhost:3000/login"),
		MessageSaveAcrossDevice: GetEnvBool("MessageSaveAcrossDevice", true),
		// --------- 微信支付 ----------
		WXAppID:  GetEnv("WXAppID", ""),
		WXSecret: GetEnv("WXSecret", ""),
		// --------- robot ----------
		RobotMessageExpire:     time.Hour * 12,
		RobotInlineQueryExpire: time.Second * 10,
		RobotEventPoolSize:     100,
		// ---------- other ----------
		RegisterOnlyChina:           GetEnvBool("RegisterOnlyChina", false),
		EndToEndEncryptionOn:        GetEnvBool("EndToEndEncryptionOn", true),
		GRPCAddr:                    GetEnv("GRPCAddr", "0.0.0.0:6979"),
		RegisterOff:                 GetEnvBool("RegisterOff", false),
		StickerAddOffOfRegister:     GetEnvBool("StickerAddOffOfRegister", false),
		GroupUpgradeWhenMemberCount: GetEnvInt("GroupUpgradeWhenMemberCount", 1000),
		ShortnoNumOn:                GetEnvBool("ShortnoNumOn", false),
		ShortnoNumLen:               GetEnvInt("ShortnoNumLen", 7),
		ShortnoEditOff:              GetEnvBool("ShortnoEditOff", false),
		PhoneSearchOff:              GetEnvBool("PhoneSearchOff", false),
		GithubAPI:                   GetEnv("GithubAPI", "https://api.github.com"),
	}

	cfg.TablePartitionConfig = newTablePartitionConfig()

	cfg.WelcomeMessage = GetEnv("WelcomeMessage", fmt.Sprintf("欢迎你来到%s。\n%s安全中心提醒您：冒充军人公职人员，谈恋爱交友，金融期货投资，套路贷款，彩票杀猪盘，刷单点赞，信用卡代还、手工兼职、抖音点赞、彩票投资、转账汇款等等都属于电信诈骗，上当受骗第一时间报警。为了加强自身安全防范意识请及时下载“国家反诈中心”APP。更多防诈信息请移步 https://fzapph5.chanct.cn \n%s不会向你索要手机号、银行卡、身份证、短信验证码等敏感信息，如果你遇到诸类情况可点击头像进行举报", cfg.AppName, cfg.AppName, cfg.AppName))

	StringEnv(&cfg.AsynctaskRedisAddr, "AsynctaskRedisAddr")
	if cfg.AsynctaskRedisAddr == "" {
		cfg.AsynctaskRedisAddr = cfg.RedisAddr
	}
	if cfg.H5BaseURL == "" {
		cfg.H5BaseURL = fmt.Sprintf("%s/web", cfg.BaseURL)
	}
	if cfg.APIBaseURL == "" {
		cfg.APIBaseURL = fmt.Sprintf("%s/v1", cfg.BaseURL)
	}
	cfg.configureLog()

	return cfg
}

func (c *Config) configureLog() {
	logLevel := c.Logger.Level
	// level
	if logLevel == zap.InfoLevel { // 没有设置
		if c.Mode == DebugMode {
			logLevel = zapcore.DebugLevel
		} else {
			logLevel = zapcore.InfoLevel
		}
		c.Logger.Level = logLevel
	}
}

// GetAvatarPath 获取用户头像path
func (c *Config) GetAvatarPath(uid string) string {
	return fmt.Sprintf("users/%s/avatar", uid)
}

// GetGroupAvatarFilePath 获取群头像上传路径
func (c *Config) GetGroupAvatarFilePath(groupNo string) string {
	avatarID := crc32.ChecksumIEEE([]byte(groupNo)) % 100
	return fmt.Sprintf("group/%d/%s.png", avatarID, groupNo)
}

// GetCommunityAvatarFilePath 获取社区头像上传路径
func (c *Config) GetCommunityAvatarFilePath(communityNo string) string {
	avatarID := crc32.ChecksumIEEE([]byte(communityNo)) % 100
	return fmt.Sprintf("community/%d/%s.png", avatarID, communityNo)
}

// GetCommunityCoverFilePath 获取社区封面上传路径
func (c *Config) GetCommunityCoverFilePath(communityNo string) string {
	avatarID := crc32.ChecksumIEEE([]byte(communityNo)) % 100
	return fmt.Sprintf("community/%d/%s_cover.png", avatarID, communityNo)
}

// IsVisitorChannel 是访客频道
func (c *Config) IsVisitorChannel(uid string) bool {

	return strings.HasSuffix(uid, "@ht")
}

// 获取客服频道真实ID
func (c *Config) GetCustomerServiceChannelID(channelID string) (string, bool) {
	if !strings.Contains(channelID, "|") {
		return "", false
	}
	channelIDs := strings.Split(channelID, "|")
	return channelIDs[1], true
}

// 获取客服频道的访客id
func (c *Config) GetCustomerServiceVisitorUID(channelID string) (string, bool) {
	if !strings.Contains(channelID, "|") {
		return "", false
	}
	channelIDs := strings.Split(channelID, "|")
	return channelIDs[0], true
}

// 组合客服ID
func (c *Config) ComposeCustomerServiceChannelID(vid string, channelID string) string {
	return fmt.Sprintf("%s|%s", vid, channelID)
}

// IsVisitor 是访客uid
func (c *Config) IsVisitor(uid string) bool {

	return strings.HasPrefix(uid, c.VisitorUIDPrefix)
}

// GetEnv 成环境变量里获取
func GetEnv(key string, defaultValue string) string {
	v := os.Getenv(key)
	if strings.TrimSpace(v) == "" {
		return defaultValue
	}
	return v
}

// GetEnvBool 成环境变量里获取
func GetEnvBool(key string, defaultValue bool) bool {
	v := os.Getenv(key)
	if strings.TrimSpace(v) == "" {
		return defaultValue
	}
	if v == "true" {
		return true
	}
	return false
}

// GetEnvInt64 环境变量获取
func GetEnvInt64(key string, defaultValue int64) int64 {
	v := os.Getenv(key)
	if strings.TrimSpace(v) == "" {
		return defaultValue
	}
	i, _ := strconv.ParseInt(v, 10, 64)
	return i
}

// GetEnvInt 环境变量获取
func GetEnvInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if strings.TrimSpace(v) == "" {
		return defaultValue
	}
	i, _ := strconv.ParseInt(v, 10, 64)
	return int(i)
}

// GetEnvFloat64 环境变量获取
func GetEnvFloat64(key string, defaultValue float64) float64 {
	v := os.Getenv(key)
	if strings.TrimSpace(v) == "" {
		return defaultValue
	}
	i, _ := strconv.ParseFloat(v, 64)
	return i
}

// StringEnv StringEnv
func StringEnv(v *string, key string) {
	vv := os.Getenv(key)
	if vv != "" {
		*v = vv
	}
}

// BoolEnv 环境bool值
func BoolEnv(b *bool, key string) {
	value := os.Getenv(key)
	if strings.TrimSpace(value) != "" {
		if value == "true" {
			*b = true
		} else {
			*b = false
		}
	}
}

// SMSProvider 短信供应者
type SMSProvider string

const (
	// SMSProviderAliyun aliyun
	SMSProviderAliyun SMSProvider = "aliyun"
	SMSProviderUnisms SMSProvider = "unisms" // 联合短信(https://unisms.apistd.com/docs/api/send/)
)

// AliyunSMSConfig 阿里云短信
type AliyunSMSConfig struct {
	AccessKeyID  string // aliyun的AccessKeyID
	AccessSecret string // aliyun的AccessSecret
	TemplateCode string // aliyun的短信模版
	SignName     string // 签名
}

// UnismsConfig unisms短信
type UnismsConfig struct {
	Signature   string
	AccessKeyID string
}

// AliyunInternationalSMSConfig 阿里云短信
type AliyunInternationalSMSConfig struct {
	AccessKeyID  string // aliyun的AccessKeyID
	AccessSecret string // aliyun的AccessSecret
	SignName     string // 签名
}

// TransferConfig 转账配置
type TransferConfig struct {
	Expire time.Duration
}

// RedpacketConfig 红包配置
type RedpacketConfig struct {
	Expire time.Duration
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
