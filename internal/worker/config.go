package worker

type WorkerConfig struct {
	PathToSelenium string
	SeleniumMode   string
	Port           int

	ProcessCount int
	PathToImage  string

	Login    string
	Password string
}

