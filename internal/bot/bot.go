package bot

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Stezok/remanga-worker/internal/database"
	"github.com/Stezok/remanga-worker/internal/hash"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Logger interface {
	Print(...interface{})
}

type TelegramBot struct {
	bot      *tgbotapi.BotAPI
	database database.Database
	logger   Logger

	admins []int64

	callbacks map[string]func()
	mu        sync.Mutex

	closeChannel chan struct{}
	errChannel   chan error
}

func (tb *TelegramBot) ListenErrors() {
	for {
		select {
		case err := <-tb.errChannel:
			tb.logger.Print("Telegram bot:", err)
		case <-tb.closeChannel:
			return
		}
	}
}

func (tb *TelegramBot) HandleMessage(update tgbotapi.Update) bool {
	if update.Message == nil || update.Message.Chat == nil {
		return false
	}

	msgArr := strings.Split(update.Message.Text, " ")
	var start, end int
	serr := fmt.Errorf("Bad length")
	eerr := fmt.Errorf("Bad length")

	if len(msgArr) == 3 {
		start, serr = strconv.Atoi(msgArr[1])
		end, eerr = strconv.Atoi(msgArr[2])
	}

	tb.database.Transaction(func() {
		totalPosts := len(tb.database.Titles)
		if len(msgArr) != 3 || serr != nil || eerr != nil || start < 0 || end > totalPosts {
			text := fmt.Sprintf(`Используйте "/history num1 num2"
			num 1 - Начиная от какого поста
			Заканчивая каким постом
			Всего постов в базе: %d`, totalPosts)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
			_, err := tb.bot.Send(msg)
			if err != nil {
				tb.errChannel <- err
			}
		}

		text := ``
		for i := start; i <= end; i++ {
			post := tb.database.Titles[i]
			text += fmt.Sprintf(`
			
			🔗 Ссылка: http://seoji.nl.go.kr/landingPage?isbn=%s
			📅 Дата обработки %s
			🇰🇷 Оригинальное название: %s
			🇷🇺 Название на русском: %s
			🇺🇸 Название на английском: %s`,
				post.EAISBN, time.Unix(int64(post.HandledAt), 0), post.Name, post.NameRu, post.NameEn)
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		_, err := tb.bot.Send(msg)
		if err != nil {
			tb.errChannel <- err
		}
	})
	return true
}

func (tb *TelegramBot) HandleCallback(update tgbotapi.Update) bool {
	// log.Print("recive")
	if update.CallbackQuery == nil || update.CallbackQuery.Data == "" {
		return false
	}

	key := update.CallbackQuery.Data
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.callbacks[key]()
	delete(tb.callbacks, key)

	edit := tgbotapi.EditMessageReplyMarkupConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      update.CallbackQuery.Message.Chat.ID,
			MessageID:   update.CallbackQuery.Message.MessageID,
			ReplyMarkup: nil,
		},
	}
	_, err := tb.bot.Send(edit)
	if err != nil {
		tb.errChannel <- err
	}

	return true
}

func (tb *TelegramBot) HandleRequests(updates tgbotapi.UpdatesChannel) {
	for {
		select {
		case <-tb.closeChannel:
			return
		case update := <-updates:
			switch {
			case tb.HandleCallback(update):
			case tb.HandleMessage(update):
			}
		}
	}
}

func (tb *TelegramBot) Run() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := tb.bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}
	go tb.ListenErrors()
	tb.HandleRequests(updates)
	return nil
}

func (tb *TelegramBot) Close() {
	for i := 0; i < 2; i++ {
		tb.closeChannel <- struct{}{}
	}
}

func (tb *TelegramBot) SendMessageWithCallback(text, callbackText string, callback func()) error {
	for _, admin := range tb.admins {
		msg := tgbotapi.NewMessage(admin, text)

		randString := hash.RandomString(32)
		// log.Print("RANDOM ", randString)
		tb.callbacks[randString] = callback

		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(callbackText, randString),
			),
		)
		msg.ParseMode = "Markdown"
		// msg.DisableWebPagePreview = true
		_, err := tb.bot.Send(msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tb *TelegramBot) SendMessage(text string) error {
	for _, admin := range tb.admins {
		msg := tgbotapi.NewMessage(admin, text)
		msg.ParseMode = "Markdown"
		_, err := tb.bot.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewTelegramBot(token string, admins []int64, logger Logger) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &TelegramBot{
		bot:       bot,
		admins:    admins,
		logger:    logger,
		callbacks: make(map[string]func()),
	}, nil
}
