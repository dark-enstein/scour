package invoke

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dark-enstein/scour/internal/parser"
	"github.com/dark-enstein/scour/internal/parser/httparser"
	"io"
	"log"
	"net/http"
	"time"
)

// Put sends a PUT HTTP request to the specified URL with the provided data.
// It manages request timeouts using context, logs relevant information,
// and returns the response headers and body as a byte slice.
func Put(ctx context.Context, url parser.Url, data []byte) (*RespHeaders, []byte, error) {
	_ = &RespHeaders{}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url.String(), bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error creating request object: %s\n", err.Error())
	}

	cli := http.Client{}
	t1 := time.Now()
	resp, err := cli.Do(req)
	if err != nil {
		log.Printf("PUT request failed with: %s\n", err.Error())
		return nil, nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	tDur := time.Since(t1)
	//fmt.Println(resp)

	respH := newHeaders(resp.Status, fmt.Sprintf("%s/1.1", url.Protocol().String()), resp.Header.Get("Date"), resp.Header.Get("Content-Type"), resp.Header.Get("Content-Length"), resp.Header.Get("Connection"), resp.Header.Get("Server"), resp.Header.Get("Access-Control-Allow-Origin"), resp.Header.Get("Access-Control-Allow-Credentials"))
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
