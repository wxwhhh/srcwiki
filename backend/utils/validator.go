package utils

import "unicode/utf8"

// ValidateUsername 校验用户名：3-32 字符，仅允许字母数字下划线
func ValidateUsername(username string) bool {
	length := utf8.RuneCountInString(username)
	if length < 3 || length > 32 {
		return false
	}
	for _, r := range username {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}

// ValidatePassword 校验密码：6-64 字符
func ValidatePassword(password string) bool {
	length := utf8.RuneCountInString(password)
	return length >= 6 && length <= 64
}

// ValidateRole 校验角色是否合法
func ValidateRole(role string) bool {
	return role == "admin" || role == "editor" || role == "reader"
}

// ValidateInviteRole 校验邀请码角色（不含 admin）
func ValidateInviteRole(role string) bool {
	return role == "editor" || role == "reader"
}
