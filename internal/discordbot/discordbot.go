package discordbot

import (
	"DiscordBot/config"
	"DiscordBot/internal/discordbot/dsErrors"
	"bytes"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
)

type BotState struct {
	RequestPipeline []string
	DefaultState    bool
	RequestState    bool
	TriesCount      int
}

type UserData struct {
	email    string
	password string
}

const triesAmount = 3

var state = BotState{DefaultState: true}
var u = UserData{}

var discordHandlers = map[string]func(s *discordgo.Session, m *discordgo.MessageCreate, log *slog.Logger, state *BotState){
	"register": func(s *discordgo.Session, m *discordgo.MessageCreate, log *slog.Logger, state *BotState) {

		_, err := s.ChannelMessageSend(m.ChannelID, "начат процесс регистрации")
		if err != nil {
			failedMessageOccurred(log, err, "registerHandler")
		}

		_, err = s.ChannelMessageSend(m.ChannelID, "пожалуйста, введите ваш email")
		if err != nil {
			failedMessageOccurred(log, err, "registerHandler")
		}

		state.setRequestStateTrue()
		state.RequestPipeline = []string{"email", "password"}
		log.Info("request pipeline", state.RequestPipeline)
	},
	"email": func(s *discordgo.Session, m *discordgo.MessageCreate, log *slog.Logger, state *BotState) {
		emailAddress, err := getMail(log, m.Content)

		if err != nil {
			_, err = s.ChannelMessageSend(m.ChannelID, "емейл не корректен, попробуйте еще раз")
			if err != nil {
				failedMessageOccurred(log, err, "emailHandler")
			}
			state.triesUp()
			if checkTries(state) {
				_, err = s.ChannelMessageSend(m.ChannelID, "количество попыток закончилось")
				if err != nil {
					failedMessageOccurred(log, err, "emailHandler")
				}
				state.setRequestStateFalse()
				return
			}
			_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("осталось |%v| попыток", triesAmount-state.TriesCount))
			return
		}

		u.email = emailAddress
		if empty := state.checkRequestPipeline(); empty {
			_, err = s.ChannelMessageSend(m.ChannelID, "процесс завершен")
			if err != nil {
				failedMessageOccurred(log, err, "passwordHandler")
			}
			log.Info("finish request pipeline")
			return
		}

		log.Info("request pipeline", state.RequestPipeline)

		_, err = s.ChannelMessageSend(m.ChannelID, "пожалуйста, введите ваш Пароль")
		if err != nil {
			failedMessageOccurred(log, err, "emailHandler")
		}
	},
	"password": func(s *discordgo.Session, m *discordgo.MessageCreate, log *slog.Logger, state *BotState) {
		err := getPassword(log, m.Content)
		if err != nil {
			if errors.Is(err, dsErrors.ErrPasswordTooBig) {
				_, err = s.ChannelMessageSend(m.ChannelID, "пароль не верный, введите его еще раз")
				if err != nil {
					failedMessageOccurred(log, err, "passwordHandler")
				}

				_, err = s.ChannelMessageSend(m.ChannelID, "возникла ошибка : "+dsErrors.ErrPasswordTooBig.Error())
				if err != nil {
					failedMessageOccurred(log, err, "passwordHandler")
				}
				return
			}

			if errors.Is(err, dsErrors.ErrPasswordTooShort) {
				_, err = s.ChannelMessageSend(m.ChannelID, "пароль не верный, введите его еще раз")
				if err != nil {
					failedMessageOccurred(log, err, "passwordHandler")
				}

				_, err = s.ChannelMessageSend(m.ChannelID, "возникла ошибка : "+dsErrors.ErrPasswordTooShort.Error())
				if err != nil {
					failedMessageOccurred(log, err, "passwordHandler")
				}
				return
			}
		}

		if empty := state.checkRequestPipeline(); empty {
			_, err = s.ChannelMessageSend(m.ChannelID, "процесс завершен")
			if err != nil {
				failedMessageOccurred(log, err, "passwordHandler")
			}
			log.Info("finish request pipeline")
			return
		}

		log.Info("request pipeline", state.RequestPipeline)
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
			log.Info("bot in request state now")
			if m.Author.ID == s.State.User.ID {
				return
			}
			getHandler(s, m, state.RequestPipeline[0], log, &state)
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

func getMail(log *slog.Logger, request string) (string, error) {
	e, err := mail.ParseAddress(request)

	if err != nil {
		return "", dsErrors.ErrInvalidEmailAddress
	}

	reqWordsCache, err := http.NewRequest(http.MethodPost, "http://localhost:8081/cache/words", bytes.NewBuffer([]byte(e.Address)))

	if err != nil {
		log.Error("failed to create reqWordsCache", slog.String("error", err.Error()))
		return "", err
	}

	client := http.Client{}

	_, err = client.Do(reqWordsCache)
	if err != nil {
		log.Error("failed to do request to words", slog.String("error", err.Error()))
		return "", err
	}

	log.Info("successfully parse mail", slog.String("email", request))
	return e.Address, nil
}

func getPassword(log *slog.Logger, password string) error {
	if len(password) < 8 {
		log.Info("password to short")
		return dsErrors.ErrPasswordTooShort
	}

	if len(password) > 20 {
		log.Info("password to big")
		return dsErrors.ErrPasswordTooBig
	}

	return nil
}

func failedMessageOccurred(log *slog.Logger, err error, op string) {
	log.Info("failed to send message to channel", slog.String("op", op))
	log.Error("error occurred", slog.String("error", err.Error()))
}

func (b *BotState) setRequestStateTrue() {
	b.RequestState = true
}

func (b *BotState) setRequestStateFalse() {
	b.RequestState = false
}

func checkTries(b *BotState) bool {
	if b.TriesCount == triesAmount {
		return true
	}
	return false
}

func (b *BotState) triesUp() {
	b.TriesCount++
}

func (b *BotState) checkRequestPipeline() bool {
	b.RequestPipeline = b.RequestPipeline[1:]
	if len(b.RequestPipeline) == 0 {
		b.setRequestStateFalse()
		return true
	}
	return false
}
