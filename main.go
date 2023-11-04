package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// buildGoogleUrls builds the google searchUrl,
// specifying the language and pages and country code
func bulidGoogleUrls(searchTerm, countryCode, languageCode string, pages, count int) ([]string, error) {
	toScrape := []string{}
	searchTerm = strings.Trim(searchTerm, " ")
	searchTerm = strings.Replace(searchTerm, " ", "+", -1)
	if googleBase, found := googleDomains[countryCode]; found {
		for i := 0; i < pages; i++ {
			start := i * count
			scrapeURL := fmt.Sprintf("%s%s&num=%d&h1=%s&start=%dfilter=0", googleBase, searchTerm, count, languageCode, start)
			toScrape = append(toScrape, scrapeURL)
		}
	} else {
		err := fmt.Errorf("country %s is not supported", countryCode)
		return nil, err
	}

	return toScrape, nil
}

// scrapeClientRequest sends an request to google
func scrapeClientRequest(searchURL string, proxyString interface{}) (*http.Response, error) {
	baseClient := getScrapeClient(proxyString)
	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", getRandomUserAgent())

	res, err := baseClient.Do(req)
	if res.StatusCode != 200 {
		err := errors.New("scrapper received a non 200 status code, suggesting a ban")
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return res, nil
}

// getScrape creates/replicates client user agent
func getScrapeClient(proxyString interface{}) *http.Client {
	switch v := proxyString.(type) {
	case string:
		proxyUrl, _ := url.Parse(v)
		return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	default:
		return &http.Client{}
	}
}

type SearchResult struct {
	ResultRank  int
	ResultURL   string
	ResultTitle string
	ResultDesc  string
}

// googleResultParsing parse the response body from server
func googleResultParsing(response *http.Response, rank int) ([]SearchResult, error) {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	results := []SearchResult{}
	sel := doc.Find("div.g")
	rank++
	for i := range sel.Nodes {
		item := sel.Eq(i)
		linkTag := item.Find("a")
		link, _ := linkTag.Attr("href")
		titleTag := item.Find("h3.r")
		descTag := item.Find("span.st")
		desc := descTag.Text()
		title := titleTag.Text()
		link = strings.Trim(link, " ")

		if link != "" && link != "#" && !strings.HasPrefix(link, "/") {
			result := SearchResult{
				rank,
				link,
				title,
				desc,
			}
			results = append(results, result)
			rank++
		}
	}

	return results, nil
}

func GoogleScrape(searchTerm, countryCode, languageCode string, proxyString interface{}, pages, count, backoff int) ([]SearchResult, error) {
	results := []SearchResult{}
	resultCounter := 0
	googlePages, err := bulidGoogleUrls(searchTerm, countryCode, languageCode, pages, count)
	if err != nil {
		return results, err
	}
	for _, page := range googlePages {
		res, err := scrapeClientRequest(page, proxyString)
		if err != nil {
			return results, err
		}
		data, err := googleResultParsing(res, resultCounter)
		if err != nil {
			return results, err
		}
		resultCounter += len(data)
		results = append(results, data...)

		time.Sleep(time.Duration(backoff) * time.Second)
	}

	return results, nil
}

func main() {
	res, err := GoogleScrape("searchTerm", "com", "en", nil, 1, 30, 10)
	if err == nil {
		for _, res := range res {
			fmt.Println(res)
		}
	}
}
