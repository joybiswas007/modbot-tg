package database

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type ItemModel struct {
	DB *sql.DB
}

// Item represents an item available for purchase in the shop
type Item struct {
	ID          int64
	Name        string
	Type        string // item type "double_points" || "lucky_bonus"
	Description string
	Price       float64 // Price in points
	Duration    int     // Duration in hours (0 if not time-based)
	CreatedAt   time.Time
}

// Get returns the specific item id
func (item ItemModel) Get(itemID int64) (*Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT * FROM shop WHERE id = ?`

	var it Item
	err := item.DB.QueryRowContext(ctx, query, itemID).Scan(
		&it.ID,
		&it.Name,
		&it.Type,
		&it.Description,
		&it.Price,
		&it.Duration,
		&it.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &it, nil
}

// Items returns all the availabl item frm shop
func (item ItemModel) Items() ([]Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := item.DB.QueryContext(ctx, `SELECT * FROM shop`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item

	for rows.Next() {
		var item Item
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Type,
			&item.Description,
			&item.Price,
			&item.Duration,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if len(items) == 0 {
		return nil, nil
	}

	return items, nil
}

func (item ItemModel) Buy(userID, chatID, itemID int64, boostType string, duration int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var expiresAt *time.Time
	if duration > 0 {
		exp := time.Now().Add(time.Duration(duration) * time.Hour)
		expiresAt = &exp
	}

	query := `INSERT INTO boosts(user_id, chat_id, item_id, boost_type, expires_at) VALUES (?, ?, ?, ?, ?)`

	stmt, err := item.DB.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, userID, chatID, itemID, boostType, expiresAt)
	if err != nil {
		return err
	}

	return nil
}
