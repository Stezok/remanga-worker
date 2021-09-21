package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/Stezok/remanga-worker/internal/bot"
	"github.com/Stezok/remanga-worker/internal/service/naver"
	"github.com/Stezok/remanga-worker/internal/translate"
	"github.com/Stezok/remanga-worker/internal/worker"
)

func main() {
	cnfg, err := naver.ReadNaverConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	bot, err := bot.NewTelegramBot("2046485020:AAGuMsJHgL0ZavUkIPeWAiND7KAVALinHcI", []int64{
		496823111,
		768413750,
	}, log.Default())
	if err != nil {
		log.Fatal(err)
	}

	worker := worker.NewWorker(cnfg.Worker, log.Default())

	translator := translate.NewTranslator(cnfg.Translate.FolderID, cnfg.Translate.ApiKey)

	service := naver.NewNaverService(worker.GetTaskChan(), bot, translator, log.Default())

	shutdownChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownChannel, os.Interrupt)
	endChan := make(chan struct{})
	go func() {
		oscall := <-shutdownChannel
		log.Printf("Shutdowning service with os call: %v", oscall)

		worker.Close()
		service.Close()
		endChan <- struct{}{}
	}()

	go worker.Run()
	service.Run()
	<-endChan
}
