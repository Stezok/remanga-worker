package worker

type SeleniumMode int

const (
	SELENIUM_DEFAULT SeleniumMode = iota
	SELENIUM_HEADLESS
)

type WorkerConfig struct {
	SeleniumMode   SeleniumMode
	PathToSelenium string `yaml:"seleniumPath"`
	Port           int    `yaml:"port"`
	PathToImage    string `yaml:"pathToImage"`
	ProcessCount   int    `yaml:"processCount"`

	Login    string `yaml:"login"`
	Password string `yaml:"password"`
}
