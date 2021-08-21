package search

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type TranslateReq struct {
	Q      string `json:"q"`
	Source string `json:"source"`
	Target string `json:"target"`
}

const (
	ENGLISH = "en"
	KOREAN  = "ko"
	RUSSIAN = "ru"
)

func Translate(text, from, to string) (string, error) {
	// val := url.Values{}
	// val.Add("text", text)

	// params := fmt.Sprintf(`translate_to=%s&translate_from=%s&%s&TTK=197520.377331`, to, from, val.Encode())

	// body := bytes.NewBufferString(params)
	// req, _ := http.NewRequest("POST", "https://www.m-translate.it/api/2/translate", body)
	// req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	// req.Header.Add("Origin", "https://www.m-translate.ru")
	// req.Header.Add("Referer", "https://www.m-translate.ru/")
	// req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36")
	// req.Header.Add("pragma", "no-cache")

	// resp, _ := http.DefaultClient.Do(req)
	// io.Copy(os.Stdout, resp.Body)

	data := TranslateReq{
		Q:      text,
		Source: from,
		Target: to,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://translate.astian.org/translate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	jsonObj := make(map[string]interface{})
	b, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(b, &jsonObj)
	if err != nil {
		return "", err
	}
	// log.Print(jsonObj)
	return jsonObj["translatedText"].(string), nil
}
