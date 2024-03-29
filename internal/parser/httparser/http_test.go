package httparser

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	// testUrls is a map containing test cases for the URL parser.
	// Each entry maps a raw URL string to an expected HTTP struct.
	testUrls = map[string]HTTP{
		// Test cases with various URL formats and expected outcomes.
		// Includes both valid URLs and URLs expected to produce errors (like IPv6 addresses).
		"http://eu.httpbin.org/get": {
			rawString: "http://eu.httpbin.org/get",
			protocol:  "http",
			host:      "eu.httpbin.org",
			port:      "80",
			path:      "get",
			err:       nil,
		},
		"http://eu.httpbin.org:4040/get": {
			rawString: "http://eu.httpbin.org:4040/get",
			protocol:  "http",
			host:      "eu.httpbin.org",
			port:      "4040",
			path:      "get",
			err:       nil,
		},
		"https://eu.httpbin.org/get": {
			rawString: "https://eu.httpbin.org/get",
			protocol:  "https",
			host:      "eu.httpbin.org",
			port:      "443",
			path:      "get",
			err:       nil,
		},
		"http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]/": {
			rawString: "http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]/",
			protocol:  "",
			host:      "",
			port:      "",
			path:      "",
			err:       ErrIPv6tError,
		},
		"http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:8080/": {
			rawString: "http://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:8080/",
			protocol:  "",
			host:      "",
			port:      "",
			path:      "",
			err:       ErrIPv6tError,
		},
	}
)

// TestParser is a test function for the NewUrl function in the parser package.
// It iterates over the testUrls, parsing each one and comparing the result to the expected HTTP struct.
func TestParser(t *testing.T) {
	i := 1
	for url, expected := range testUrls {
		ctx := context.WithValue(context.Background(), KeyV, true)
		fmt.Printf("Order: %d\n", i)
		actual, _ := NewUrl(ctx, url)
		assert.Equal(t, expected, *actual)
		i++
	}
}

//func TestParser(t *testing.T) {
//	assert := assert.New(t)
//
//	// assert equality
//	assert.Equal(123, 123, "they should be equal")
//
//	// assert inequality
//	assert.NotEqual(123, 456, "they should not be equal")
//
//	// assert for nil (good for errors)
//	assert.Nil(object)
//
//	// assert for not nil (good when you expect something)
//	if assert.NotNil(object) {
//
//		// now we know that object isn't nil, we are safe to make
//		// further assertions without causing any errors
//		assert.Equal("Something", object.Value)
//	}
//}
