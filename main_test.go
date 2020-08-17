package main

import (
	"net/http"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestContains(t *testing.T) {
	tables := []struct {
		x   []string
		y   string
		ans bool
	}{
		{x: []string{"a", "b", "c"}, y: "h", ans: false},
		{x: []string{"h", "fhh", "3c"}, y: "h", ans: true},
	}

	for _, table := range tables {
		ans := contains(table.x, table.y)
		if ans != table.ans {
			t.Errorf("got: %t, want: %t.", ans, table.ans)
		}
	}
}
func TestTitleReader(t *testing.T) {
	resp, err := http.Get("http://symbolic.com/")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		t.Fatal(err)
	}

	if v := doc.Find("title").Contents().Text(); v != "Welcome!" {
		t.Fatalf("expected title 'Welcome!', got '%s'", v)
	}
}

type MatcherMock struct {
	sortMock func([]byte) (int, error)
}

func (m MatcherMock) sort(p []byte) (int, error) {
	return m.sortMock(p)
}
