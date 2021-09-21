package naver

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Stezok/remanga-worker/internal/listener/naver"
	"github.com/Stezok/remanga-worker/internal/models"
	"github.com/Stezok/remanga-worker/internal/translate"
)

func (service *NaverService) TelegramNotifyOnFind() func(naver.Title) {
	return func(title naver.Title) {
		var wg sync.WaitGroup

		var ruName string
		wg.Add(1)
		go func(ruName *string, text string) {
			defer wg.Done()
			var err error
			*ruName, err = service.translator.Translate(text, translate.KOREAN, translate.RUSSIAN)
			if err != nil {
				service.errChannel <- err
			}
		}(&ruName, title.Title)

		var enName string
		wg.Add(1)
		go func(enName *string, text string) {
			defer wg.Done()
			var err error
			*enName, err = service.translator.Translate(text, translate.KOREAN, translate.ENGLISH)
			if err != nil {
				service.errChannel <- err
			}
		}(&enName, title.Title)

		onPublish := func() {
			text := "Комикс %s опубликован\nСcылка: %s"
			text = fmt.Sprintf(text, title.Title, title.Link)
			err := service.Bot.SendMessage(text)
			if err != nil {
				service.errChannel <- err
			}

			for {
				err = os.Remove(title.Title + ".png")
				if err == nil {
					break
				}
				time.Sleep(1 * time.Second)
			}
		}

		wg.Wait()
		text := `
		Новый[ ](%s)комикс на Naver!
		🔗 Ссылка: %s
		🇨🇳 Оригинальное название: %s
		🇷🇺 Название на русском: %s
		🇺🇸 Название на английском: %s`
		text = fmt.Sprintf(text, title.Photo, title.Link, title.Title, ruName, enName)

		err := service.Bot.SendMessageWithCallback(text, "Опубликовать", func() {
			resp, err := http.Get(title.Photo)
			if err != nil {
				service.logger.Print(err)
				return
			}
			defer resp.Body.Close()

			file, err := os.Create(title.Title + ".png")
			if err != nil {
				service.logger.Print(err)
				return
			}
			defer file.Close()

			io.Copy(file, resp.Body)
			file.Close()
			path, _ := os.Getwd()

			service.taskChannel <- models.Task{
				ID:       title.ID,
				KrName:   title.Title,
				RuName:   ruName,
				EnName:   enName,
				Status:   "1",
				Type:     "1",
				Link:     title.Link,
				Photo:    path + "\\" + file.Name(),
				Callback: onPublish,
			}
		})
		if err != nil {
			service.errChannel <- err
		}

	}
}
