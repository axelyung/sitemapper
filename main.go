package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"
)

var (
	baseURL       string              // url of home page
	fileName      string              // name of file to write sitemap to
	includeQuery  bool                // whether or not to include query urls
	maxDepth      int                 // max depth of recursive scraping, set -1 for no limit
	addresses     sync.Map            // mapped to Location
	dates         sync.Map            // mapped to LastMod
	points        sync.Map            // mapped to Priority
	hostRegExp    *regexp.Regexp      // match href contents against this
	scrapersGroup sync.WaitGroup      // wait group for scraper routines
	urlNormalizer func(string) string // a function to normalize urls
)

type page struct {
	Location string  `xml:"loc"`
	LastMod  string  `xml:"lastmod"`
	Priority float64 `xml:"priority"`
}

type urlset struct {
	Xmlns string `xml:"xmlns,attr"`
	Urls  []page `xml:"url"`
}

func main() {
	givenURL := ""
	givenURL, fileName, maxDepth, includeQuery = readArguments()

	// set start time
	start := time.Now()
	parsedURL, err := url.Parse(givenURL)
	catch(err)

	host := parsedURL.Host
	fmt.Println("Creating sitemap for", host)
	fmt.Println("")

	urlNormalizer = createURLNormalizer(host)
	baseURL = "http://" + host + "/"

	hostRegExp = createRegExpMatcher(host)

	// set address to prevent revisit
	addresses.Store(baseURL, baseURL)
	points.Store(baseURL, float64(1))
	scrapersGroup.Add(1)
	// launch goroutine for homeURL
	go fetchAndScrape(baseURL, 0)
	// wait for all fetchAndScrape routines to finish
	scrapersGroup.Wait()

	normalizePoints()
	xml := createURLSet()

	encodeAndWriteToXML(fileName, xml)

	// print summary
	fmt.Println("\nSUMMARY:")
	fmt.Println("Host:", host)
	fmt.Println("Sitemap:", fileName)
	fmt.Println("Visited:", len(xml.Urls))
	fmt.Println("Depth:", maxDepth)
	fmt.Println("Excecution time:", time.Since(start))
}

func createURLSet() urlset {
	toXML := urlset{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
	}

	addresses.Range(func(k interface{}, a interface{}) bool {
		d, _ := dates.Load(k)
		p, _ := points.Load(k)
		toXML.Urls = append(toXML.Urls, page{
			a.(string),
			d.(string),
			p.(float64),
		})
		return true
	})

	return toXML
}

// Given a target URL, find all linked pages on same domain,
// save metadata to relevant maps and repeat recursively for
// all found URLs. Skip recursion if max depth is reached.
func fetchAndScrape(targetURL string, depth int) {
	defer scrapersGroup.Done()

	// add protocol + host if missing
	targetURL = urlNormalizer(targetURL)

	// fetch page
	resp, err := http.Get(targetURL)
	catch(err)
	defer resp.Body.Close()

	// check if response is html
	contentType := resp.Header.Get("content-type")
	isHTML, err := regexp.MatchString("^text/html", contentType)
	if !isHTML {
		// if page is not html then delete from
		// addresses and cancel rest of scrape
		addresses.Delete(targetURL)
		return
	}

	// set page data
	dates.Store(targetURL, getLastModifiedHeader(resp))

	// read html string
	html, err := ioutil.ReadAll(resp.Body)
	catch(err)

	// scrape page for href tags
	fmt.Println("Scraping", targetURL)
	foundLinks := scrape(string(html))
	fmt.Println("Found", len(foundLinks), "link(s) on", targetURL)

	// update points for found links
	updatePoints(foundLinks)

	// we don't want exceed the max depth
	// of pages to recursively traverse
	if maxDepth > -1 && depth > maxDepth {
		return
	}

	// visit filteredURLS
	for _, newURL := range foundLinks {
		// check if already visited
		if urlTaken(newURL) {
			continue
		}
		scrapersGroup.Add(1)
		go fetchAndScrape(newURL, depth+1)
	}
}

// given html string find all relevant urls
func scrape(htmlStr string) []string {
	// find all hrefs
	urls := parse(htmlStr)

	// filter to host
	filteredURLs := []string{}
	for _, href := range urls {
		u := href
		if hostRegExp.MatchString(u) {
			u = urlNormalizer(u)
			filteredURLs = appendIfMissing(filteredURLs, u)
		}
	}

	return filteredURLs
}

// check if another scraper is currently
// working on the same url, if not then
func urlTaken(u string) bool {
	_, taken := addresses.LoadOrStore(u, u)
	return taken
}
