package models

import (
	"database/sql"
	"litewiki/database"
	"time"
)

// Credit 致谢模型
type Credit struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	IconURL     string    `json:"icon_url"`
	License     string    `json:"license"`
	Stars       string    `json:"stars"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateCredit 创建致谢条目
func CreateCredit(c *Credit) (int64, error) {
	result, err := database.DB.Exec(
		"INSERT INTO credits (name, url, description, icon_url, license, stars, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?)",
		c.Name, c.URL, c.Description, c.IconURL, c.License, c.Stars, c.SortOrder,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetAllCredits 获取所有致谢条目（按 sort_order 排序）
func GetAllCredits() ([]Credit, error) {
	rows, err := database.DB.Query(
		"SELECT id, name, url, description, icon_url, license, stars, sort_order, created_at, updated_at FROM credits ORDER BY sort_order ASC, id ASC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credits []Credit
	for rows.Next() {
		var c Credit
		if err := rows.Scan(&c.ID, &c.Name, &c.URL, &c.Description, &c.IconURL, &c.License, &c.Stars, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		credits = append(credits, c)
	}
	return credits, nil
}

// GetCreditByID 根据 ID 获取致谢条目
func GetCreditByID(id int64) (*Credit, error) {
	c := &Credit{}
	err := database.DB.QueryRow(
		"SELECT id, name, url, description, icon_url, license, stars, sort_order, created_at, updated_at FROM credits WHERE id = ?",
		id,
	).Scan(&c.ID, &c.Name, &c.URL, &c.Description, &c.IconURL, &c.License, &c.Stars, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

// UpdateCredit 更新致谢条目
func UpdateCredit(c *Credit) error {
	_, err := database.DB.Exec(
		"UPDATE credits SET name = ?, url = ?, description = ?, icon_url = ?, license = ?, stars = ?, sort_order = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		c.Name, c.URL, c.Description, c.IconURL, c.License, c.Stars, c.SortOrder, c.ID,
	)
	return err
}

// DeleteCredit 删除致谢条目
func DeleteCredit(id int64) error {
	_, err := database.DB.Exec("DELETE FROM credits WHERE id = ?", id)
	return err
}
