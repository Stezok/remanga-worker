package acqq

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ACQQListener struct {
	LastID int64
	mu     sync.Mutex

	Delay  time.Duration
	onFind func(Title)

	closeChan chan struct{}
}

func (acqq *ACQQListener) GetLastID() int64 {
	acqq.mu.Lock()
	defer acqq.mu.Unlock()
	return acqq.LastID
}

func (acqq *ACQQListener) tryGet(id int64) (Title, error) {
	url := fmt.Sprintf("https://ac.qq.com/Comic/comicInfo/id/%d", id)

	resp, err := http.Get(url)
	if err != nil {
		return Title{}, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return Title{}, err
	}
	var titleText string
	doc.Find(".works-intro-title strong").Each(func(_ int, s *goquery.Selection) {
		titleText, _ = s.Html()
	})

	if titleText == "" {
		return Title{}, fmt.Errorf("Title with id %d not found", id)
	}

	var title Title
	title.ID = id
	title.Title = titleText
	title.Link = url

	return title, nil
}

func (acqq *ACQQListener) Check() {
	// log.Print("CHECK FROM ", acqq.LastID)
	acqq.mu.Lock()
	defer acqq.mu.Unlock()
	for i := 1; i <= 10; i++ {
		newID := acqq.LastID + int64(i)
		title, err := acqq.tryGet(newID)
		if err == nil {
			acqq.LastID = newID
			acqq.onFind(title)
			return
		}
	}
}

func (acqq *ACQQListener) Listen() {
	for {
		time.Sleep(acqq.Delay)
		select {
		case <-acqq.closeChan:
			return
		default:
			acqq.Check()
		}
	}
}

func (acqq *ACQQListener) ListenAsync() {
	go acqq.Listen()
}

func (acqq *ACQQListener) Close() {
	for i := 0; i < 1; i++ {
		acqq.closeChan <- struct{}{}
	}
}

func NewACQQListener(onfind Handler) *ACQQListener {
	return &ACQQListener{
		onFind:    onfind,
		closeChan: make(chan struct{}, 1),
	}
}
