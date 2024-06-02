package discordbot

import (
	"DiscordBot/config"
	"errors"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"strconv"
	"strings"
)

type BotState struct {
	Request      []string
	DefaultState bool
	RequestState bool
	TriesCount   int
}

var state = BotState{DefaultState: true}

var discordHandlers = map[string]func(s *discordgo.Session, m *discordgo.MessageCreate, log *slog.Logger, state *BotState){
	"register": func(s *discordgo.Session, m *discordgo.MessageCreate, log *slog.Logger, state *BotState) {
		_, err := s.ChannelMessageSend(m.ChannelID, "начат процесс регистрации")
		if err != nil {
			failedMessageOccurred(log, err, "registerHandler")
		}
		_, err = s.ChannelMessageSend(m.ChannelID, "пожалуйста, введи ваш email")
		if err != nil {
			failedMessageOccurred(log, err, "registerHandler")
		}

		state.RequestState = true
		state.Request = []string{"email", "password"}
	},
	"email": func(s *discordgo.Session, m *discordgo.MessageCreate, log *slog.Logger, state *BotState) {
		//todo убрать s и m из getMail
		err := getMail(s, m, log, m.Content)
		if err != nil {
			_, err = s.ChannelMessageSend(m.ChannelID, "емейл не прошел проверку, введите его еще раз")
			if err != nil {
				failedMessageOccurred(log, err, "emailHandler")
			}
			state.TriesCount++
			if state.TriesCount == 3 {
				_, err = s.ChannelMessageSend(m.ChannelID, "количество попыток закончилось")
				if err != nil {
					failedMessageOccurred(log, err, "emailHandler")
				}
				state.RequestState = false
			}
			return
		}
		state.Request = state.Request[1:]
		if len(state.Request) == 0 {
			log.Info("finish request pipeline")
			state.RequestState = false
			return
		}
		log.Info("request pipeline", state.Request)

		_, err = s.ChannelMessageSend(m.ChannelID, "пожалуйста, введите ваш Пароль")

		if err != nil {
			failedMessageOccurred(log, err, "emailHandler")
		}
	},
	"password": func(s *discordgo.Session, m *discordgo.MessageCreate, log *slog.Logger, state *BotState) {
		err := getPassword(s, m, log, m.Content)
		if err != nil {
			_, err = s.ChannelMessageSend(m.ChannelID, "пароль не верный, введите его еще раз")
			if err != nil {
				failedMessageOccurred(log, err, "passwordHandler")
			}
			return
		}
		state.Request = state.Request[1:]
		if len(state.Request) == 0 {
			state.RequestState = false
			log.Info("finish request pipeline")
			return
		}
		log.Info("request pipeline", state.Request)
	},
}

func InitHandlers(cfg config.Config, log *slog.Logger) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		const op = "discordbot/InitHandlers"

		if m.Author.ID == s.State.User.ID {
			return
		}

		log.Info("current state", slog.String("state", strconv.FormatBool(state.RequestState)))

		if !state.RequestState {
			str := strings.Split(m.Content, "/")

			if str[0] != cfg.DiscordBot.Prefix {
				log.Info("message was send not to bot", slog.String("message", m.Content))
				return
			}

			if len(str) == 1 {
				log.Info("there was not any handler in message", slog.String("message", m.Content))
				_, err := s.ChannelMessageSend(m.ChannelID, "боту не было отправлено команды")
				if err != nil {
					failedMessageOccurred(log, err, op)
				}
				return
			}
			if len(str) > 2 {
				log.Info("to many commands in message", slog.String("message", m.Content))
				_, err := s.ChannelMessageSend(m.ChannelID, "боту было отправлено более чем одна команда")
				if err != nil {
					failedMessageOccurred(log, err, op)
				}
				return
			}
			getHandler(s, m, str[1], log, &state)
		} else {
			_, err := s.ChannelMessageSend(m.ChannelID, "мы в текущем состоянии")
			if err != nil {
				failedMessageOccurred(log, err, op)
			}
			if m.Author.ID == s.State.User.ID {
				return
			}
			getHandler(s, m, state.Request[0], log, &state)
		}
	}
}

func getHandler(s *discordgo.Session, m *discordgo.MessageCreate, handler string, log *slog.Logger, state *BotState) {
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
	fn(s, m, log, state)
}

func getMail(s *discordgo.Session, m *discordgo.MessageCreate, log *slog.Logger, mail string) error {
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
			return errors.New("incorrect email")
		}
	}
	if len(mail) < 8 {
		log.Info("Email to short")
		_, err := s.ChannelMessageSend(m.ChannelID, "ваш email слишком короткий")
		if err != nil {
			failedMessageOccurred(log, err, op)
			return errors.New("incorrect email")
		}
	}
	if len(mail) > 30 {
		log.Info("Email to big")
		_, err := s.ChannelMessageSend(m.ChannelID, "ваш email слишком длинный")
		if err != nil {
			failedMessageOccurred(log, err, op)
			return errors.New("incorrect email")
		}
	}

	log.Info("successfully parse mail", slog.String("mail", mail))
	_, err := s.ChannelMessageSend(m.ChannelID, "ваш email успешно добавлен в базу")
	if err != nil {
		failedMessageOccurred(log, err, op)
		return nil
	}
	return nil
}

func getPassword(s *discordgo.Session, m *discordgo.MessageCreate, log *slog.Logger, password string) error {
	if len(password) < 8 {
		log.Info("password to short")
		return errors.New("password to short")
	}

	if len(password) > 30 {
		log.Info("password to big")
		return errors.New("password to big")
	}

	return nil
}

func failedMessageOccurred(log *slog.Logger, err error, op string) {
	log.Info("failed to send message to channel", slog.String("op", op))
	log.Error("error occurred", slog.String("error", err.Error()))
}
