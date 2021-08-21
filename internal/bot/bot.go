package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Stezok/remanga-worker/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramBot struct {
	bot      *tgbotapi.BotAPI
	database database.Database

	admins []int64

	closeChannel chan struct{}
	errChannel   chan error
}

func (tb *TelegramBot) ListenErrors() {
	for {
		select {
		case err := <-tb.errChannel:
			log.Print("Telegram bot:", err)
		case <-tb.closeChannel:
			return
		}
	}
}

func (tb *TelegramBot) HandleUpdate(update tgbotapi.Update) {
	if update.Message == nil || update.Message.Chat == nil {
		return
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
			return
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
}

func (tb *TelegramBot) HandleRequests(updates tgbotapi.UpdatesChannel) {
	for {
		select {
		case <-tb.closeChannel:
			return
		case update := <-updates:
			tb.HandleUpdate(update)
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

func (tb *TelegramBot) SendNotify(text string) error {
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

func NewTelegramBot(token string, admins []int64) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &TelegramBot{
		bot:    bot,
		admins: admins,
	}, nil
}