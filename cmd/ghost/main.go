package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"
)

type config struct {
	regex   string
	term string
	terms string
	timeout int
	url     string
}

type ghost struct {
	client *http.Client
	config config
	query  interface{}
}

func main() {
	var config config
	flag.StringVar(&config.regex, "r", "", "regex pattern for searching")
	flag.StringVar(&config.term, "term", "", "term or phrase for searching")
	flag.StringVar(&config.terms, "terms", "", "file containing term list for searching")
	flag.IntVar(&config.timeout, "t", 0, "timeout in milliseconds")
	flag.StringVar(&config.url, "u", "", "url for searching")
	flag.Parse()

	start := time.Now()

	g := &ghost{
		config: config,
	}

	g.client = g.makeClient(config.timeout)
	g.query = regexp.MustCompile(config.regex)

	if config.url == "" {
		log.Fatal("search url must be provided")
	}

	g.getQuery()

	const snapPrefix = "http://web.archive.org/cdx/search/cdx?output=json&url="
	u := fmt.Sprintf("%s%s", snapPrefix, config.url)
	body, err := g.getData(u, g.client)
	if err != nil {
		log.Fatal(err)
	}

	snaps, err := g.getSnaps(body)
	if err != nil {
		fmt.Println(err)
	}

	var filteredSnaps []string
	for _, v := range snaps {
		if v[4] == "200" {
			filteredSnaps = append(filteredSnaps, v[1])
		}
	}

	for _, u := range filteredSnaps {
		url := fmt.Sprintf("https://web.archive.org/web/%s/%s", u, config.url)
		page, err := g.getData(url, g.client)
		if err != nil {
			fmt.Println(err)
		}
		g.parsePage(string(page), g.query)
	}

	fmt.Printf("took: %f seconds\n", time.Since(start).Seconds())
}
