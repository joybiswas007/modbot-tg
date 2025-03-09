package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joybiswas007/modbot-tg/internal/database"
)

// sendMessage is a helper function for sending messages in a Telegram chat.
//
// Parameters:
// - ctx: The context for the request.
// - b: The bot instance used to send messages.
// - chatID: The chat where the message will be sent.
// - messageID: The ID of the message to reply to (if reply is true).
// - text: The content of the message to be sent.
// - reply: If true, the message will be sent as a reply to messageID.
// - delete: If true, the original message (messageID) will be deleted after sending.
//
// This function allows flexibility in message handling by providing options
// to reply to a specific message and delete the user's original command/message.
func sendMessage(ctx context.Context, b *bot.Bot, chatID int64, messageID int, text string, reply, delete bool) {
	var msgParams *bot.SendMessageParams

	msgParams = &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeMarkdownV1,
	}

	// If reply is enabled, attach reply parameters
	if reply {
		msgParams.ReplyParameters = &models.ReplyParameters{
			MessageID: messageID,
		}
	}

	// Send the message
	b.SendMessage(ctx, msgParams)

	// If delete is enabled, remove the original message
	if delete {
		b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    chatID,
			MessageID: messageID,
		})
	}
}

// randRange generates a random integer between min (inclusive) and max (exclusive).
func randRange(min, max int) int {
	return rand.IntN(max-min) + min
}

// displayName determines the appropriate display name for a user.
// It prioritizes using both the first name and username, but falls back gracefully.
func displayName(firstName, username string) string {
	switch {
	case username != "" && firstName != "":
		return fmt.Sprintf("%s (@%s)", firstName, username) // FirstName (@username)
	case username != "":
		return fmt.Sprintf("@%s", username) // Only @username
	case firstName != "":
		return firstName // Only FirstName
	default:
		return "User" // Fallback
	}
}

// formatUserDetails formats user details using Markdown
func formatUserDetails(firstName, username string, user *database.User) string {
	var sb strings.Builder

	// Format the message
	sb.WriteString(fmt.Sprintf("*Stats of %s*\n", displayName(firstName, username)))
	sb.WriteString(fmt.Sprintf("- *Total Points:* `%.2f`\n", user.Points))
	sb.WriteString(fmt.Sprintf("- *Last Activity:* `%s`\n", user.UpdatedAt.Format("2006-01-02 15:04:05")))

	return sb.String()
}

// Function to calculate points based on message type
func calculatePoints(msg *models.Message) float64 {
	var point float64

	switch {
	case msg.Text != "":
		point = pointMap["text"]()
	case msg.Document != nil:
		point = pointMap["document"]()
	case msg.Photo != nil:
		point = pointMap["photo"]()
	case msg.Sticker != nil:
		point = pointMap["sticker"]()
	case msg.Audio != nil:
		point = pointMap["audio"]()
	case msg.Animation != nil:
		point = pointMap["animation"]()
	}

	return point
}

// formatRankingMessage generates a nicely formatted Markdown message for the leaderboard.
// It includes a title, a numbered list of ranked users, and a fallback message if there are no rankings.
// The function uses Telegram's user mention format (tg://user?id=user_id) for clickable usernames.
func formatRankingMessage(title string, points []database.Point) string {
	var sb strings.Builder
	sb.WriteString(title + "\n\n")

	for i, p := range points {
		sb.WriteString(fmt.Sprintf("%d. [%d](tg://user?id=%d) - *%.2f points*\n", i+1, p.UserID, p.UserID, p.Amount))
	}

	return sb.String()
}

// formatHistory formats the point history for a user in Markdown mode.
// It includes the user's name, username, points earned/lost, reasons, and timestamps.
func formatHistory(chatUser *models.User, history []database.Point, limit int) string {
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("*Last %d point history for %s:*\n\n", limit, displayName(chatUser.FirstName, chatUser.Username)))

	// Iterate through history and append formatted records
	for _, entry := range history {
		timeFormatted := entry.TimeStamp.Format("2006-01-02 15:04:05") // Format timestamp
		symbol := "➕"                                                  // Default for gained points

		if entry.Change == "loss" {
			symbol = "➖" // Change symbol if points were lost
		}

		msg.WriteString(fmt.Sprintf("🕒 *%s* - %s *%.2f points* (%s)\n", timeFormatted, symbol, entry.Amount, entry.Source))
	}

	return msg.String()
}

// getAdmins retrieves the list of administrators for a given chat.
func getAdmins(ctx context.Context, b *bot.Bot, chatID int64) ([]models.ChatMember, error) {
	// Request the list of chat administrators from the Telegram bot API
	admins, err := b.GetChatAdministrators(ctx, &bot.GetChatAdministratorsParams{
		ChatID: chatID,
	})
	if err != nil {
		return nil, err
	}
	// Return the list of admins
	return admins, nil
}

// isAdmin checks if a given user ID belongs to an admin in the provided list of chat members.
func isAdmin(userID int64, admins []models.ChatMember) bool {
	for _, admin := range admins {
		// Check if the user is the chat owner
		if admin.Owner != nil && admin.Owner.User.ID == userID {
			return true
			// Check if the user is an administrator
		} else if admin.Administrator != nil && admin.Administrator.User.ID == userID {
			return true
		}
	}
	return false // Return false if the user is not found in the admin list
}

// checkBotPermission checks if the bot has administrator privileges in the chat.
// It searches for the bot's ID in the list of chat administrators.
func checkBotPermission(botID int64, admins []models.ChatMember) *models.ChatMemberAdministrator {
	// Iterate through the list of administrators
	for _, bot := range admins {
		// Check if the current admin entry belongs to the bot
		if bot.Administrator != nil && bot.Administrator.User.ID == botID {
			// Return the bot's admin details if found
			return bot.Administrator

		}
	}
	// Return an nil if the bot is not found in the admin list
	return nil
}
