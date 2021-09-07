package acqq

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/Stezok/remanga-worker/internal/translate"
	"github.com/Stezok/remanga-worker/internal/worker"
	"gopkg.in/yaml.v2"
)

type ACQQConfig struct {
	SearchStart int                       `yaml:"start_id"`
	Translate   translate.TranslateConfig `yaml:"translate"`
	Worker      worker.WorkerConfig       `yaml:"worker"`
}

func ReadACQQConfig(path string) (ACQQConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return ACQQConfig{}, err
	}
	defer file.Close()

	configData, err := ioutil.ReadAll(file)
	if err != nil {
		return ACQQConfig{}, err
	}

	var config ACQQConfig
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return ACQQConfig{}, err
	}

	return config, nil
}

func BackupACQQConfig(path string) error {
	out, err := os.Create(path + ".BACKUP")
	if err != nil {
		return err
	}
	defer out.Close()

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(out, file)
	return err
}

func UpdateACQQConfig(config ACQQConfig, path string) error {
	err := BackupACQQConfig(path)
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(b)
	return err
}
