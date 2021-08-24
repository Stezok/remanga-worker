package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/Stezok/remanga-worker/internal/bot"
	"github.com/Stezok/remanga-worker/internal/database"
	"github.com/Stezok/remanga-worker/internal/search"
	"github.com/Stezok/remanga-worker/internal/worker"
)

func main() {
	logFile, err := os.Create("output.log")
	if err != nil {
		log.Fatal(err)
	}
	logger := log.New(logFile, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)

	config := worker.WorkerConfig{
		PathToSelenium: "./chromedriver",
		// PathToSelenium: "C:/Drivers/chromedriver.exe",
		SeleniumMode: "headless",
		Port:         9515,
		ProcessCount: 4,
		// PathToImage:  "C:/Users/Артем/Desktop/Arti Manga Downloader/parser/1.jpg",
		PathToImage: "/root/parser/remanga-worker/cmd/app/1.jpg",
		Login:       "Leaderq",
		Password:    "12345QawDse",
	}
	w := worker.NewWorker(config, logger)

	bot, err := bot.NewTelegramBot("1811774567:AAFVSUivdtW-nJOCvsW3aZG-Fl3BqwzmW-s", []int64{
		496823111,
		768413750,
	}, logger)
	if err != nil {
		log.Fatal(err)
	}

	database := database.NewDatabase("./db.json")
	service := search.NewSearchService(database, bot, w, logger)

	shutdownChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownChannel, os.Interrupt)
	go func() {
		oscall := <-shutdownChannel
		log.Printf("Shutdowning service with os call: %v", oscall)
		service.Close()
	}()

	err = service.Run()
	if err != nil {
		log.Fatal(err)
	}
}
