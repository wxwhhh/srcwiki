// Package services 提供业务逻辑层
package services

import (
	"errors"
	"litewiki/models"
	"litewiki/utils"
	"time"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaID   string `json:"captcha_id" binding:"required"`
	CaptchaCode string `json:"captcha_code" binding:"required"`
}

// RegisterRequest 注册请求
// invite_code 在 open 模式下可选，invite 模式下必填（后端强制校验）
type RegisterRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	InviteCode  string `json:"invite_code"`
	CaptchaID   string `json:"captcha_id" binding:"required"`
	CaptchaCode string `json:"captcha_code" binding:"required"`
}

// Login 用户登录
func Login(req *LoginRequest, jwtSecret string) (string, *models.User, error) {
	// 查找用户
	user, err := models.GetUserByUsername(req.Username)
	if err != nil {
		return "", nil, errors.New("系统错误")
	}
	if user == nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status == "disabled" {
		return "", nil, errors.New("账号已被禁用")
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.Password) {
		return "", nil, errors.New("用户名或密码错误")
	}

	// 生成 JWT
	token, err := utils.GenerateJWT(user.ID, user.Username, user.Role, jwtSecret)
	if err != nil {
		return "", nil, errors.New("生成 Token 失败")
	}

	return token, user, nil
}

// Register 用户注册（支持 open / invite 模式）
func Register(req *RegisterRequest) (*models.User, error) {
	// 校验用户名
	if !utils.ValidateUsername(req.Username) {
		return nil, errors.New("用户名需 3-32 字符，仅允许字母数字下划线")
	}

	// 校验密码
	if !utils.ValidatePassword(req.Password) {
		return nil, errors.New("密码需 6-64 字符")
	}

	// 检查用户名是否已存在
	existing, err := models.GetUserByUsername(req.Username)
	if err != nil {
		return nil, errors.New("系统错误")
	}
	if existing != nil {
		return nil, errors.New("用户名已存在")
	}

	// 读取注册模式
	registerMode := models.GetRegisterMode()

	var role string

	if registerMode == "open" {
		// 开放注册模式：默认角色 reader，无需邀请码
		role = "reader"
	} else {
		// 邀请制模式：必须提供有效邀请码
		if req.InviteCode == "" {
			return nil, errors.New("当前为邀请制注册，请提供有效邀请码")
		}

		// 验证邀请码
		invite, err := models.GetInviteCodeByCode(req.InviteCode)
		if err != nil {
			return nil, errors.New("系统错误")
		}
		if invite == nil {
			return nil, errors.New("邀请码无效")
		}

		// 检查邀请码是否过期
		if invite.ExpiresAt != nil && invite.ExpiresAt.Before(time.Now()) {
			return nil, errors.New("邀请码已过期")
		}

		// 检查邀请码是否用完
		if invite.UseCount >= invite.MaxUses {
			return nil, errors.New("邀请码已用完")
		}

		role = invite.Role

		// 使用邀请码
		_ = models.UseInviteCode(invite.ID)
	}

	// 哈希密码
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("系统错误")
	}

	// 创建用户
	user := &models.User{
		Username: req.Username,
		Password: hash,
		Role:     role,
		Status:   "active",
	}
	userID, err := models.CreateUser(user)
	if err != nil {
		return nil, errors.New("创建用户失败")
	}
	user.ID = userID

	return user, nil
}



// ChangePassword 修改密码
func ChangePassword(userID int64, oldPassword, newPassword string) error {
	user, err := models.GetUserByID(userID)
	if err != nil || user == nil {
		return errors.New("用户不存在")
	}

	if !utils.CheckPassword(oldPassword, user.Password) {
		return errors.New("原密码错误")
	}

	if !utils.ValidatePassword(newPassword) {
		return errors.New("新密码需 6-64 字符")
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("系统错误")
	}

	return models.UpdatePassword(userID, hash)
}
