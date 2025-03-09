package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/joybiswas007/modbot-tg/config"
	"github.com/joybiswas007/modbot-tg/internal/database"
	"github.com/spf13/viper"
)

type application struct {
	models database.Models
}

func main() {
	var cfgFile string

	flag.StringVar(&cfgFile, "config", "", "config file (default is .modbot.yaml)")
	flag.Parse()

	config.InitConfig(cfgFile)

	token := viper.GetString("bot.token")
	if token == "" {
		log.Fatal("bot.token field is empty")
	}

	db := database.New()
	defer func() {
		db.Close()
		log.Println("Disconnected from database")
	}()

	if err := database.Migrate(db); err != nil {
		log.Fatal(err)
	}

	app := &application{
		models: database.NewModels(db),
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	b, err := bot.New(token)
	if err != nil {
		log.Fatal(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/rank", bot.MatchTypePrefix, ensureGroupChat(app.topUsers))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/gift", bot.MatchTypePrefix, ensureGroupChat(app.gift))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/stats", bot.MatchTypeExact, ensureGroupChat(app.userStats))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/history", bot.MatchTypeExact, ensureGroupChat(adminMiddleware(app.pointHistory)))
	// b.RegisterHandler(bot.HandlerTypeMessageText, "/shop", bot.MatchTypeExact, ensureGroupChat(app.shop))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, app.start)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/id", bot.MatchTypeExact, app.getID)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, app.help)

	b.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypePrefix, ensureGroupChat(app.countMessage))

	me, err := b.GetMe(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch bot info: %v", err)
	}

	fmt.Printf("@%s started...\n", me.Username)

	b.Start(ctx)
}
