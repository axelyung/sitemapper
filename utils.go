package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// Create regular expression for a given
// domain/host name to match new URLs against
func createRegExpMatcher(host string) *regexp.Regexp {
	regexpStr := strings.Join([]string{
		"^((https?://)?(",
		regexp.QuoteMeta(host),
		"))?(/.*)?$",
	}, "")
	return regexp.MustCompile(regexpStr)
}

// Creates utility to normalize url to standard format
func createURLNormalizer(host string) func(string) string {
	return func(u string) string {
		ret := ""
		parsedURL, err := url.Parse(u)
		catch(err)
		// add default protocol if missing
		if parsedURL.Scheme == "" {
			ret = ret + "http://"
		} else {
			ret = ret + parsedURL.Scheme + "://"
		}
		// add host if missing
		if parsedURL.Host == "" {
			ret = ret + host
		} else {
			ret = ret + parsedURL.Host
		}
		// path cannot be empty according to W3 spec
		// (https://www.w3.org/Protocols/rfc2616/rfc2616-sec3.html#sec3.2.2)
		// plus we want to prevent lookup of both http://example.com
		// and http://example.com/
		if parsedURL.Path == "" {
			ret = ret + "/"
		} else {
			ret = ret + parsedURL.Path
		}
		// return with query if enabled
		if includeQuery && parsedURL.RawQuery != "" {
			ret = ret + "?" + parsedURL.RawQuery
		}
		return ret
	}

}

// Given an http.Response pointer return the "last-modified"
// or "date" header in proper format. If neither header is
// present, return and empty string
func getLastModifiedHeader(resp *http.Response) string {
	lastmod := ""
	switch {
	case resp.Header.Get("last-modified") != "":
		lastmod = resp.Header.Get("last-modified")
	case resp.Header.Get("date") != "":
		lastmod = resp.Header.Get("date")
	default:
		return lastmod
	}
	t, err := time.Parse(time.RFC1123, lastmod)
	catch(err)
	return t.Format("2006-01-02T15:04:05.9+00.00")
}

// Encode data to XML format and write data to file "fn"
func encodeAndWriteToXML(fn string, data interface{}) {
	fmt.Println("Writing to file", fn)
	writer, err := os.Create(fn)
	catch(err)
	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "	")
	catch(encoder.Encode(data))
}

// Append string if missing from the slice and return slice
func appendIfMissing(slice []string, str string) []string {
	contained, _ := contains(slice, str)
	if !contained {
		slice = append(slice, str)
	}
	return slice
}

// Check if string is in the given slice and return true and
// index location otherwise false, -1
func contains(slice []string, match string) (bool, int) {
	for i, item := range slice {
		if item == match {
			return true, i
		}
	}
	return false, -1
}

var protocolRegExp = regexp.MustCompile("^https?://")

func readArguments() (string, string, int, bool) {
	urlFlag := flag.String("url", "", "The URL of the site you want to map.")
	fileNameFlag := flag.String("file", "sitemap.xml", "The name of the file you want to write to.")
	depthFlag := flag.Int("depth", -1, "Max depth of recursive scraping. Set -1 for no limit.")
	includeQueryFlag := flag.Bool("query", true, "Include routes with unique query parameters in indexing.")

	flag.Parse()

	if *urlFlag == "" {
		panic("an argument for url is required!")
	} else if !protocolRegExp.MatchString(*urlFlag) {
		*urlFlag = "http://" + *urlFlag
	}
	if *fileNameFlag == "" {
		panic("an argument for filename is required!")
	}

	return *urlFlag, *fileNameFlag, *depthFlag, *includeQueryFlag
}

// Check to see if err == nil
// and panic(err) if true
func catch(err error) {
	if err != nil {
		panic(err)
	}
}
