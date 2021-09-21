package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Stezok/remanga-worker/internal/bot"
	"github.com/Stezok/remanga-worker/internal/database"
	"github.com/Stezok/remanga-worker/internal/models"
	"github.com/Stezok/remanga-worker/internal/translate"
)

type Worker interface {
	Run() error
	Push(models.Task)
	Close()
}

type Logger interface {
	Print(...interface{})
}

type SearchService struct {
	db     *database.Database
	bot    *bot.TelegramBot
	logger Logger

	translator *translate.Translator
	worker     Worker

	errChannel   chan error
	closeChannel chan struct{}

	wg sync.WaitGroup
}

func (ss *SearchService) Search(keyword string) ([]Document, int, int, error) {
	srchTotalVal := url.Values{}
	srchTotalVal.Add("tSrch_total", keyword)

	qVal := url.Values{}
	qVal.Add("q", keyword)

	data := `wt=json&indent=on&start=0&fq=&detailSearchYn=N&result_detail_txt=&facet.field=EBOOK_YN&totalCnt=275&totalPages=28&rows=10&page=1&cip_id=&ebook_yn=&cip_yn=&tSrch_subject=&bib_yn=&deposit_yn=&h1.preserveMulti=true&parent_facet_yn=&tSrch_title=&tSrch_author=&tSrch_publisher=&tSrch_isbn=&tSrch_control_no=&%s&tSrch_issn=&media_code=&ddc_1s=&pub_status=&acquisit_yn=&sort=INPUT_DATE+DESC&fq_select=tSrch_total&%s`
	// data := `%s&%s`
	data = fmt.Sprintf(data, srchTotalVal.Encode(), qVal.Encode())

	body := bytes.NewBufferString(data)
	req, err := http.NewRequest("POST", "http://seoji.nl.go.kr/landingPage/SearchAjax.do", body)
	if err != nil {
		return nil, 0, 0, err
	}

	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36 OPR/77.0.4054.298")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Origin", "http://seoji.nl.go.kr")
	req.Header.Add("Referer", "http://seoji.nl.go.kr/landingPage/SearchList.do?q=1")
	req.Header.Add("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	// req.Header.Add("Cookie", "WMONID=R74ALiuHDRy; PCID=3e8ba060-8d2f-9aa4-f225-3864fe24856b-1627640010789; JSESSIONID=zhWbSGY1Y5gRIjC48zUNzJPA2dQcKa29NvhwmZmpnr9fVHHop6dXfI0BRbN5BFNw.amV1c19kb21haW4vUFRMV0FTMDJfc2Vvamk=;")

	startTime := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, 0, err
	}
	seps := (time.Now().UnixNano() - startTime.UnixNano()) / 1e6

	startTime = time.Now()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, 0, err
	}

	var respObj SearchResponse
	err = json.Unmarshal(b, &respObj)
	if err != nil {
		return nil, 0, 0, err
	}

	result := make([]Document, 0, len(respObj.Response.Documents)+1)
	borderDate := time.Now().Add(-4 * time.Hour * 24)

	for _, doc := range respObj.Response.Documents {
		if doc.GetTime().Unix() < borderDate.Unix() {
			continue
		}

		unique := true

		ss.db.Transaction(func() {
			for _, id := range append(ss.db.Searched, ss.db.Posted...) {
				if doc.EAISBN == id {
					unique = false
					break
				}
			}
			if unique {
				doc.Title = EraseQuotes(doc.Title)
				ss.db.Searched = append(ss.db.Searched, doc.EAISBN)
				result = append(result, doc)

				ss.wg.Add(2)
				go func(ptr *string, text string) {
					defer ss.wg.Done()
					val, err := ss.translator.Translate(text, translate.KOREAN, translate.ENGLISH)
					if err == nil {
						*ptr = val
					} else {
						ss.logger.Print(err)
					}
				}(&result[len(result)-1].TitleEng, doc.Title)

				go func(ptr *string, text string) {
					defer ss.wg.Done()
					val, err := ss.translator.Translate(text, translate.KOREAN, translate.RUSSIAN)
					if err == nil {
						*ptr = val
					} else {
						ss.logger.Print(err)
					}
				}(&result[len(result)-1].TitleRus, doc.Title)
			}
		})
	}

	peps := (time.Now().UnixNano() - startTime.UnixNano()) / 1e6

	return result, int(seps), int(peps), nil
}

func (ss *SearchService) ListenErrors() {
	for {
		select {
		case err := <-ss.errChannel:
			ss.logger.Print(err)
		case <-ss.closeChannel:
			return
		}
	}
}

func (ss *SearchService) Notify(doc Document) {
	task := models.Task{
		KrName: doc.Title,
		RuName: doc.TitleRus,
		EnName: doc.TitleEng,
		Link:   doc.GetLink(),
		Status: "4",
		Type:   "1",
		Callback: func() {
			title := database.Title{
				Name:      doc.Title,
				NameRu:    doc.TitleRus,
				NameEn:    doc.TitleEng,
				HandledAt: time.Now().Unix(),
				EAISBN:    doc.EAISBN,
			}
			ss.db.Transaction(func() {
				ss.db.Titles = append(ss.db.Titles, title)
			})

			text := fmt.Sprintf(`ℹ️ *Вышла новая запись (seoji):*
		
		🔗 *Ссылка:*
		http://seoji.nl.go.kr/landingPage?isbn=%s
		
		📅 Дата добавления на сайт: %s
		📆 Дата выхода: %s
		
		🇰🇷 *Оригинальное название: *
		%s
		
		🇷🇺 *Название на русском: *
		%s
		
		🇺🇸 *Название на английском:*
		%s`,
				doc.EAISBN, doc.GetTime(), doc.GetPostTime(), doc.Title, doc.TitleRus, doc.TitleEng)
			err := ss.bot.SendMessage(text)
			if err != nil {
				ss.errChannel <- err
			}
		},
	}
	ss.worker.Push(task)
}

func (ss *SearchService) LoopSearch() {
	for {
		select {
		case <-ss.closeChannel:
			return
		default:
			for _, keyword := range ss.db.Querys {
				// fmt.Printf("Поиск %s\n", keyword)
				result, seps, peps, err := ss.Search(keyword)
				if err != nil {
					ss.errChannel <- err
				} else {
					fmt.Printf("Время поиска: %d\n", seps)
					fmt.Printf("Время обработки: %d\n", peps)
					ss.wg.Wait()
					var wg sync.WaitGroup
					for _, res := range result {
						wg.Add(1)
						go func(res Document) {
							defer wg.Done()
							ss.Notify(res)
						}(res)
					}
					wg.Wait()
				}
			}
			err := ss.db.UpdateDB()
			if err != nil {
				ss.errChannel <- err
			}
		}
	}
}

func (ss *SearchService) Run() error {
	err := ss.db.ReadDB()
	if err != nil {
		return err
	}

	go func() {
		defer ss.worker.Close()
		err := ss.worker.Run()
		if err != nil {
			ss.errChannel <- err
		}
	}()

	go func() {
		defer ss.bot.Close()
		err := ss.bot.Run()
		if err != nil {
			ss.errChannel <- err
		}
	}()

	go ss.ListenErrors()
	ss.LoopSearch()

	return nil
}

func (ss *SearchService) Close() {
	for i := 0; i < 2; i++ {
		ss.closeChannel <- struct{}{}
	}
	return
}

func NewSearchService(db *database.Database, bot *bot.TelegramBot, translator *translate.Translator, worker Worker, logger Logger) *SearchService {
	return &SearchService{
		db:           db,
		translator:   translator,
		worker:       worker,
		bot:          bot,
		logger:       logger,
		errChannel:   make(chan error, 10),
		closeChannel: make(chan struct{}, 2),
	}
}
