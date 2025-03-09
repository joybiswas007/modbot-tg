package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PointModel struct {
	DB *sql.DB
}

// Point represents a user's earned points in a chat
type Point struct {
	ID        int64     // Unique identifier for the point record (primary key).
	ChatID    int64     // ID of the Telegram group/chat where the points were earned.
	UserID    int64     // Telegram user ID of the participant earning the points.
	Amount    float64   // Number of points awarded for a specific action.
	Source    string    // Source of points "chatting" || "gift" || "double_coins"
	Change    string    // Point was addec or deducted "gain" || "loss"
	TimeStamp time.Time // Timestamp when the points were recorded.
}

// Insert adds a new point record for a user in a chat
func (p PointModel) Insert(point *Point) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO point_history(chat_id, user_id, amount, change, source) VALUES(?, ?, ?, ?, ?)`

	stmt, err := p.DB.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare points statement: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, point.ChatID, point.UserID, point.Amount, point.Change, point.Source)
	if err != nil {
		return fmt.Errorf("failed to execute points statement: %v", err)
	}

	return nil
}

// GetRanking retrieves the top users based on points for a given time period
// (e.g.): "daily" || "weekly" || "monthly"
// limit how many results we are fetching per request
func (p PointModel) Ranking(chatID int64, limit int, period string) ([]Point, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var query string
	switch period {
	case "daily":
		query = `SELECT user_id, source, SUM(amount) AS total_points 
		         FROM point_history 
		         WHERE chat_id = ? AND DATE(timestamp) = DATE('now') 
		         GROUP BY user_id 
		         ORDER BY total_points DESC 
		         LIMIT ?`
	case "weekly":
		query = `SELECT user_id, source, SUM(amount) AS total_points 
		         FROM point_history 
		         WHERE chat_id = ? 
		         AND timestamp >= DATE('now', 'weekday 0', '-6 days') 
		         GROUP BY user_id 
		         ORDER BY total_points DESC 
		         LIMIT ?`
	case "monthly":
		query = `SELECT user_id, source, SUM(amount) AS total_points 
		         FROM point_history 
		         WHERE chat_id = ? 
		         AND strftime('%Y-%m', timestamp) = strftime('%Y-%m', 'now') 
		         GROUP BY user_id 
		         ORDER BY total_points DESC 
		         LIMIT ?`
	}

	rows, err := p.DB.QueryContext(ctx, query, chatID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute ranking query: %v", err)
	}
	defer rows.Close()

	var rankings []Point
	for rows.Next() {
		var p Point
		err := rows.Scan(&p.UserID, &p.Source, &p.Amount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		rankings = append(rankings, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(rankings) == 0 {
		return nil, nil
	}

	// If no results, return an empty slice instead of an error
	return rankings, nil
}

// History retrieves the point history for a specific user in a given chat.
// It fetches the most recent records based on the provided limit.
func (p PointModel) History(chatID, userID int64, limit int) ([]Point, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT chat_id, user_id, amount, change, source, timestamp
		  FROM point_history
		  WHERE chat_id = ? AND user_id = ?
		  ORDER BY timestamp DESC
		  LIMIT ?`
	rows, err := p.DB.QueryContext(ctx, query, chatID, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}

	defer rows.Close()

	var points []Point
	for rows.Next() {
		var p Point
		err := rows.Scan(&p.ChatID, &p.UserID, &p.Amount, &p.Change, &p.Source, &p.TimeStamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		points = append(points, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(points) == 0 {
		return nil, nil
	}

	return points, nil
}
