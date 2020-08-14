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

//Fetchresult is
type Fetchresult struct {
	version  string
	title    string
	urls     []string
	headings []string
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
	// url := "http://jana.berlin"
	url := "http://symbolic.com/"
	doc, err := parse(url)
	if err != nil {
		return
	}
	if doc == nil {
		return
	}

	version, err := versionReader(doc)
	if err != nil {
		fmt.Println("Error loading version", err)
	}
	fmt.Println("version: ", version)

	title := doc.Find("title").Contents().Text()
	fmt.Println("Title: ", title)

	hs := getHeadings(doc)
	fmt.Println("headings:", hs)

	urls := getURL(doc)
	fmt.Println("URLS:", urls)

	//make channels
	c := make(chan string)

	//checkLinks concurrently
	for _, u := range urls {
		go checkLink(u, c)
	}

	// receive inaccessible links from channel
	ia := []string{}
	for l := range c {
		ia = append(ia, l)
	}
	fmt.Println("inaccessible links", ia)

	//internal/external and in/accessible loop
	il := countIl(url, urls)
	fmt.Println("count internal urls: ", il)

}

// Find all headings H1-H6
// todo: return map with H*: count
func getHeadings(doc *goquery.Document) []string {
	hs := []string{}
	for i := 1; i <= 6; i++ {
		str := strconv.Itoa(i)
		doc.Find("h" + str).Each(func(i int, s *goquery.Selection) {
			// fmt.Println(s.Text())
			hs = append(hs, s.Text())
		})
	}
	return hs
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

// Called for each HTML element found
func getURL(doc *goquery.Document) []string {
	foundUrls := []string{}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		u, _ := s.Attr("href")
		foundUrls = append(foundUrls, u)
	})
	return foundUrls
}

//helper func checks HTML version and returns first match
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
