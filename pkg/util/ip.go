package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

// GetExternalIP 获取本服务器的外网IP
func GetExternalIP() (string, error) {
	resp, err := http.Get("https://ipw.cn/api/ip/myip")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	resultBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(resultBytes)), nil
}

// GetClientPublicIP 尽最大努力实现获取客户端公网 IP 的算法。
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx 或 haproxy）可以正常工作。
func GetClientPublicIP(r *http.Request) string {
	var ip string
	for _, ip = range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			return ip
		}
	}
	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

// GetIPAddress 通过IP获取地址
func GetIPAddress(ip string) (province string, city string, err error) {
	var resp *http.Response
	resp, err = http.Get(fmt.Sprintf("https://restapi.amap.com/v3/ip?key=7e30415c3e9ce73d93d20189b9539be8&ip=%s", ip))
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("查询地址失败！")
		return
	}
	var data []byte
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var resultMap map[string]interface{}
	resultMap, err = JsonToMap(string(data))
	if err != nil {
		return
	}
	provinceObj := resultMap["province"]
	cityObj := resultMap["city"]
	if provinceObj != nil && cityObj != nil {
		var ok bool
		province, ok = provinceObj.(string)
		if !ok {
			return
		}
		city, ok = cityObj.(string)
		if !ok {
			return
		}
		return
	}
	return
}
