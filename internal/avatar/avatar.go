// Package avatar 根据 Gravatar/Cravatar 协议生成头像 URL。
package avatar

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"
)

const (
	// baseURL Cravatar 服务地址（Gravatar 国内镜像）。
	baseURL = "https://cravatar.com/avatar"
	// defaultSize 返回图片的像素尺寸（前端 CSS 会缩放到实际显示尺寸）。
	defaultSize = 128
	// noreplyDomain 用于无邮箱用户生成伪邮箱，确保 identicon 不会撞到真实注册头像。
	noreplyDomain = "@noreply.local"
)

// AvatarURL 根据 email（可空）和 username 生成 Cravatar 头像 URL。
// 有邮箱则用邮箱的 MD5；无邮箱则用 "<username>@noreply.local" 生成，
// Cravatar 会基于该 hash 返回一张独特的 identicon 几何头像兜底。
func AvatarURL(email *string, username string) string {
	var raw string
	if email != nil && *email != "" {
		raw = *email
	} else {
		raw = strings.ToLower(username) + noreplyDomain
	}
	hash := md5Hex(raw)
	return baseURL + "/" + hash + "?d=identicon&s=" + strconv.Itoa(defaultSize)
}

// md5Hex 计算字符串的 MD5 哈希并返回小写十六进制。
// 遵循 Gravatar 规范：邮箱先转小写并去除首尾空格。
func md5Hex(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
