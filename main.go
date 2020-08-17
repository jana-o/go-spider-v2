package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

//fetchResult contains information found on website
type fetchResult struct {
	version  string
	title    string
	headings map[string]int
	urls     []string
}

//link contains information about the type of link
type link struct {
	url          string
	linkType     string
	inaccessible bool
	login        bool
}

// Matcher defines the behavior required to implement search type
type Matcher interface {
	sort(l *link, baseURL string) (*link, error)
}

type myMatcher struct{}

var matcher myMatcher
var links []*link

//sort implements the behavior for the matcher
func (m myMatcher) sort(l *link, baseURL string) (*link, error) {
	//check if link is internal
	isInternals := strings.HasPrefix(l.url, baseURL) || strings.HasPrefix(l.url, "/") || strings.HasPrefix(l.url, "#")
	if isInternals {
		l.linkType = "internal"

		_, err := http.Get(baseURL + "/" + l.url)
		if err != nil {
			l.inaccessible = true
		}
	} else {
		_, err := http.Get(l.url)
		if err != nil {
			l.inaccessible = true
		}
	}

	//check if internal links contain login
	s := strings.ToUpper(l.url)
	containsLoginURL := strings.Contains(s, "LOGIN") || strings.Contains(s, "SIGNIN")
	if containsLoginURL {
		l.login = true
	}

	return l, nil
}

func main() {
	inputURL := os.Args[1]
	if inputURL == "" {
		log.Fatalln("missing url")
	}

	doc, err := parsePage(inputURL)
	if err != nil || doc == nil {
		return
	}
	parsed, err := url.Parse(inputURL)
	if err != nil {
		panic(err)
	}
	baseURL := parsed.Scheme + "://" + parsed.Host

	//collect fetchResult from site
	fResult := fetch(doc)

	//make channel and wait to process all urls
	results := make(chan *link)
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(fResult.urls))

	// check each url in goroutine
	for _, url := range fResult.urls {
		nl := &link{
			url: url, linkType: "external"}

		go func(matcher Matcher, nl *link) {
			match(matcher, nl, baseURL, results)
			waitGroup.Done()
		}(matcher, nl)
	}

	//monitor when all the work is done
	go func() {
		waitGroup.Wait()
		close(results)
	}()

	display(fResult, results)
	fmt.Println("the end")
}

//match analyses each url
func match(matcher Matcher, l *link, baseURL string, results chan<- *link) {
	sortResults, err := matcher.sort(l, baseURL)
	if err != nil {
		log.Println(err)
		return
	}
	results <- sortResults
}

//display blocks channel until result is written
func display(fr *fetchResult, results chan *link) {
	//map would be better,
	//idea of Link struct was to persist data and search more convenient with map[int]*link but does not work
	//now I sort/filter multiple times which I wanted to avoid.
	for result := range results {
		log.Printf("Result:\n url: %s type: %s inaccessible: %t login: %t\n\n", result.url, result.linkType, result.inaccessible, result.login)
		links = append(links, result)
	}

	fmt.Printf("Website title: %s \nHTML version: %s\nHeadings count by level:\n", fr.title, fr.version)
	for k, v := range fr.headings {
		fmt.Printf("%d - %s\n", v, k)
	}
	return
}

//parsePage returns *goquery documents
func parsePage(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	//check status code
	if res.StatusCode != http.StatusOK {
		log.Fatalf("Error response status code was %d", res.StatusCode)
	}

	//create a goquery document from the HTTP response
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body ", err)
	}
	return doc, nil
}

//fetch finds elements on website and returns a fetchResult
func fetch(doc *goquery.Document) *fetchResult {
	fr := fetchResult{}

	v, err := versionReader(doc)
	if err != nil {
		fmt.Println("Error loading version", err)
	}
	fr.version = v
	fr.title = doc.Find("title").Contents().Text()
	fr.headings = getHeadings(doc)
	fr.urls = getURLs(doc)

	return &fr
}

// getHeadings finds all headings H1-H6 and returns map of headings count by level
func getHeadings(doc *goquery.Document) map[string]int {
	hs := map[string]int{
		"h1": 0,
		"h2": 0,
		"h3": 0,
		"h4": 0,
		"h5": 0,
		"h6": 0,
	}
	for i := 1; i <= 6; i++ {
		str := strconv.Itoa(i)
		doc.Find("h" + str).Each(func(i int, s *goquery.Selection) {
			hs["h"+str] = +1
		})
	}
	return hs
}

//getURLs finds all urls and returns slice of unique urls
//the contains check could be removed if urls do not need to be unique
func getURLs(doc *goquery.Document) []string {
	foundUrls := []string{}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		u, _ := s.Attr("href")
		if !contains(foundUrls, u) {
			foundUrls = append(foundUrls, u)
		}
	})
	return foundUrls
}

//contains returns true if slice already contains url
func contains(urls []string, url string) bool {
	for _, v := range urls {
		if v == url {
			return true
		}
	}
	return false
}

//versionReader finds HTML version and returns first match
func versionReader(doc *goquery.Document) (string, error) {
	doctypes := map[string]string{
		"HTML 5":                 `<!DOCTYPE html>`,
		"HTML 4.01 Strict":       `"-//W3C//DTD HTML 4.01//EN"`,
		"HTML 4.01 Transitional": `"-//W3C//DTD HTML 4.01 Transitional//EN"`,
		"HTML 4.01 Frameset":     `"-//W3C//DTD HTML 4.01 Frameset//EN"`,
		"XHTML 1.0 Strict":       `"-//W3C//DTD XHTML 1.0 Strict//EN"`,
		"XHTML 1.0 Transitional": `"-//W3C//DTD XHTML 1.0 Transitional//EN"`,
		"XHTML 1.0 Frameset":     `"-//W3C//DTD XHTML 1.0 Frameset//EN"`,
		"XHTML 1.1":              `"-//W3C//DTD XHTML 1.1//EN"`,
	}
	//e.g. http://symbolic.com/  =>  XHTML 1.0 Transitional
	html, err := doc.Html()
	if err != nil {
		return "", err
	}
	version := ""
	for d, m := range doctypes {
		if strings.Contains(html, m) {
			version = d
		}
	}
	return version, nil
}

// search doc for form. Inside form I look for an input of id password
// I assume that that the user needs to input a password because there are too many labels for name/username/email etc.
// May have to look for oauth aus well. Use searchForm if no url with login found
func searchForm(doc *goquery.Document) bool {
	doc.Find(".form").Each(func(i int, s *goquery.Selection) {
		form := s.Find("#Password").Text()
		fmt.Println("form", form)
	})
	return true
}
