package worker

import (
	"fmt"
	"sync"
	"time"

	"github.com/Stezok/remanga-worker/internal/models"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

type Logger interface {
	Print(...interface{})
}

type Worker struct {
	WorkerConfig
	logger Logger

	TaskChannel  chan models.Task
	closeChannel chan struct{}
	errorChannel chan error
}

func (w *Worker) ListenErrors() {
	for {
		select {
		case err := <-w.errorChannel:
			w.logger.Print(err)
		case <-w.closeChannel:
			return
		}
	}
}

func (w *Worker) Do(task models.Task, wd selenium.WebDriver) error {
	err := Post(wd, task.Type, task.RuName, task.EnName, task.KrName, task.Link)
	if err != nil {
		return err
	}

	task.Callback()
	return nil
}

func (w *Worker) Work(wd selenium.WebDriver) {
	err := Auth(wd, w.Login, w.Password)
	if err != nil {
		w.errorChannel <- err
		return
	}
	time.Sleep(1 * time.Second)
	err = Prepare(wd, w.PathToImage)
	if err != nil {
		w.errorChannel <- err
		return
	}

	for {
		select {
		case task := <-w.TaskChannel:
			err := w.Do(task, wd)
			if err != nil {
				w.errorChannel <- err
			}
			err = Prepare(wd, w.PathToImage)
			if err != nil {
				w.errorChannel <- err
			}
		case <-w.closeChannel:
			return
		}
	}
}

func (w *Worker) Run() error {

	options := []selenium.ServiceOption{}
	service, err := selenium.NewChromeDriverService(w.PathToSelenium, w.Port, options...)
	if err != nil {
		return err
	}
	defer service.Stop()

	caps := selenium.Capabilities{
		"browserName": "chrome",
	}

	args := []string{
		"--no-sandbox",
		"--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/604.4.7 (KHTML, like Gecko) Version/11.0.2 Safari/604.4.7",
	}
	if w.SeleniumMode == SELENIUM_HEADLESS {
		args = append(args, "--headless")
	}

	chromeCaps := chrome.Capabilities{
		Path: "",
		Args: args,
	}
	caps.AddChrome(chromeCaps)
	addr := fmt.Sprintf("http://127.0.0.1:%d/wd/hub", w.Port)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		w.ListenErrors()
	}()

	for i := 0; i < w.ProcessCount; i++ {
		wd, err := selenium.NewRemote(caps, addr)
		if err != nil {
			return err
		}

		wg.Add(1)
		go w.Work(wd)
	}
	wg.Wait()

	return nil
}

func (w *Worker) Push(task models.Task) {
	w.TaskChannel <- task
}

func (w *Worker) Close() {
	for i := 0; i < 1+w.ProcessCount; i++ {
		w.closeChannel <- struct{}{}
	}
}

func (w *Worker) GetTaskChan() chan<- models.Task {
	return w.TaskChannel
}

func NewWorker(config WorkerConfig, logger Logger) *Worker {
	return &Worker{
		logger:       logger,
		WorkerConfig: config,
		TaskChannel:  make(chan models.Task, 4*config.ProcessCount),
		closeChannel: make(chan struct{}, config.ProcessCount),
		errorChannel: make(chan error, config.ProcessCount),
	}
}
