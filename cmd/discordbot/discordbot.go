package discordbot

import (
	"DiscordBot/config"
	"DiscordBot/internal/discordbot"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"os"
)

func MustStartSession(ctx context.Context, cfg config.Config, log *slog.Logger) {
	const op = "discordbot/MustStartSession"

	env := os.Getenv("DSBOT_TOKEN")
	if env == "" {
		fmt.Println("DSBOT_TOKEN is not set")
		os.Exit(1)
	}

	sess, err := discordgo.New("Bot " + env)
	if err != nil {
		log.Info("failed to start session with discord bot", slog.String("op", op))
		log.Error("error occurred", slog.String("error", err.Error()))
		return
	}

	sess.AddHandler(discordbot.InitHandlers(cfg, log))

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()
	if err != nil {
		log.Error("failed to open discord session", slog.String("error", err.Error()))
		return
	}
	defer sess.Close()

	log.Info("discord bot online")
	<-ctx.Done()
}
