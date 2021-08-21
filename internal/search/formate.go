package search

import "fmt"

func EraseQuotes(s string) string {
	tmp := []rune(s)
	ss := tmp[:0]
	bal := 0

	for i := 0; i < len(tmp); i++ {
		if tmp[i] == '(' || tmp[i] == '[' {
			bal++
			continue
		}
		if tmp[i] == ')' || tmp[i] == ']' {
			bal--
			continue
		}

		if bal == 0 {
			ss = append(ss, tmp[i])
		}
	}

	return string(ss)
}

func MakeLink(id string) string {
	return fmt.Sprintf("http://seoji.nl.go.kr/landingPage?isbn=%s", id)
}
