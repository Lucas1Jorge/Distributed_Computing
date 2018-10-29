package main

import (
	"fmt"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Global
var results chan *fakeResult

func fetch(url string, depth int, fetcher Fetcher, results chan *fakeResult) {
	// if depth > 0 {
		body, urls, _ := fetcher.Fetch(url)
		results <- &fakeResult{body, urls}
	// }
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	if depth <= 0 {
		// fmt.Println("Depth = 0")
		return
	}

	if results == nil {
		results = make(chan *fakeResult)
	}

	done := make(map[string]bool)

	go fetch(url, depth, fetcher, results)
	done[url] = true

	for found := 1; found > 0; found-- {
		result := <-results

		fmt.Printf("found: %s %q\n", url, result.body)

		for _, URL := range result.urls {
			if !done[URL] {
				found += 1
				go fetch(URL, depth - 1, fetcher, results)
				done[URL] = true
			}
		}
	}

	// body, urls, err := fetcher.Fetch(url)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Printf("found: %s %q\n", url, body)
	// for _, u := range urls {
	// 	if _, ok := f[u]; !ok {
	// 		Crawl(u, depth-1, fetcher)
	// 		f[u] = true
	// 	}
	// }
	
	close(results)
}

func main() {
	Crawl("https://golang.org/", 4, fetcher)
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
