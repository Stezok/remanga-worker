package acqq

import (
	"fmt"
	"sync"

	"github.com/Stezok/remanga-worker/internal/listener/acqq"
	"github.com/Stezok/remanga-worker/internal/models"
	"github.com/Stezok/remanga-worker/internal/translate"
)

func (service *ACQQService) TelegramNotifyOnFind() func(acqq.Title) {
	return func(title acqq.Title) {
		var wg sync.WaitGroup

		var ruName string
		wg.Add(1)
		go func(ruName *string, text string) {
			defer wg.Done()
			var err error
			*ruName, err = service.translator.Translate(text, translate.CHINESE, translate.RUSSIAN)
			if err != nil {
				service.errChannel <- err
			}
		}(&ruName, title.Title)

		var enName string
		wg.Add(1)
		go func(enName *string, text string) {
			defer wg.Done()
			var err error
			*enName, err = service.translator.Translate(text, translate.CHINESE, translate.ENGLISH)
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
		}

		wg.Wait()
		text := `
		Новый[ ](%s)комикс на ACQQ!
		🔗 Ссылка: %s
		🇨🇳 Оригинальное название: %s
		🇷🇺 Название на русском: %s
		🇺🇸 Название на английском: %s`
		text = fmt.Sprintf(text, title.Photo, title.Link, title.Title, ruName, enName)

		err := service.Bot.SendMessageWithCallback(text, "Опубликовать", func() {
			service.taskChannel <- models.Task{
				ID:       title.ID,
				KrName:   title.Title,
				RuName:   ruName,
				EnName:   enName,
				Type:     "2",
				Link:     title.Link,
				Callback: onPublish,
			}
		})
		if err != nil {
			service.errChannel <- err
		}

	}
}
