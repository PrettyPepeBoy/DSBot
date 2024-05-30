package main

import (
	"DiscordBot/cmd/discordbot"
	"DiscordBot/cmd/logger"
	"DiscordBot/config"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"os"
	"os/signal"
	"strings"
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

func ddd() {

	var log *slog.Logger
	sess, err := discordgo.New("Bot MTI0NDkyMjA4OTE1NzE2NTA4OA.GtZjfW.gv0s3kR2VDRjY8Hfz0d2NJOB8NcgrW8T5if-tY")
	if err != nil {
		panic("failed to connect with discord bot")
	}

	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		prefix := "!Bot_Iluha"

		command := strings.Split(m.Content, "/")

		if len(command) == 1 {
			_, err = s.ChannelMessageSend(m.ChannelID, "write any command")
			if err != nil {
				log.Info("failed to send message to chanel")
			}
		}

		if command[0] != prefix {
			log.Info("command line should start with prefix")
			return
		}

		if m.Content == "help" {
			//todo set command lines

		}

		if m.Content == "hello" {
			_, err = s.ChannelMessageSend(m.ChannelID, "world")
			if err != nil {
				panic("failed to send message to chanel")
			}
		}
	})

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()
	if err != nil {
		panic("failed to open discord session")
	}
	defer sess.Close()
	fmt.Println("bot is online")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-ch
}
