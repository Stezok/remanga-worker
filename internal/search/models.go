package search

import (
	"fmt"
	"time"
)

type Document struct {
	Title          string `json:"TITLE"`
	TitleEng       string `json:"-"`
	TitleRus       string `json:"-"`
	EAISBN         string `json:"EA_ISBN"`
	PublishPredate string `json:"PUBLISH_PREDATE"`
	InputDate      string `json:"INPUT_DATE"`
}

func (doc *Document) GetLink() string {
	return MakeLink(doc.EAISBN)
}

func (doc *Document) GetTime() time.Time {
	defer recover()
	date := fmt.Sprintf("%s-%s-%s", doc.InputDate[:4], doc.InputDate[4:6], doc.InputDate[6:8])
	t, _ := time.Parse("2006-01-02", date)
	return t
}

func (doc *Document) GetPostTime() time.Time {
	defer recover()
	date := fmt.Sprintf("%s-%s-%s", doc.PublishPredate[:4], doc.PublishPredate[4:6], doc.PublishPredate[6:8])
	t, _ := time.Parse("2006-01-02", date)
	return t
}

type Response struct {
	Documents []Document `json:"docs"`
}

type SearchResponse struct {
	Response Response `json:"response"`
}
