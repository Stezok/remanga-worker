package acqq

import (
	"log"

	"github.com/Stezok/remanga-worker/internal/bot"
	"github.com/Stezok/remanga-worker/internal/listener/acqq"
	"github.com/Stezok/remanga-worker/internal/models"
	"github.com/Stezok/remanga-worker/internal/translate"
)

type ACQQService struct {
	Listener   *acqq.ACQQListener
	Bot        *bot.TelegramBot
	logger     *log.Logger
	translator *translate.Translator

	taskChannel  chan<- models.Task
	errChannel   chan error
	closeChannel chan struct{}
}

func (service *ACQQService) GetLastID() int64 {
	return service.Listener.GetLastID()
}

func (service *ACQQService) ListenErrors() {
	for {
		select {
		case <-service.closeChannel:
			return
		case err := <-service.errChannel:
			service.logger.Print(err)
		}
	}
}

func (service *ACQQService) Run() {
	go service.ListenErrors()
	go service.Bot.Run()
	service.Listener.Listen()
}

func (service *ACQQService) Close() {
	service.Listener.Close()
	service.Bot.Close()
	for i := 0; i < 1; i++ {
		service.closeChannel <- struct{}{}
	}
}

func NewACQQService(lastID int, taskChannel chan<- models.Task, bot *bot.TelegramBot, translator *translate.Translator, logger *log.Logger) *ACQQService {
	service := &ACQQService{
		Bot:        bot,
		logger:     logger,
		translator: translator,

		taskChannel:  taskChannel,
		errChannel:   make(chan error, 10),
		closeChannel: make(chan struct{}, 2),
	}
	service.Listener = acqq.NewACQQListener(service.TelegramNotifyOnFind())
	service.Listener.LastID = int64(lastID)
	return service
}
