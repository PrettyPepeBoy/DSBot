package main

import (
	"DiscordBot/cmd/discordbot"
	"DiscordBot/cmd/logger"
	"DiscordBot/config"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustSetupConfig()
	log := logger.MustSetupLogger(cfg.Environment)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()
	discordbot.MustStartSession(ctx, cfg, log)
	<-ctx.Done()
	log.Info("stopping bot")
	time.Sleep(time.Second * 5)
}
