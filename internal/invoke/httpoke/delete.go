package httpoke

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dark-enstein/scour/internal/invoke"
	"github.com/dark-enstein/scour/internal/parser"
	"github.com/dark-enstein/scour/internal/parser/httparser"
	"io"
	"log"
	"net/http"
	"time"
)

// Delete sends a DELETE HTTP request to the specified URL and returns the response headers and body.
// It uses a context with a timeout to ensure the request does not hang indefinitely.
// The function handles the creation of the request, sending it, and processing the response.
// It returns the response headers and the response body as a byte slice.
// If verbose logging is enabled in the context, it logs various stages of the request and response process.
func Delete(ctx context.Context, url parser.Url) (*invoke.RespHeaders, []byte, error) {
	_ = &invoke.RespHeaders{}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), bytes.NewBuffer([]byte{}))
	if err != nil {
		log.Printf("Error creating request object: %s\n", err.Error())
	}

	cli := http.Client{}

	t1 := time.Now()

	resp, err := cli.Do(req)
	if err != nil {
		log.Printf("DELETE request failed with: %s\n", err.Error())
		return nil, nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	tDur := time.Since(t1)

	respH := invoke.NewHeaders(resp.Status, fmt.Sprintf("%s/1.1", url.Protocol().String()), resp.Header.Get("Date"), resp.Header.Get("Content-Type"), resp.Header.Get("Content-Length"), resp.Header.Get("Connection"), resp.Header.Get("Server"), resp.Header.Get("Access-Control-Allow-Origin"), resp.Header.Get("Access-Control-Allow-Credentials"))
	// Logging response headers if verbose logging is enabled...
	if httparser.ParseLogLevelFromCtx(ctx, httparser.KeyV) == true {
		log.Printf("Response: %v\n", respH)
	}

	responseStream, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error receiving response:", err.Error())
		return nil, nil, err
	}
	if httparser.ParseLogLevelFromCtx(ctx, httparser.KeyV) == true {
		log.Printf("Buffer length: %d\n", len(responseStream))
	}
	if err != nil {
		log.Printf("Error encountered decoding response body: %s\n", err.Error())
		return nil, nil, err
	}
	if httparser.ParseLogLevelFromCtx(ctx, httparser.KeyV) == true {
		log.Printf("Time taken: %s\n", tDur.String())
	}
	return respH, responseStream, nil
}
