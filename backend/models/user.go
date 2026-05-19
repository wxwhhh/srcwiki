package models

import (
	"database/sql"
	"litewiki/database"
	"time"
)

// User 用户模型
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // 永不序列化密码
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUser 创建用户
func CreateUser(u *User) (int64, error) {
	result, err := database.DB.Exec(
		"INSERT INTO users (username, password, role, status) VALUES (?, ?, ?, ?)",
		u.Username, u.Password, u.Role, u.Status,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetUserByUsername 根据用户名查找用户
func GetUserByUsername(username string) (*User, error) {
	u := &User{}
	err := database.DB.QueryRow(
		"SELECT id, username, password, role, status, created_at, updated_at FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

// GetUserByID 根据 ID 查找用户
func GetUserByID(id int64) (*User, error) {
	u := &User{}
	err := database.DB.QueryRow(
		"SELECT id, username, password, role, status, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

// ListUsers 分页查询用户列表
func ListUsers(page, size int) ([]User, int64, error) {
	var total int64
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	rows, err := database.DB.Query(
		"SELECT id, username, role, status, created_at, updated_at FROM users ORDER BY id DESC LIMIT ? OFFSET ?",
		size, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

// UpdateUserRole 修改用户角色
func UpdateUserRole(id int64, role string) error {
	_, err := database.DB.Exec(
		"UPDATE users SET role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		role, id,
	)
	return err
}

// UpdateUserStatus 修改用户状态
func UpdateUserStatus(id int64, status string) error {
	_, err := database.DB.Exec(
		"UPDATE users SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, id,
	)
	return err
}

// DeleteUser 删除用户
func DeleteUser(id int64) error {
	_, err := database.DB.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

// UpdatePassword 修改密码
func UpdatePassword(id int64, hashedPassword string) error {
	_, err := database.DB.Exec(
		"UPDATE users SET password = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		hashedPassword, id,
	)
	return err
}
