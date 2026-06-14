package service

import "github.com/full-finger/user-system/internal/apperror"

// ValidCoverThemes 允许的封面主题 key 集合。空串（无封面）由 *string == "" 表达，不在此集合中。
// 新增主题时同步更新前端 web/src/config/coverThemes.js。
var ValidCoverThemes = map[string]struct{}{
	"sunset":   {},
	"lavender": {},
	"ocean":    {},
	"forest":   {},
	"amber":    {},
	"blossom":  {},
	"mint":     {},
	"dawn":     {},
}

// normalizeCoverTheme 规范化封面主题输入。
//   - nil：不更新，返回 nil。
//   - ""：明确清空为"无封面"，原样返回空串。
//   - 非空：校验是否在白名单内，非法返回 BadRequest。
func normalizeCoverTheme(p *string) (*string, error) {
	if p == nil {
		return nil, nil
	}
	if *p == "" {
		empty := ""
		return &empty, nil
	}
	if _, ok := ValidCoverThemes[*p]; !ok {
		return nil, apperror.BadRequest("无效的封面主题")
	}
	v := *p
	return &v, nil
}

// MottoMaxLength 座右铭最大长度（字符数）。
const MottoMaxLength = 100

// normalizeMotto 规范化座右铭输入。
//   - nil：不更新，返回 nil。
//   - ""：明确清空为未设置，原样返回空串。
//   - 非空：校验长度不超过 MottoMaxLength，超出返回 BadRequest。
func normalizeMotto(p *string) (*string, error) {
	if p == nil {
		return nil, nil
	}
	if *p == "" {
		empty := ""
		return &empty, nil
	}
	if len([]rune(*p)) > MottoMaxLength {
		return nil, apperror.BadRequest("座右铭长度不超过 100 字符")
	}
	v := *p
	return &v, nil
}
