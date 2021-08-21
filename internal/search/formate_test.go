package search_test

import (
	"testing"

	"github.com/Stezok/remanga-worker/internal/search"
)

func TestEraseQuotes(t *testing.T) {
	str := "[] 예상(하)다"
	res := search.EraseQuotes(str)
	expect := " 예상다"
	if res != expect {
		t.Errorf("Expect %s, but got %s on test [] ab(c)d", expect, res)
		t.Fail()
	}
}
