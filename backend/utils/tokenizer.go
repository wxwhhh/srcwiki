package utils

import (
	"strings"
	"unicode"
)

// Tokenize 对文本进行中文分词（轻量级方案：中文字符 unigram + bigram）
// 英文单词保持原样，中文按字符和二元组切分
func Tokenize(text string) string {
	if len(text) == 0 {
		return text
	}

	runes := []rune(text)
	var tokens []string
	var engBuf strings.Builder

	flushEng := func() {
		if engBuf.Len() > 0 {
			tokens = append(tokens, engBuf.String())
			engBuf.Reset()
		}
	}

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if unicode.IsLetter(r) && r < 0x4E00 { // ASCII/拉丁字母
			engBuf.WriteRune(r)
			continue
		}
		if unicode.IsDigit(r) {
			engBuf.WriteRune(r)
			continue
		}
		flushEng()

		// CJK 统一汉字范围
		if r >= 0x4E00 && r <= 0x9FFF || r >= 0x3400 && r <= 0x4DBF || r >= 0x20000 && r <= 0x2A6DF {
			// unigram
			tokens = append(tokens, string(r))
			// bigram
			if i+1 < len(runes) {
				next := runes[i+1]
				if next >= 0x4E00 && next <= 0x9FFF || next >= 0x3400 && next <= 0x4DBF {
					tokens = append(tokens, string(r)+string(next))
				}
			}
		}
		// 其他字符（标点等）跳过
	}
	flushEng()

	return strings.Join(tokens, " ")
}

// TokenizeForSearch 对搜索查询进行分词
// 英文单词保持原样，中文只用 bigram（去掉 unigram 以提高精确度）
func TokenizeForSearch(query string) string {
	if len(query) == 0 {
		return query
	}

	runes := []rune(query)
	var tokens []string
	var engBuf strings.Builder

	flushEng := func() {
		if engBuf.Len() > 0 {
			tokens = append(tokens, engBuf.String())
			engBuf.Reset()
		}
	}

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if unicode.IsLetter(r) && r < 0x4E00 {
			engBuf.WriteRune(r)
			continue
		}
		if unicode.IsDigit(r) {
			engBuf.WriteRune(r)
			continue
		}
		flushEng()

		if r >= 0x4E00 && r <= 0x9FFF || r >= 0x3400 && r <= 0x4DBF || r >= 0x20000 && r <= 0x2A6DF {
			// 只用 bigram，去掉 unigram 以提高精确匹配
			if i+1 < len(runes) {
				next := runes[i+1]
				if next >= 0x4E00 && next <= 0x9FFF || next >= 0x3400 && next <= 0x4DBF {
					tokens = append(tokens, string(r)+string(next))
				}
			}
		}
	}
	flushEng()

	if len(tokens) == 0 {
		return query
	}
	return strings.Join(tokens, " ")
}
