package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

//fetchResult is
type fetchResult struct {
	version  string
	title    string
	headings map[string]int
	urls     []string
	login    bool
}

//parse returns *goquery documents
func parse(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	//check status code
	if res.StatusCode != http.StatusOK {
		log.Fatalf("Error response status code was %d", res.StatusCode)
	}

	// Create a goquery document from the HTTP response
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body ", err)
	}
	return doc, nil
}

func main() {
	url := "http://symbolic.com/"

	// url := os.Args[1]
	// if url == "" {
	// 	os.Exit(1)
	// }

	doc, err := parse(url)
	if err != nil {
		return
	}
	if doc == nil {
		return
	}

	fresult := fetch(doc)
	fmt.Println("FR", fresult)

	//make channels
	c := make(chan string)

	//checkLinks concurrently
	for _, u := range fresult.urls {
		go checkLink(u, c)
	}

	// receive inaccessible links from channel
	ia := []string{}
	for l := range c {
		ia = append(ia, l)
	}
	fmt.Println("inaccessible links", ia)

	//internal/external and in/accessible loop
	il := countIl(url, fresult.urls)
	fmt.Println("count internal urls: ", il)

}

func fetch(doc *goquery.Document) fetchResult {
	fr := fetchResult{}

	v, err := versionReader(doc)
	if err != nil {
		fmt.Println("Error loading version", err)
	}
	fr.version = v
	fr.title = doc.Find("title").Contents().Text()
	fr.headings = getHeadings(doc)
	fr.urls = getURL(doc)

	// findForm(doc)

	return fr
}

//checkLink checks if link is accessible
func checkLink(link string, c chan string) {
	_, err := http.Get(link)
	if err != nil {
		// fmt.Println(link, "down")
		c <- link //send to channel
		return
	}
	// fmt.Println(link, "is up")
	// c <- link
	time.Sleep(3 * time.Second)
	close(c)
}

//countIl counts the internal links found on site
func countIl(baseURL string, urls []string) int {
	il := 0
	for _, u := range urls {
		if strings.HasPrefix(u, baseURL) {
			il++
		}
		if strings.HasPrefix(u, "/") {
			il++
		}
	}
	return il
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
			// fmt.Println(headings)
		})
	}
	// fmt.Println(hs)
	return hs
}

// Called for each HTML element found
func getURL(doc *goquery.Document) []string {
	foundUrls := []string{}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		u, _ := s.Attr("href")
		foundUrls = append(foundUrls, u)
	})
	return foundUrls
}

//checks HTML version and returns first match
func versionReader(doc *goquery.Document) (string, error) {
	var doctypes = map[string]string{
		"HTML 5":                 `<!DOCTYPE html>`,
		"HTML 4.01 Strict":       `"-//W3C//DTD HTML 4.01//EN"`,
		"HTML 4.01 Transitional": `"-//W3C//DTD HTML 4.01 Transitional//EN"`,
		"HTML 4.01 Frameset":     `"-//W3C//DTD HTML 4.01 Frameset//EN"`,
		"XHTML 1.0 Strict":       `"-//W3C//DTD XHTML 1.0 Strict//EN"`,
		"XHTML 1.0 Transitional": `"-//W3C//DTD XHTML 1.0 Transitional//EN"`,
		"XHTML 1.0 Frameset":     `"-//W3C//DTD XHTML 1.0 Frameset//EN"`,
		"XHTML 1.1":              `"-//W3C//DTD XHTML 1.1//EN"`,
	}
	//html version?? //http://symbolic.com/  =>  XHTML 1.0 Transitional
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

// TODO
//findForm checks if page contains a login form
func findForm(doc *goquery.Document) bool {
	fmt.Println("Hello")
	//find "form" _ children
	// "login" class, value, submit
	doc.Find("form").Each(func(i int, s *goquery.Selection) {
		l, _ := s.Attr("value")
		if l == "login" || l == "Login" {
			return
		}
	})
	return false
}
