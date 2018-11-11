package main

import "math"

// Normalizes point scores to range [0,1]
// and round to two decimal places.
func normalizePoints() {
	// find max of all points
	max := float64(0)
	points.Range(func(key interface{}, value interface{}) bool {
		p := value.(float64)
		if p > max {
			max = p
		}
		return true
	})

	points.Range(func(key interface{}, value interface{}) bool {
		p := value.(float64)
		normalizedValue := math.Round(100*p/max) / 100
		points.Store(key, normalizedValue)
		return true
	})
}

// Given a slice of found urls, increment point scores for
// each page
func updatePoints(urls []string) {
	for _, url := range urls {
		value, loaded := points.LoadOrStore(url, float64(1))
		if loaded {
			curr := value.(float64)
			// this is a naive implementation of decreasing
			// returns for additional matches
			next := curr + 1/(10*curr)
			points.Store(url, next)
		}
	}
}
