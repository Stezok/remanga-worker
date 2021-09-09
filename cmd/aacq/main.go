package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/Stezok/remanga-worker/internal/bot"
	"github.com/Stezok/remanga-worker/internal/service/acqq"
	"github.com/Stezok/remanga-worker/internal/translate"
	"github.com/Stezok/remanga-worker/internal/worker"
)

func main() {

	cnfg, err := acqq.ReadACQQConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// 1811774567:AAFVSUivdtW-nJOCvsW3aZG-Fl3BqwzmW-s
	// 1453696928:AAHWpvYdAWvvUyGDIMjZDmMPeVWBlQIVmho
	// 1991965691:AAFzUPWS-s4WFgTiEE9kmeYs7JY12v7kp4o
	tgBot, err := bot.NewTelegramBot("1991965691:AAFzUPWS-s4WFgTiEE9kmeYs7JY12v7kp4o", []int64{
		496823111,
		768413750,
	}, log.Default())
	if err != nil {
		panic(err)
	}

	workerLogger := log.New(os.Stdout, "WORKER: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	workerConfig := cnfg.Worker
	worker := worker.NewWorker(workerConfig, workerLogger)

	translator := translate.NewTranslator(cnfg.Translate.FolderID, cnfg.Translate.ApiKey)

	acqqServiceLogger := log.New(os.Stdout, "ACQQ SERVICE: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	service := acqq.NewACQQService(cnfg.SearchStart, worker.GetTaskChan(), tgBot, translator, acqqServiceLogger)

	shutdownChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownChannel, os.Interrupt)
	go func() {
		oscall := <-shutdownChannel
		log.Printf("Shutdowning service with os call: %v", oscall)

		cnfg.SearchStart = int(service.GetLastID())
		acqq.UpdateACQQConfig(cnfg, "config.yaml")
		worker.Close()
		service.Close()
	}()

	go worker.Run()
	service.Run()
}
