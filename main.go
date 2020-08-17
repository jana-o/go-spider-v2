package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type fetchResult struct {
	version  string
	title    string
	headings map[string]int
	urls     []string
}

type sortResult struct {
	internals    int
	inaccessible int
	login        bool
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

func main() {
	inputURL := os.Args[1]
	if inputURL == "" {
		log.Fatalln("missing url")
	}

	doc, err := parsePage(inputURL)
	if err != nil || doc == nil {
		return
	}
	//collect fetchResult from site
	fresult := fetch(doc)
	fmt.Printf("website title: %s \nwith HTML version %s\n", fresult.title, fresult.version)
	fmt.Println("headings:", fresult.headings)

	//analyse urls from fresult
	parsed, err := url.Parse(inputURL)
	if err != nil {
		panic(err)
	}
	baseURL := parsed.Scheme + "://" + parsed.Host

	//find internal links
	findinternals := func(s string) bool {
		return strings.HasPrefix(s, baseURL) || strings.HasPrefix(s, "/") || strings.HasPrefix(s, "#")
	}
	internals := filter(fresult.urls, findinternals)
	fmt.Printf("found %d internal links and %d external links \n", len(internals), len(fresult.urls)-len(internals))

	//check if link is inaccessible
	pingLink := func(link string) bool {
		_, err := http.Get(link)
		if err != nil {
			return true
		}
		return false
	}
	inaccessible := filter(internals, pingLink)
	fmt.Printf("found %d inaccessible links\n", len(inaccessible))

	//check if internal links contain login (could be done with regex as well)
	containsLoginByURL := func(il string) bool {
		s := strings.ToUpper(il)
		return strings.Contains(s, "LOGIN") || strings.Contains(s, "SIGNIN")
	}
	login := filter(internals, containsLoginByURL)
	if len(login) == 0 {
		fmt.Println("no login link found")
	} else {
		fmt.Println("login link found")
	}
}

//filter finds sublist of links
func filter(ss []string, f func(string) bool) (filtered []string) {
	for _, s := range ss {
		if f(s) {
			filtered = append(filtered, s)
		}
	}
	return
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

//Contains returns true if slice already contains url
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

// search doc for form. Inside form I look for an input of type or id password
// I assume that that the user needs to input a password because there are too many labels for name/username/email etc.
// May have to look for oauth aus well
func searchForm(doc *goquery.Document) bool {
	doc.Find(".form").Each(func(i int, s *goquery.Selection) {
		form := s.Find("#Password").Text()
		fmt.Println("form", form)
	})
	return true
}

//displays results
func display(fr *fetchResult, r *sortResult) {
	fmt.Printf("Website title: %s \nHTML version: %s\nHeadings count by level:\n", fr.title, fr.version)
	for k, v := range fr.headings {
		fmt.Printf("%d - %s\n", v, k)
	}

	fmt.Printf("Amount of internal links: %d\namount of innaccessible links: %d\n", r.internals, r.inaccessible)
	fmt.Printf("Contains login is: %t", r.login)
}
