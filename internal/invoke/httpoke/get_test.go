package httpoke

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dark-enstein/scour/internal/parser"
	"github.com/dark-enstein/scour/internal/parser/httparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"os/exec"
	"testing"
)

var (
	testGetUrls = []string{
		"http://eu.httpbin.org/get",
		"https://eu.httpbin.org:443/get",
	}
	Order = `
-------------------------------
Test Order: %d
----------`
)

// TestHttpGet_Response tests Get response against sample urls, while removing variablility using TestResponse from their response inorder to compare and test that they match
func TestHttpGet_Response(t *testing.T) {
	verbose := false

	for i := 0; i < len(testGetUrls); i++ {
		fmt.Printf(Order, i+1)
		ctx := context.WithValue(context.Background(), httparser.KeyV, verbose)

		var url parser.Url

		curlArgs := []string{testGetUrls[i], "-s", "-v"}
		fmt.Println("curl args:", curlArgs)
		curlStdOutput, _, err := curl(curlArgs)
		if err != nil {
			log.Println("error executing command:", err.Error())
		}
		fmt.Printf("curl stdoutput: %s\n", curlStdOutput)

		url, err = httparser.NewUrl(ctx, testGetUrls[i])
		assert.NoError(t, err)
		_, respBytes, err := Get(ctx, url)
		assert.NoError(t, err)
		fmt.Printf("respOutput: %s", respBytes)
		require.Equal(t, DecodeAndClean(curlStdOutput), DecodeAndClean(respBytes), "bytes output not equal")
	}
}

// curl runs the curl command with the provided arguments
func curl(argsStr []string) (stdout, stderr []byte, err error) {
	cmd := exec.Command("curl", argsStr...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err = cmd.Run()
	return stdoutBuf.Bytes(), stderrBuf.Bytes(), err
}

// TestResponse is the response from the server
type TestResponse struct {
	Args struct {
	} `json:"args"`
	Headers struct {
		Host string `json:"Host"`
	} `json:"headers"`
}

func DecodeAndClean(data []byte) []byte {
	if !json.Valid(data) {
		log.Println("json invalid", string(data))
		return nil
	}

	var resp = TestResponse{}
	err := json.Unmarshal(data, &resp)
	if err != nil {
		log.Println("cannot unmarshal json:", err.Error())
		return nil
	}

	b, _ := json.Marshal(resp)
	return b
}
