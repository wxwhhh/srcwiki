package models

import (
	"litewiki/database"
	"time"
)

// AuditLog 审计日志模型
type AuditLog struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Username   string    `json:"username"`
	Action     string    `json:"action"`
	TargetType string    `json:"target_type"`
	TargetID   int64     `json:"target_id"`
	Detail     string    `json:"detail"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
}

// LoginStats 登录统计结果
type LoginStats struct {
	TotalLogins  int64              `json:"total_logins"`
	TodayLogins  int64              `json:"today_logins"`
	ActiveUsers  int64              `json:"active_users"`
	LoginsByUser []UserLoginStat    `json:"logins_by_user"`
	RecentLogins []AuditLog         `json:"recent_logins"`
}

// UserLoginStat 用户登录统计
type UserLoginStat struct {
	Username  string    `json:"username"`
	Count     int64     `json:"count"`
	LastIP    string    `json:"last_ip"`
	LastLogin time.Time `json:"last_login"`
}

// GetLoginStats 获取登录统计
func GetLoginStats() (*LoginStats, error) {
	stats := &LoginStats{}

	// 总登录次数
	err := database.DB.QueryRow("SELECT COUNT(*) FROM audit_log WHERE action = 'login'").Scan(&stats.TotalLogins)
	if err != nil {
		return nil, err
	}

	// 今日登录次数
	err = database.DB.QueryRow("SELECT COUNT(*) FROM audit_log WHERE action = 'login' AND DATE(created_at) = DATE('now')").Scan(&stats.TodayLogins)
	if err != nil {
		return nil, err
	}

	// 活跃用户数
	err = database.DB.QueryRow("SELECT COUNT(DISTINCT username) FROM audit_log WHERE action = 'login'").Scan(&stats.ActiveUsers)
	if err != nil {
		return nil, err
	}

	// 按用户分组统计
	rows, err := database.DB.Query(`
		SELECT username, COUNT(*) as cnt, COALESCE(MAX(ip), '') as last_ip, COALESCE(MAX(created_at), '') as last_login
		FROM audit_log WHERE action = 'login'
		GROUP BY username ORDER BY cnt DESC LIMIT 50
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var u UserLoginStat
		var lastLoginStr string
		if err := rows.Scan(&u.Username, &u.Count, &u.LastIP, &lastLoginStr); err != nil {
			return nil, err
		}
		if t, err := time.Parse("2006-01-02 15:04:05", lastLoginStr); err == nil {
			u.LastLogin = t
		}
		stats.LoginsByUser = append(stats.LoginsByUser, u)
	}

	// 最近 20 条登录记录
	rows2, err := database.DB.Query(`
		SELECT id, username, COALESCE(ip, ''), COALESCE(user_agent, ''), created_at
		FROM audit_log WHERE action = 'login'
		ORDER BY created_at DESC LIMIT 20
	`)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var l AuditLog
		var createdAtStr string
		if err := rows2.Scan(&l.ID, &l.Username, &l.IP, &l.UserAgent, &createdAtStr); err != nil {
			return nil, err
		}
		if t, err := time.Parse("2006-01-02 15:04:05", createdAtStr); err == nil {
			l.CreatedAt = t
		}
		stats.RecentLogins = append(stats.RecentLogins, l)
	}

	return stats, nil
}

// InsertAuditLog 插入审计日志
func InsertAuditLog(log *AuditLog) {
	database.DB.Exec(
		`INSERT INTO audit_log (user_id, username, action, target_type, target_id, detail, ip, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		log.UserID, log.Username, log.Action, log.TargetType, log.TargetID, log.Detail, log.IP, log.UserAgent,
	)
}

// ListAuditLogs 分页查询审计日志
func ListAuditLogs(page, size int, userID int64, action string) ([]AuditLog, int64, error) {
	where := "WHERE 1=1"
	args := []interface{}{}

	if userID > 0 {
		where += " AND user_id = ?"
		args = append(args, userID)
	}
	if action != "" {
		where += " AND action = ?"
		args = append(args, action)
	}

	var total int64
	err := database.DB.QueryRow("SELECT COUNT(*) FROM audit_log "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	queryArgs := append(args, size, offset)
	rows, err := database.DB.Query(
		"SELECT id, user_id, username, action, target_type, target_id, detail, ip, user_agent, created_at FROM audit_log "+where+" ORDER BY created_at DESC LIMIT ? OFFSET ?",
		queryArgs...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Username, &l.Action, &l.TargetType, &l.TargetID, &l.Detail, &l.IP, &l.UserAgent, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	return logs, total, nil
}
