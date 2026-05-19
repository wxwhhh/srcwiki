package models

import (
	"database/sql"
	"litewiki/database"
	"time"
)

// InviteCode 邀请码模型
type InviteCode struct {
	ID        int64      `json:"id"`
	Code      string     `json:"code"`
	Role      string     `json:"role"`
	MaxUses   int        `json:"max_uses"`
	UseCount  int        `json:"use_count"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedBy int64      `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
}

// CreateInviteCode 创建邀请码
func CreateInviteCode(ic *InviteCode) (int64, error) {
	result, err := database.DB.Exec(
		"INSERT INTO invite_codes (code, role, max_uses, use_count, expires_at, created_by) VALUES (?, ?, ?, 0, ?, ?)",
		ic.Code, ic.Role, ic.MaxUses, ic.ExpiresAt, ic.CreatedBy,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetInviteCodeByCode 根据邀请码字符串查找
func GetInviteCodeByCode(code string) (*InviteCode, error) {
	ic := &InviteCode{}
	err := database.DB.QueryRow(
		"SELECT id, code, role, max_uses, use_count, expires_at, created_by, created_at FROM invite_codes WHERE code = ?",
		code,
	).Scan(&ic.ID, &ic.Code, &ic.Role, &ic.MaxUses, &ic.UseCount, &ic.ExpiresAt, &ic.CreatedBy, &ic.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return ic, err
}

// UseInviteCode 使用邀请码（原子递增 use_count）
func UseInviteCode(id int64) error {
	_, err := database.DB.Exec(
		"UPDATE invite_codes SET use_count = use_count + 1 WHERE id = ?",
		id,
	)
	return err
}

// ListInviteCodes 分页查询邀请码列表
func ListInviteCodes(page, size int) ([]InviteCode, int64, error) {
	var total int64
	err := database.DB.QueryRow("SELECT COUNT(*) FROM invite_codes").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	rows, err := database.DB.Query(
		"SELECT id, code, role, max_uses, use_count, expires_at, created_by, created_at FROM invite_codes ORDER BY id DESC LIMIT ? OFFSET ?",
		size, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var codes []InviteCode
	for rows.Next() {
		var ic InviteCode
		if err := rows.Scan(&ic.ID, &ic.Code, &ic.Role, &ic.MaxUses, &ic.UseCount, &ic.ExpiresAt, &ic.CreatedBy, &ic.CreatedAt); err != nil {
			return nil, 0, err
		}
		codes = append(codes, ic)
	}
	return codes, total, nil
}

// DeleteInviteCode 删除邀请码
func DeleteInviteCode(id int64) error {
	_, err := database.DB.Exec("DELETE FROM invite_codes WHERE id = ?", id)
	return err
}
