package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testBin = "scour_test"
)

var (
	scopeURLs = map[string]string{
		"http://eu.httpbin.org/get": `
connecting to eu.httpbin.org
`,
		"http://eu.httpbin.org:443/get": `
connecting to eu.httpbin.org
`,
		"https://eu.httpbin.org/get": `
connecting to eu.httpbin.org
`,
		"http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]/":      "",
		"http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:8080/": "",
	}
	Order = `
-------------------------------
Test Order: %d
-------------------------------
`
	VERBOSE = true
)

func TestScour(t *testing.T) {
	//code := _build()
	//if code > 0 {
	//	log.Printf("Error occured while building binary\n")
	//}
	i := 1
	for url, expected := range scopeURLs {
		fmt.Printf(Order, i)
		_, actual := _main([]string{url})
		fmt.Println(actual)
		if VERBOSE {
			assert.Contains(t, actual, expected)
		}
		i++
	}
}

func TestPrettyPrint(t *testing.T) {}
