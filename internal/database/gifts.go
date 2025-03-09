package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// GiftModel handles operations related to the Gift table.
type GiftModel struct {
	DB *sql.DB
}

// Gift represents a gift transaction where a user sends points to another user.
type Gift struct {
	ID         int64     `json:"id"`
	ChatID     int64     `json:"chat_id"`
	SenderID   int64     `json:"sender_id"`
	ReceiverID int64     `json:"receiver_id"`
	Amount     float64   `json:"amount"`
	Timestamp  time.Time `json:"timestamp"`
}

// Insert saves a gift transaction in the database.
func (gm GiftModel) Insert(gift *Gift) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO gifts (chat_id, sender_id, receiver_id, amount, timestamp)
		VALUES (?, ?, ?, ?, ?);
	`
	stmt, err := gm.DB.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare gift statement: %v", err)
	}

	defer stmt.Close()

	_, err = gm.DB.ExecContext(ctx, query, gift.ChatID, gift.SenderID, gift.ReceiverID, gift.Amount, gift.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to insert gift: %v", err)
	}
	return nil
}
