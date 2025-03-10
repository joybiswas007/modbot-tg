package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joybiswas007/modbot-tg/internal/database"
	"github.com/spf13/viper"
)

var (
	// Define a map to store point values for each message type
	pointMap = map[string]func() float64{
		"text": func() float64 {
			min := viper.GetInt("bot.point.text.min")
			max := viper.GetInt("bot.point.text.max")
			return float64(randRange(min, max))
		},
		"docment":   func() float64 { return viper.GetFloat64("bot.point.document") },
		"photo":     func() float64 { return viper.GetFloat64("bot.point.photo") },
		"sticker":   func() float64 { return viper.GetFloat64("bot.point.sticker") },
		"audio":     func() float64 { return viper.GetFloat64("bot.point.audio") },
		"animation": func() float64 { return viper.GetFloat64("bot.point.animation") },
	}

	// pointSources defines different ways users can earn points.
	pointSources = map[int]string{
		1: "chatting",
		2: "doublePoints",
		3: "gift",
		4: "luckyBonus",
		5: "boughtBoost",
		6: "penalty",
	}
)

func (app *application) start(ctx context.Context, b *bot.Bot, update *models.Update) {
	me, err := b.GetMe(ctx)
	if err != nil {
		log.Println("Failed to get bot details:", err)
		return
	}

	// deleteCmd determines if commands should be deleted after execution
	deleteCmd := viper.GetBool("bot.deleteCommand")

	msgId := update.Message.ID
	chatID := update.Message.Chat.ID

	startMessage := fmt.Sprintf(
		"Hello! I'm *%s*.\n\n"+
			"I help track user activity, assign points, and generate rankings based on participation in this chat.\n\n"+
			"Use /help to see all available commands.\n\n"+
			"⚠️ *To start counting messages and assigning points, make sure to promote me as an admin.*",
		me.FirstName,
	)

	sendMessage(ctx, b, chatID, msgId, startMessage, true, deleteCmd)
}

// getID sends the chat and user id
func (app *application) getID(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	msgId := update.Message.ID
	var userID int64

	// deleteCmd determines if commands should be deleted after execution
	deleteCmd := viper.GetBool("bot.deleteCommand")

	if update.Message.ReplyToMessage != nil {
		userID = update.Message.ReplyToMessage.From.ID
	} else {
		userID = update.Message.From.ID
	}

	message := fmt.Sprintf("**Chat ID:** `%d`\n**User ID:** `%d`", chatID, userID)

	sendMessage(ctx, b, chatID, msgId, message, true, deleteCmd)
}

func (app *application) help(ctx context.Context, b *bot.Bot, update *models.Update) {
	deleteCmd := viper.GetBool("bot.deleteCommand")
	helpMessage := `*Available Commands:*

*/rank [daily|weekly|monthly]* - Show ranking based on activity.
*/history* - Show your last 50 activity records.
*/stats* - Display your overall stats in the chat.
*/gift [userid amount]* - gift points to users.
*/gift [amount]* - reply to user(s) message.
*/shop* - display all the items available in shop.
*/boost* - display users avilable boost.
*/buy [itemid]* - buy any item specified by item id.
*/id* - Display your user ID and the chat ID.
*/help* - Display this help message.

_Use these commands to track your activity and rankings!_`

	chatID := update.Message.Chat.ID
	msgId := update.Message.ID

	sendMessage(ctx, b, chatID, msgId, helpMessage, true, deleteCmd)
}

// countMessage handles the logic for counting points based on message type
func (app *application) countMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := update.Message
	userID := msg.From.ID
	chatID := msg.Chat.ID

	// deleteCmd determines if commands should be deleted after execution
	deleteCmd := viper.GetBool("bot.deleteCommand")

	newMembers := update.Message.NewChatMembers

	// Bot will automatically send a message after adding to a group
	if len(newMembers) != 0 {
		me, err := b.GetMe(ctx)
		if err != nil {
			sendMessage(ctx, b, chatID, msg.ID, ErrUnknown, true, deleteCmd)
			return
		}
		for _, user := range newMembers {
			if me.Username == user.Username {
				// Found the bot in the new members list
				meMsg := fmt.Sprintf("Hello! I'm %s. Thanks for adding me to this group. Use `/help` to see what I can do!\n\nTo unlock my full potential, make me an admin.", me.FirstName)

				sendMessage(ctx, b, chatID, msg.ID, meMsg, false, deleteCmd)
				return
			}
		}
		return
	}

	point := calculatePoints(msg)

	p := &database.Point{
		ChatID: chatID,
		UserID: userID,
		Source: pointSources[1],
		Change: "gain",
	}

	user, err := app.models.Users.Get(chatID, userID)
	if err != nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrUnknown, true, deleteCmd)
		return
	}

	// If user doesn't exist, insert the point and user data for the first time
	if user == nil {
		err = app.models.Users.Insert(chatID, userID, point)
		if err != nil {
			sendMessage(ctx, b, chatID, msg.ID, ErrUnknown, true, deleteCmd)
			return
		}

		p.Amount = point
		p.Source = pointSources[1]

		err = app.models.Points.Insert(p)
		if err != nil {
			sendMessage(ctx, b, chatID, msg.ID, ErrUnknown, true, deleteCmd)
			return
		}
		return
	}

	boost, err := app.models.Users.ActiveBoost(userID, chatID)
	if err != nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrUnknown, true, deleteCmd)
		return
	}

	if boost != nil {
		switch boost.Type {
		case "doublePoints":
			point *= 2
			p.Source = pointSources[2]
		case "luckyBonus":
			bonus := calculateLuckyBonus(point)
			point += bonus
			p.Source = pointSources[4]
		}
	}

	err = app.models.Users.Update(chatID, userID, user.Points+point)
	if err != nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrUnknown, true, deleteCmd)
		return
	}

	p.Amount = point

	err = app.models.Points.Insert(p)
	if err != nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrUnknown, true, deleteCmd)
		return
	}
}

// userStats retrieves and displays user statistics
func (app *application) userStats(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	msgId := update.Message.ID
	var userID int64
	var chatUser *models.User

	// deleteCmd determines if commands should be deleted after execution
	deleteCmd := viper.GetBool("bot.deleteCommand")

	if update.Message.ReplyToMessage != nil {
		userID = update.Message.ReplyToMessage.From.ID
		chatUser = update.Message.ReplyToMessage.From
	} else {
		userID = update.Message.From.ID
		chatUser = update.Message.From
	}

	user, err := app.models.Users.Get(chatID, userID)
	if err != nil {
		sendMessage(ctx, b, chatID, msgId, ErrNoStatsFound, true, deleteCmd)
		return
	}

	if user == nil {
		sendMessage(ctx, b, chatID, msgId, ErrNoStatsFound, true, deleteCmd)
		return
	}

	// Send the nicely formatted message
	msg := formatUserDetails(chatUser.FirstName, chatUser.Username, user)
	sendMessage(ctx, b, chatID, msgId, msg, true, deleteCmd)
}

// topUsers retrieves and displays the top users based on the ranking type
func (app *application) topUsers(ctx context.Context, b *bot.Bot, update *models.Update) {
	// Trim and clean the input after "/rank"
	rankType := strings.ToLower(strings.TrimSpace(strings.Replace(update.Message.Text, "/rank", "", 1)))

	validTypes := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}

	chatID := update.Message.Chat.ID
	msgId := update.Message.ID

	// deleteCmd determines if commands should be deleted after execution
	deleteCmd := viper.GetBool("bot.deleteCommand")

	if _, exists := validTypes[rankType]; !exists {
		msg := "Use `/rank daily`, `/rank weekly`, or `/rank monthly`."
		sendMessage(ctx, b, chatID, msgId, msg, true, deleteCmd)
		return
	}

	rankingLimit := 20

	points, err := app.models.Points.Ranking(chatID, rankingLimit, rankType)
	if err != nil {
		sendMessage(ctx, b, chatID, msgId, ErrNoRankingsYet, true, deleteCmd)
		return
	}
	if points == nil {
		sendMessage(ctx, b, chatID, msgId, ErrNoRankingsYet, true, deleteCmd)
		return
	}

	// Handle ranking logic based on rankType
	var msg string
	switch rankType {
	case "daily":
		msg = formatRankingMessage("🏆 *Daily Rankings*", points)
	case "weekly":
		msg = formatRankingMessage("🌟 *Weekly Rankings*", points)
	case "monthly":
		msg = formatRankingMessage("🎖 *Monthly Rankings*", points)
	}

	sendMessage(ctx, b, chatID, msgId, msg, true, deleteCmd)
}

// pointHistory retrieves and displays the point history for a user
func (app *application) pointHistory(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	msgId := update.Message.ID
	var userID int64
	var chatUser *models.User

	if update.Message.ReplyToMessage != nil {
		userID = update.Message.ReplyToMessage.From.ID
		chatUser = update.Message.ReplyToMessage.From
	} else {
		userID = update.Message.From.ID
		chatUser = update.Message.From
	}

	historyLimit := 50

	// deleteCmd determines if commands should be deleted after execution
	deleteCmd := viper.GetBool("bot.deleteCommand")

	history, err := app.models.Points.History(chatID, userID, historyLimit)
	if err != nil {
		sendMessage(ctx, b, chatID, msgId, ErrNoHistoryFound, true, deleteCmd)
		return
	}
	if history == nil {
		sendMessage(ctx, b, chatID, msgId, ErrNoHistoryFound, true, deleteCmd)
		return
	}

	historyMsg := formatHistory(chatUser, history, historyLimit)
	sendMessage(ctx, b, chatID, msgId, historyMsg, true, deleteCmd)
}

// gift handles the logic for gifting points to another user
func (app *application) gift(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	msgId := update.Message.ID
	var receiverID int64
	var giftAmount int

	gift := strings.TrimSpace(strings.Replace(update.Message.Text, "/gift", "", 1))
	parts := strings.Fields(gift)

	// deleteCmd determines if commands should be deleted after execution
	deleteCmd := viper.GetBool("bot.deleteCommand")

	if update.Message.ReplyToMessage != nil {
		// If user replied to a message, extract the user ID from the replied message
		receiverID = update.Message.ReplyToMessage.From.ID

		// Ensure the user provided a gift amount
		if len(parts) != 1 {
			msg := "Usage: Reply to a message with `/gift amount`."
			sendMessage(ctx, b, chatID, msgId, msg, true, deleteCmd)
			return
		}

		// Parse the gift amount
		parsedAmount, err := strconv.Atoi(parts[0])
		if err != nil || parsedAmount <= 0 {
			sendMessage(ctx, b, chatID, msgId, ErrInvalidGiftAmount, true, deleteCmd)
			return
		}
		giftAmount = parsedAmount

	} else {
		if len(parts) != 2 {
			msg := "Usage: `/gift user_id amount`."
			sendMessage(ctx, b, chatID, msgId, msg, true, deleteCmd)
			return
		}

		// Convert userID to int64
		parsedUserID, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			sendMessage(ctx, b, chatID, msgId, ErrInvalidID, true, deleteCmd)
			return
		}
		receiverID = parsedUserID

		// Convert gift amount to int
		parsedAmount, err := strconv.Atoi(parts[1])
		if err != nil || parsedAmount <= 0 {
			sendMessage(ctx, b, chatID, msgId, ErrInvalidGiftAmount, true, deleteCmd)
			return
		}
		giftAmount = parsedAmount
	}

	if receiverID == update.Message.From.ID {
		sendMessage(ctx, b, chatID, msgId, ErrGiftToSelf, true, deleteCmd)
		return
	}

	me, err := b.GetMe(ctx)
	if err != nil {
		sendMessage(ctx, b, chatID, msgId, ErrUserDoesNotExist, true, deleteCmd)
		return
	}

	if receiverID == me.ID {
		sendMessage(ctx, b, chatID, msgId, ErrGiftToBot, true, deleteCmd)
		return
	}

	// Ensure valid gift amount
	if giftAmount <= 0 {
		sendMessage(ctx, b, chatID, msgId, ErrGiftAmountZero, true, deleteCmd)
		return
	}

	// User who is giving the gift
	issuer, err := app.models.Users.Get(chatID, update.Message.From.ID)
	if err != nil {
		sendMessage(ctx, b, chatID, msgId, ErrNoPointsYet, true, deleteCmd)
		return
	}

	// Check if gift issuer is valid or not
	if issuer == nil {
		sendMessage(ctx, b, chatID, msgId, ErrNoPointsYet, true, deleteCmd)
		return
	}

	// Check if issuer has enough points
	if issuer.Points == 0 || issuer.Points < float64(giftAmount) {
		sendMessage(ctx, b, chatID, msgId, ErrNoPointsToGift, true, deleteCmd)
		return
	}

	// User who is going to receive the gift
	receiver, err := app.models.Users.Get(chatID, receiverID)
	if err != nil {
		sendMessage(ctx, b, chatID, msgId, ErrUserDoesNotExist, true, deleteCmd)
		return
	}

	// Check if receiver exists
	if receiver == nil {
		sendMessage(ctx, b, chatID, msgId, ErrUserDoesNotExist, true, deleteCmd)
		return
	}

	p := &database.Point{
		ChatID: chatID,
		Amount: float64(giftAmount),
		Source: pointSources[3],
		Change: "loss",
	}

	// Update total points of issuer
	err = app.models.Users.Update(chatID, update.Message.From.ID, issuer.Points-float64(giftAmount))
	if err != nil {
		sendMessage(ctx, b, chatID, msgId, ErrGiftProcessingFail, true, deleteCmd)
		return
	}
	p.UserID = update.Message.From.ID

	err = app.models.Points.Insert(p)
	if err != nil {
		sendMessage(ctx, b, chatID, msgId, ErrGiftProcessingFail, true, deleteCmd)
		return
	}

	// Update total points of receiver
	err = app.models.Users.Update(chatID, receiverID, receiver.Points+float64(giftAmount))
	if err != nil {
		sendMessage(ctx, b, chatID, msgId, ErrGiftProcessingFail, true, deleteCmd)
		return
	}

	p.UserID = receiverID
	p.Change = "gain"

	// Add gift points to the user
	err = app.models.Points.Insert(p)
	if err != nil {
		sendMessage(ctx, b, chatID, msgId, ErrGiftProcessingFail, true, deleteCmd)
		return
	}

	gft := &database.Gift{
		ChatID:     chatID,
		SenderID:   issuer.UserID,
		ReceiverID: receiver.UserID,
		Amount:     float64(giftAmount),
		Timestamp:  time.Now(),
	}

	err = app.models.Gifts.Insert(gft)
	if err != nil {
		sendMessage(ctx, b, chatID, msgId, ErrGiftProcessingFail, true, deleteCmd)
		return
	}

	// Send confirmation message
	msg := fmt.Sprintf("🎁 %d points have been gifted to user %d!", giftAmount, receiverID)
	sendMessage(ctx, b, chatID, msgId, msg, true, deleteCmd)
}

// shop display available items
func (app *application) shop(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := update.Message
	chatID := msg.Chat.ID

	// deleteCmd determines if commands should be deleted after execution
	deleteCmd := viper.GetBool("bot.deleteCommand")

	items, err := app.models.Shop.Items()
	if err != nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrUnknown, true, deleteCmd)
		return
	}

	if items == nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrNoItem, true, deleteCmd)
		return
	}

	sendMessage(ctx, b, chatID, msg.ID, formatShopItems(items), true, deleteCmd)
}

// buy any item specified by item id
func (app *application) buyItem(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := update.Message
	chatID := msg.Chat.ID
	userID := msg.From.ID

	// deleteCmd determines if commands should be deleted after execution
	deleteCmd := viper.GetBool("bot.deleteCommand")

	item := strings.TrimSpace(strings.Replace(update.Message.Text, "/buy", "", 1))
	parts := strings.Fields(item)

	if len(parts) != 1 {
		message := "Usage: `/buy item_id`."
		sendMessage(ctx, b, chatID, msg.ID, message, true, deleteCmd)
		return
	}

	buyer, err := app.models.Users.Get(chatID, userID)
	if err != nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrUnknown, true, deleteCmd)
		return
	}

	if buyer == nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrNoPointsYet, true, deleteCmd)
		return
	}

	// Convert itemID to int64
	itemID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrInvalidID, true, deleteCmd)
		return
	}

	//Check if boost already exist or not
	// users can buy different type of boost simultaneously but not the same one
	boost, err := app.models.Users.ActiveBoost(userID, chatID)
	if err != nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrNoBoost, true, deleteCmd)
		return
	}

	// don't let users buy another boost if they already have one
	if boost != nil {
		message := fmt.Sprintf("🚫 *You already have an active boost of this type!*\n⏳ *It will expire on `%s`.*\n💡 *Try again after it expires.*",
			boost.ExpiresAt.Format("2006-01-02 15:04:05"))
		sendMessage(ctx, b, chatID, msg.ID, message, true, deleteCmd)
		return
	}

	itm, err := app.models.Shop.Get(itemID)
	if err != nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrNoItemFound, true, deleteCmd)
		return
	}

	if itm == nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrNoItemFound, true, deleteCmd)
		return
	}

	if buyer.Points < itm.Price {
		sendMessage(ctx, b, chatID, msg.ID, ErrNotEnoughPoints, true, deleteCmd)
		return
	}

	//buy the item
	err = app.models.Shop.Buy(userID, chatID, itm.ID, itm.Type, itm.Duration)
	if err != nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrBuyingItem, true, deleteCmd)
		return
	}

	//after successfully buying the item deduct price from the buyer
	err = app.models.Users.Update(chatID, userID, buyer.Points-itm.Price)
	if err != nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrNotEnoughPoints, true, deleteCmd)
		return
	}

	//get boost by item id
	boost, err = app.models.Users.GetBoostByItem(userID, chatID, itm.ID)
	if err != nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrNoBoost, true, deleteCmd)
		return
	}

	if boost == nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrNoBoost, true, deleteCmd)
		return
	}

	p := &database.Point{
		ChatID: chatID,
		Amount: itm.Price,
		UserID: userID,
		Source: pointSources[5],
		Change: "loss",
	}

	//update points history
	err = app.models.Points.Insert(p)
	if err != nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrUnknown, true, deleteCmd)
		return
	}

	message := fmt.Sprintf("✅ *Successfully purchased:* `%s`\n"+
		"⏳ *Expires on:* `%s`",
		itm.Name, boost.ExpiresAt.Format("2006-01-02 15:04:05"),
	)
	sendMessage(ctx, b, chatID, msg.ID, message, true, deleteCmd)
}

// display all available boost
func (app *application) boost(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := update.Message
	userID := msg.From.ID
	chatID := msg.Chat.ID

	// deleteCmd determines if commands should be deleted after execution
	deleteCmd := viper.GetBool("bot.deleteCommand")

	boost, err := app.models.Users.ActiveBoost(userID, chatID)
	if err != nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrNoBoost, true, deleteCmd)
		return
	}

	if boost == nil {
		sendMessage(ctx, b, chatID, msg.ID, ErrNoBoost, true, deleteCmd)
		return
	}

	sendMessage(ctx, b, chatID, msg.ID, formatBoost(boost), true, deleteCmd)
}
