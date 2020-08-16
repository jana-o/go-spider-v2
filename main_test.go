package main

import (
	"reflect"
	"strings"
	"testing"
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
