package models

type Task struct {
	ID       int64
	RuName   string
	EnName   string
	KrName   string
	Link     string
	Callback func()
}
