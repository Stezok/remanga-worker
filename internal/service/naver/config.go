package naver

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/Stezok/remanga-worker/internal/translate"
	"github.com/Stezok/remanga-worker/internal/worker"
	"gopkg.in/yaml.v2"
)

type NaverConfig struct {
	Worker    worker.WorkerConfig       `yaml:"worker" json:"-"`
	Translate translate.TranslateConfig `yaml:"translate" json:"-"`
	TitlesID  []int64                   `json:"titles_id"`
}

func ReadNaverConfig(path string) (NaverConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return NaverConfig{}, err
	}
	defer file.Close()

	configData, err := ioutil.ReadAll(file)
	if err != nil {
		return NaverConfig{}, err
	}

	var config NaverConfig
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return NaverConfig{}, err
	}

	return config, nil
}

func ReadConfig(filename string) (NaverConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return NaverConfig{}, err
	}
	defer file.Close()

	d, err := ioutil.ReadAll(file)
	if err != nil {
		return NaverConfig{}, err
	}
	var conf NaverConfig
	err = json.Unmarshal(d, &conf)
	return conf, err
}

func UpdateConfig(conf NaverConfig, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := json.Marshal(conf)
	if err != nil {
		return err
	}
	_, err = file.Write(b)
	return err
}
