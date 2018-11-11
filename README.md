# Sitemapper

A simple console application to generate XML sitemaps written in Go.

## Build

```bash
go build
```

## Instructions

The application can be run from a `go` build or by simply running `go *.go`. The following options are included:
```bash
  -url string
        The URL of the site you want to map.
  -file string
        The name of the file you want to write to. (default "sitemap.xml")
  -query bool
        Include routes with unique query parameters in indexing. (default true)
  -depth int
        Max depth of recursive scraping. Set -1 for no limit. (default -1)
```

For example, `go *go -url=http://example.com -file=mysitemap.xml -query=false -depth=3` will write a sitemap file for "http://example.com" called "mysitemap.xml", excluding unique paths with query parameters and only traversing pages three levels deep. That is to say if page A points to B points to C points to D, then page D will be excluded from sitemap (conditional that another page with two or less levels deep does not also point to it).

## Priority

The application includes a naive implementation of a `priority` assignment where each page is assigned one point upon first discovery, but awarded decreasing fractions of a point for each additional discovery. See the `updatePoints` function in `points.go` for the implementation.