package translate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type TranslateReq struct {
	FolderID           string   `json:"folderId"`
	Texts              []string `json:"texts"`
	TargetLanguageCode string   `json:"targetLanguageCode"`
	SourceLanguageCode string   `json:"sourceLanguageCode"`
}

type TranslateResponse struct {
	Translations []struct {
		Text string `json:"text"`
	} `json:"translations"`
}

const (
	ENGLISH = "en"
	KOREAN  = "ko"
	RUSSIAN = "ru"
	CHINESE = "zh"
)

type Translator struct {
	folderId string
	apiKey   string
}

func (t *Translator) Translate(text, from, to string) (string, error) {
	translateReq := TranslateReq{
		FolderID:           t.folderId,
		Texts:              []string{text},
		TargetLanguageCode: to,
		SourceLanguageCode: from,
	}
	// log.Print(translateReq)
	b, err := json.Marshal(translateReq)
	if err != nil {
		return "", err
	}
	body := bytes.NewReader(b)

	req, err := http.NewRequest("POST", "https://translate.api.cloud.yandex.net/translate/v2/translate", body)
	if err != nil {
		return "", err
	}
	apiKeyHeader := fmt.Sprintf("Api-Key %s", t.apiKey)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", apiKeyHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var translateResp TranslateResponse
	err = json.Unmarshal(b, &translateResp)
	if err != nil {
		return "", err
	}

	if len(translateResp.Translations) == 0 {
		return "", errors.New("TRANSLATE: No data response")
	}

	return translateResp.Translations[0].Text, nil
}

func NewTranslator(folderId string, apiKey string) *Translator {
	return &Translator{
		folderId: folderId,
		apiKey:   apiKey,
	}
}
