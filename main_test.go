package main

import (
	"net/http"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

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
