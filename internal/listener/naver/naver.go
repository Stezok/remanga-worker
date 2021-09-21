package naver

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type NaverListener struct {
	Found []int64
	mu    sync.Mutex

	Logger log.Logger

	Delay  time.Duration
	onFind func(Title)

	doneChan  chan struct{}
	closeChan chan struct{}
	errChan   chan error
}

func (nl *NaverListener) ListenErros() {
	for {
		select {
		case <-nl.closeChan:
			return
		case err := <-nl.errChan:
			nl.Logger.Print(err)
		}
	}
}

func (nl *NaverListener) IsUnique(id int64) bool {
	for _, f := range nl.Found {
		if f == id {
			return false
		}
	}
	return true
}

func (nl *NaverListener) AddFound(ids ...int64) {
	nl.Found = append(nl.Found, ids...)
}

func (nl *NaverListener) GetFound() []int64 {
	return nl.Found
}

func (nl *NaverListener) FindNewTitles() ([]Title, error) {
	resp, err := http.Get("https://series.naver.com/comic/recentList.series")
	if err != nil {
		return []Title{}, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return []Title{}, err
	}
	var titles []Title
	doc.Find("ul.lst_thum.v1.lst_thum_last li a").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("href")
		title, _ := s.Attr("title")

		splited := strings.Split(link, "=")
		// log.Print(splited[len(splited)-1])
		id, _ := strconv.ParseInt(splited[len(splited)-1], 10, 64)

		var photo string
		photo, _ = s.Find("img").Attr("src")
		photo = strings.Split(photo, "?")[0]
		// log.Print(photo)
		titles = append(titles, Title{
			ID:    id,
			Title: title,
			Link:  "https://series.naver.com" + link,
			Photo: photo,
		})
	})

	filtered := titles[:0]
	for i := 0; i < len(titles); i++ {
		if nl.IsUnique(titles[i].ID) {
			filtered = append(filtered, titles[i])
		}
	}
	titles = titles[:len(filtered)]
	return titles, nil
}

func (nl *NaverListener) Listen() {
	for {
		time.Sleep(nl.Delay)
		select {
		case <-nl.closeChan:
			return
		default:
			titles, err := nl.FindNewTitles()
			if err != nil {
				nl.errChan <- err
				break
			}
			// log.Printf("%#v", titles)
			for _, title := range titles {
				nl.onFind(title)
				nl.AddFound(title.ID)
			}
		}
	}
}

func (nl *NaverListener) Close() {
	for i := 0; i < 2; i++ {
		nl.closeChan <- struct{}{}
	}
	return
}

func NewNaverListener(onfind Handler) *NaverListener {
	return &NaverListener{
		onFind:    onfind,
		errChan:   make(chan error, 10),
		doneChan:  make(chan struct{}, 2),
		closeChan: make(chan struct{}, 2),
	}
}
