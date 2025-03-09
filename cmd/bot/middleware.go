package main

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/spf13/viper"
)

// ensureGroupChat blocks execution if the chat is private or a channel.
func ensureGroupChat(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		chatType := update.Message.Chat.Type
		if chatType == models.ChatTypePrivate || chatType == models.ChatTypeChannel {
			log.Println("Points are only counted in group chats.")
			return
		}
		next(ctx, b, update)
	}
}

// adminMiddleware is a middleware that ensures the user issuing the command is an admin or the owner of the chat.
// If the user is not an admin, they will receive an error message, and the command will not proceed.
// It also handles the case where the command is issued by the "GroupAnonymousBot".
func adminMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		deleteCmd := viper.GetBool("bot.deleteCommand")
		chatID := update.Message.Chat.ID
		userID := update.Message.From.ID

		// Retrieve the list of admins in the chat
		admins, err := getAdmins(ctx, b, chatID)
		if err != nil {
			log.Printf("Failed to retrieve admins: %v\n", err)
			sendMessage(ctx, b, chatID, update.Message.ID, "⚠️ *Error:* Unable to verify admin status. Please try again later.", true, deleteCmd)
			return
		}

		// Check if the user is an admin or the owner
		if !isAdmin(userID, admins) {
			//TODO: We will handle anonymous admins later
			// Special case: Allow GroupAnonymousBot (used for anonymous admins)
			// if update.Message.From.Username == "GroupAnonymousBot" {
			// 	sendMessage(ctx, b, chatID, update.Message.ID, "I'm unable to handle anon users at the moment :(", true, deleteCmd)
			// 	return
			// }

			// Notify the user that they need admin privileges
			sendMessage(ctx, b, chatID, update.Message.ID, "🚫 *Permission Denied:* You must be an admin to use this command.", true, deleteCmd)
			return
		}

		// If the user is an admin, proceed with the next handler
		next(ctx, b, update)
	}
}
