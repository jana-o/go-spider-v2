package main

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func Testcontains(t *testing.T) {
	ss := []string{"hello", "hi", "hola"}
	ans := contains(ss, "hi")
	if ans != true {
		t.Errorf("got %t, want true", ans)
	}
}

func Testfilter(t *testing.T) {
	baseURL := " http://symbolic.com/"

	findinternals := func(s string) bool {
		return strings.HasPrefix(s, baseURL) || strings.HasPrefix(s, "/") || strings.HasPrefix(s, "#")
	}
	ss := []string{"/home", "/", "http://external./com", "#home", "fffff#ffff/f", " http://symbolic.com/linksomewhere"}

	want := []string{"/home", "/", "#home", " http://symbolic.com/linksomewhere"}
	got := filter(ss, findinternals)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
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
