package services

import (
	"errors"
	"litewiki/models"
	"litewiki/utils"
	"time"

	"github.com/google/uuid"
)

// CreateInviteRequest 创建邀请码请求
type CreateInviteRequest struct {
	Role    string `json:"role" binding:"required"`
	MaxUses int    `json:"max_uses"`
	Expires string `json:"expires"` // ISO 8601 格式，可选
}

// GenerateInviteCode 生成邀请码
func GenerateInviteCode(req *CreateInviteRequest, createdBy int64) (*models.InviteCode, error) {
	if !utils.ValidateInviteRole(req.Role) {
		return nil, errors.New("无效的角色，仅支持 editor 和 reader")
	}

	maxUses := req.MaxUses
	if maxUses < 1 {
		maxUses = 1
	}
	if maxUses > 100 {
		maxUses = 100
	}

	code := &models.InviteCode{
		Code:      uuid.New().String()[:8], // 8位短码
		Role:      req.Role,
		MaxUses:   maxUses,
		UseCount:  0,
		CreatedBy: createdBy,
	}

	if req.Expires != "" {
		t, err := time.Parse(time.RFC3339, req.Expires)
		if err != nil {
			return nil, errors.New("过期时间格式错误，请使用 RFC3339 格式")
		}
		code.ExpiresAt = &t
	}

	id, err := models.CreateInviteCode(code)
	if err != nil {
		return nil, errors.New("创建邀请码失败")
	}
	code.ID = id

	return code, nil
}

// ListInviteCodes 邀请码列表
func ListInviteCodes(page, size int) ([]models.InviteCode, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return models.ListInviteCodes(page, size)
}

// DeleteInviteCode 作废邀请码
func DeleteInviteCode(id int64) error {
	return models.DeleteInviteCode(id)
}

// BatchCreateInviteRequest 批量创建邀请码请求
// expires_in_hours: 可选，过期小时数，0=永不过期
// max_uses: 可选，最大使用次数，0=不限（默认1）
type BatchCreateInviteRequest struct {
	Count          int    `json:"count" binding:"required,min=1,max=500"`
	Role           string `json:"role" binding:"required"`
	ExpiresInHours int    `json:"expires_in_hours"`
	MaxUses        int    `json:"max_uses"`
}

// BatchGenerateInviteCodes 批量生成邀请码
func BatchGenerateInviteCodes(req *BatchCreateInviteRequest, createdBy int64) ([]models.InviteCode, error) {
	if !utils.ValidateInviteRole(req.Role) {
		return nil, errors.New("无效的角色，仅支持 editor 和 reader")
	}

	maxUses := req.MaxUses
	if maxUses < 1 {
		maxUses = 1
	}
	if maxUses > 100 {
		maxUses = 100
	}

	var expiresAt *time.Time
	if req.ExpiresInHours > 0 {
		t := time.Now().Add(time.Duration(req.ExpiresInHours) * time.Hour)
		expiresAt = &t
	}

	codes := make([]models.InviteCode, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		code := &models.InviteCode{
			Code:      uuid.New().String()[:8],
			Role:      req.Role,
			MaxUses:   maxUses,
			UseCount:  0,
			ExpiresAt: expiresAt,
			CreatedBy: createdBy,
		}
		id, err := models.CreateInviteCode(code)
		if err != nil {
			return codes, errors.New("创建邀请码失败")
		}
		code.ID = id
		codes = append(codes, *code)
	}

	return codes, nil
}
