package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/Stezok/remanga-worker/internal/bot"
	"github.com/Stezok/remanga-worker/internal/database"
	"github.com/Stezok/remanga-worker/internal/search"
	"github.com/Stezok/remanga-worker/internal/translate"
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
		SeleniumMode: worker.SELENIUM_HEADLESS,
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

	translator := translate.NewTranslator("b1g6d3iaj8fvksk5lu69", "AQVNyVcSdLfJvhW8HMAz4WnX8ZQjS0yt1XaDN0a8")

	database := database.NewDatabase("./db.json")
	service := search.NewSearchService(database, bot, translator, w, logger)

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
