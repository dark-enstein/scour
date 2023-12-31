package invoke

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dark-enstein/scour/internal/parser"
	"io"
	"log"
	"net/http"
	"time"
)

func Delete(ctx context.Context, url *parser.HTTP) (*RespHeaders, []byte) {
	_ = &RespHeaders{}
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
		return nil, nil
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	tDur := time.Since(t1)

	respH := newHeaders(resp.Status, fmt.Sprintf("%s/1.1", url.Protocol().String()), resp.Header.Get("Date"), resp.Header.Get("Content-Type"), resp.Header.Get("Content-Length"), resp.Header.Get("Connection"), resp.Header.Get("Server"), resp.Header.Get("Access-Control-Allow-Origin"), resp.Header.Get("Access-Control-Allow-Credentials"))
	if parser.ParseLogLevelFromCtx(ctx, parser.KeyV) == true {
		log.Printf("Response: %v\n", respH)
	}

	var buf bytes.Buffer
	tmp := make([]byte, READ_PAGESIZE)
	for {
		n, err := resp.Body.Read(tmp)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error receiving response:", err.Error())
				buf.Write(tmp[:n])
			}
			break
		}
		buf.Write(tmp[:n])
	}
	if parser.ParseLogLevelFromCtx(ctx, parser.KeyV) == true {
		log.Printf("Buffer length: %d\n", buf.Len())
	}
	if err != nil {
		log.Printf("Error encountered decoding response body: %s\n", err.Error())
	}
	if parser.ParseLogLevelFromCtx(ctx, parser.KeyV) == true {
		log.Printf("Time taken: %s\n", tDur.String())
	}
	return respH, buf.Bytes()
}
