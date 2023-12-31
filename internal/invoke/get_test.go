package invoke

import (
	"context"
	"fmt"
	"github.com/dark-enstein/scour/internal/parser"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	testGetUrls = []string{
		"http://eu.httpbin.org/get",
		"https://eu.httpbin.org:443/get",
	}
)

func TestGet(t *testing.T) {
	for i := 0; i < len(testGetUrls); i++ {
		fmt.Printf(Order, i+1)
		ctx := context.WithValue(context.Background(), parser.KeyV, false)
		u, _ := parser.NewUrl(ctx, testGetUrls[i])
		_, actual := Get(ctx, u)
		assert.NotEmpty(t, actual)
	}
}
