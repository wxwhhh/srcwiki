package services

import (
	"errors"
	"litewiki/models"
	"litewiki/utils"
)

// ListUsers 用户列表
func ListUsers(page, size int) ([]models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return models.ListUsers(page, size)
}

// UpdateUserRole 修改用户角色
func UpdateUserRole(id int64, role string) error {
	if !utils.ValidateRole(role) {
		return errors.New("无效的角色")
	}
	return models.UpdateUserRole(id, role)
}

// DeleteUser 删除用户（不能删除自己）
func DeleteUser(id int64, operatorID int64) error {
	if id == operatorID {
		return errors.New("不能删除自己的账号")
	}
	user, err := models.GetUserByID(id)
	if err != nil || user == nil {
		return errors.New("用户不存在")
	}
	return models.DeleteUser(id)
}

// UpdateUserStatus 修改用户状态
func UpdateUserStatus(id int64, status string) error {
	if status != "active" && status != "disabled" {
		return errors.New("无效的状态")
	}
	return models.UpdateUserStatus(id, status)
}
