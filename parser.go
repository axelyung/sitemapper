package main

import (
	"strings"

	"golang.org/x/net/html"
)

// Given an html string find all anchor tags and
// return contents of their href tags. Removes
// duplicate entries.
func parse(htmlStr string) []string {
	reader := strings.NewReader(htmlStr)

	doc, _ := html.Parse(reader)

	nodes := findAnchorTags(doc)
	links := []string{}
	for _, node := range nodes {
		href := findHrefAttribute(node)
		if href != "" {
			links = appendIfMissing(links, href)
		}
	}
	return links
}

func findAnchorTags(n *html.Node) []*html.Node {
	if n.Type == html.ElementNode && n.Data == "a" {
		return []*html.Node{n}
	}
	ret := []*html.Node{}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ret = append(ret, findAnchorTags(c)...)
	}
	return ret
}

func findHrefAttribute(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			return attr.Val
		}
	}
	return ""
}
