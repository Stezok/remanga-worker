package database

type AppData struct {
	Querys   []string `json:"querys"`
	Searched []string `json:"searched"`
	Posted   []string `json:"posted"`
	Titles   []Title  `json:"titles"`
}

type Title struct {
	Name      string `json:"name"`
	NameRu    string `json:"name_ru"`
	NameEn    string `json:"name_en"`
	HandledAt int64  `json:"time_handler"`
	EAISBN    string `json:"EA_ISBN"`
}
