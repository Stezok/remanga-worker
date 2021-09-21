package naver

import (
	"log"
	"time"

	"github.com/Stezok/remanga-worker/internal/bot"
	"github.com/Stezok/remanga-worker/internal/listener/naver"
	"github.com/Stezok/remanga-worker/internal/models"
	"github.com/Stezok/remanga-worker/internal/translate"
)

type NaverService struct {
	Listener   *naver.NaverListener
	Bot        *bot.TelegramBot
	logger     *log.Logger
	translator *translate.Translator

	taskChannel  chan<- models.Task
	errChannel   chan error
	closeChannel chan struct{}
}

func (service *NaverService) ListenErrors() {
	for {
		select {
		case <-service.closeChannel:
			return
		case err := <-service.errChannel:
			service.logger.Print(err)
		}
	}
}

func (service *NaverService) Run() {
	go service.ListenErrors()
	go service.Bot.Run()
	go service.Listener.ListenErros()
	service.Listener.Listen()
}

func (service *NaverService) Close() {
	service.Listener.Close()
	service.Bot.Close()

	for i := 0; i < 1; i++ {
		service.closeChannel <- struct{}{}
	}
	config := NaverConfig{
		TitlesID: service.Listener.GetFound(),
	}
	log.Print(config)
	UpdateConfig(config, "config.json")
}

func NewNaverService(taskChannel chan<- models.Task, bot *bot.TelegramBot, translator *translate.Translator, logger *log.Logger) *NaverService {
	service := &NaverService{
		Bot:        bot,
		logger:     logger,
		translator: translator,

		taskChannel:  taskChannel,
		errChannel:   make(chan error, 10),
		closeChannel: make(chan struct{}, 2),
	}
	service.Listener = naver.NewNaverListener(service.TelegramNotifyOnFind())
	conf, _ := ReadConfig("config.json")
	service.Listener.Found = conf.TitlesID
	service.Listener.Delay = time.Second * 5

	return service
}
