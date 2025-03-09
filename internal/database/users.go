package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type UserModel struct {
	DB *sql.DB
}

// User represents a participant in a Telegram chat with tracked points.
type User struct {
	ID        int64     // Unique identifier for the user record (primary key).
	UserID    int64     // Telegram user ID of the participant.
	ChatID    int64     // ID of the Telegram group/chat where the user is active.
	Points    float64   // Points earned by the user for their activity in the chat.
	CreatedAt time.Time // Timestamp when the user entry was first created.
	UpdatedAt time.Time // Timestamp when the user entry was last updated.
}

// Boost represents a temporary boost effect for a user, such as double coins.
type Boost struct {
	ID        int64     `json:"id"`         // Unique boost ID
	UserID    int64     `json:"user_id"`    // ID of the user receiving the boost
	ChatID    int64     `json:"chat_id"`    // Chat ID where the boost is applied
	Type      string    `json:"boost_type"` // Type of boost (e.g., "double_coins")
	ExpiresAt time.Time `json:"expires_at"` // Expiration timestamp of the boost
}

// Get retrieves a user’s data from the database for a specific chat.
func (u UserModel) Get(chatID, userID int64) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := "SELECT * FROM users WHERE chat_id = ? AND user_id = ?"

	var user User
	err := u.DB.QueryRowContext(ctx, query, chatID, userID).Scan(
		&user.ID,
		&user.UserID,
		&user.ChatID,
		&user.Points,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

// Insert adds a new user entry into the database.
func (u UserModel) Insert(chatID, userID int64, point float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO users(user_id, chat_id, points, created_at, updated_at) VALUES(?, ?, ?, ?, ?)`

	stmt, err := u.DB.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, userID, chatID, point, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert user: %v", err)
	}

	return nil
}

// Update modifies the points of an existing user in the database.
func (u UserModel) Update(chatID, userID int64, point float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE users SET points = ?, updated_at = ? WHERE chat_id = ? AND user_id = ?`

	_, err := u.DB.ExecContext(ctx, query, point, time.Now(), chatID, userID)
	if err != nil {
		return err
	}
	return nil
}

// Leaderboard fetches the top N users with the highest points in a specific chat.
func (u UserModel) Leaderboard(chatID int64, topN int) ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT user_id, points, updated_at FROM users WHERE chat_id = ? ORDER BY points DESC LIMIT ?`

	rows, err := u.DB.QueryContext(ctx, query, chatID, topN)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var toppers []User

	for rows.Next() {
		var topper User

		err := rows.Scan(
			&topper.UserID,
			&topper.Points,
			&topper.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}
		toppers = append(toppers, topper)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(toppers) == 0 {
		return nil, nil
	}

	return toppers, nil
}

func (u UserModel) ActiveBoost(userID, chatID int64) (*Boost, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, user_id, chat_id, boost_type, expires_at FROM boosts 
	          WHERE user_id = ? AND chat_id = ? AND expires_at > CURRENT_TIMESTAMP LIMIT 1`

	var boost Boost
	err := u.DB.QueryRowContext(ctx, query, userID, chatID).Scan(&boost.ID, &boost.UserID, &boost.ChatID, &boost.Type, &boost.ExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No active boost
		}
		return nil, err
	}

	return &boost, nil
}
