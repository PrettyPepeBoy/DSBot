package discordbot

import (
	"DiscordBot/config"
	"DiscordBot/internal/discordbot"
	"context"
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

func MustStartSession(ctx context.Context, cfg config.Config, log *slog.Logger) {
	const op = "discordbot/MustStartSession"
	sess, err := discordgo.New("Bot " + cfg.DiscordBot.Token)
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
