package utils

import "regexp"

// yuqueLinkRegex 匹配语雀来源链接，如 <https://www.yuque.com/lemonlove7/wiki/eccmbpwmaq7x9at5>
// 支持 http/https，任意路径
var yuqueLinkRegex = regexp.MustCompile(`\n?<https?://www\.yuque\.com/[^\s>]+>`)

// CleanYuqueLinks 清理 Markdown 内容中的语雀来源链接
func CleanYuqueLinks(content string) string {
	return yuqueLinkRegex.ReplaceAllString(content, "")
}
