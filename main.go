package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
		fmt.Errorf("scrapper received a non 200 status code, suggesting a ban")
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return res, nil
}

func getScrapeClient(proxyString interface{}) *http.Client {
	switch v := proxyString.(type) {
	case string:
		proxyUrl, _ := url.Parse(v)
		return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	default:
		return &http.Client{}
	}
}
