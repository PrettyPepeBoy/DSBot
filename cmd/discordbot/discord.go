package discordbot

import (
	"DiscordBot/config"
	"context"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"strings"
)

var discordHandlers = map[string]func(s *discordgo.Session, m *discordgo.MessageCreate, log *slog.Logger, request string){
	"register": func(s *discordgo.Session, m *discordgo.MessageCreate, log *slog.Logger, request string) {
		getMail(s, m, log, request)
		return
	},
}

func MustStartSession(ctx context.Context, cfg config.Config, log *slog.Logger) {
	const op = "discordbot/MustStartSession"
	sess, err := discordgo.New("Bot " + cfg.DiscordBot.Token)
	if err != nil {
		log.Info("failed to start session with discord bot", slog.String("op", op))
		log.Error("error occurred", slog.String("error", err.Error()))
		return
	}

	chHandler, chRequest := initChannels()
	close(chRequest)

	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		str := strings.Split(m.Content, "/")

		if str[0] != cfg.DiscordBot.Prefix {
			log.Info("message was send not to bot", slog.String("message", m.Content))
			return
		}

		if len(str) == 1 {
			log.Info("there was not any handler in message", slog.String("message", m.Content))
			_, err = s.ChannelMessageSend(m.ChannelID, "боту не было отправлено команды")
			if err != nil {
				failedMessageOccurred(log, err, op)
			}
			return
		}

		if len(str) > 2 {
			log.Info("to many commands in message", slog.String("message", m.Content))
			_, err = s.ChannelMessageSend(m.ChannelID, "боту было отправлено более чем одна команда")
			if err != nil {
				failedMessageOccurred(log, err, op)
			}
			return
		}
		chHandler <- str[1]
		getHandler(s, m, <-chHandler, log)
	})

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

func getHandler(s *discordgo.Session, m *discordgo.MessageCreate, handler string, log *slog.Logger) {
	const op = "discordbot/getHandler"
	fn, ok := discordHandlers[handler]
	if !ok {
		log.Info("incorrect handler", slog.String("handler", handler))
		_, err := s.ChannelMessageSend(m.ChannelID, "такой функции не существует")
		if err != nil {
			failedMessageOccurred(log, err, op)
		}
		return
	}
	log.Info("found handler", slog.String("handler", handler))
	fn(s, m, log, handler)
}

func register(s *discordgo.Session, m *discordgo.MessageCreate, cfg config.Config, log *slog.Logger) func() {
	return func() {
		request := strings.Split(m.Content, "/")
		if request[0] != cfg.DiscordBot.Prefix {
			return
		}

		if request[1] != "register" {
			return
		}
		//mail.Address{}

	}
}

func getMail(s *discordgo.Session, m *discordgo.MessageCreate, log *slog.Logger, mail string) {
	var emailSuffixes = []string{"@gmail.com", "@mail.ru", "@yandex.ru"}
	const op = "discordbot/getMail"

	for i, elem := range emailSuffixes {
		if strings.HasSuffix(mail, elem) {
			break
		}
		if i == len(emailSuffixes)-1 {
			log.Info("incorrect email suffix", slog.String("mail", mail))
			_, err := s.ChannelMessageSend(m.ChannelID, "ваш email некоректный")
			if err != nil {
				failedMessageOccurred(log, err, op)
			}
			return
		}
	}
	if len(mail) < 8 {
		log.Info("Email to short")
		_, err := s.ChannelMessageSend(m.ChannelID, "ваш email слишком короткий")
		if err != nil {
			failedMessageOccurred(log, err, op)
			return
		}
	}
	if len(mail) > 30 {
		log.Info("Email to big")
		_, err := s.ChannelMessageSend(m.ChannelID, "ваш email слишком длинный")
		if err != nil {
			failedMessageOccurred(log, err, op)
			return
		}
	}

	log.Info("successfully parse mail", slog.String("mail", mail))
	_, err := s.ChannelMessageSend(m.ChannelID, "ваш email успешно добавлен в базу")
	if err != nil {
		failedMessageOccurred(log, err, op)
		return
	}
}

func initChannels() (chan string, chan string) {
	chHandler := make(chan string)
	chRequest := make(chan string)
	return chHandler, chRequest
}

func failedMessageOccurred(log *slog.Logger, err error, op string) {
	log.Info("failed to send message to channel", slog.String("op", op))
	log.Error("error occurred", slog.String("error", err.Error()))
}
