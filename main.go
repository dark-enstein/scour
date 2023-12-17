package main

import (
	"context"
	"fmt"
	"github.com/dark-enstein/scour/internal/config"
	"github.com/dark-enstein/scour/internal/invoke"
	"github.com/dark-enstein/scour/internal/parser"
	"github.com/spf13/pflag"
	"log"
	"net/http"
)

var (
	VERBOSE    = false
	METHOD     = http.MethodGet
	DATA       = ""
	CLIHEADERS = ""
	FLGS       = NewFlags()
	ScourASCII = `
 _______  _______  _______  __   __  ______   
|       ||       ||       ||  | |  ||    _ |  
|  _____||       ||   _   ||  | |  ||   | ||  
| |_____ |       ||  | |  ||  |_|  ||   |_||_ 
|_____  ||      _||  |_|  ||       ||    __  |
 _____| ||     |_ |       ||       ||   |  | |
|_______||_______||_______||_______||___|  |_|

Debug = %v
`
	ParsedUrlOutput = `
connecting to %s
> GET /%s %s/1.1
> Host: %s
> Accept: */
`
	InvokeOutput = `
< %s/1.1 %s
< Date: %s
< Content-Type: %s
< Content-Length: %s
< Connection: %s
< Server: %s
< Access-Control-Allow-Origin: %s
< Access-Control-Allow-Credentials: %s
`
)

func NewFlags() *config.Flags {
	return &config.Flags{}
}

func main() {
	initFlags()
	fmt.Printf(ScourASCII, VERBOSE)
	def, out := _main(pflag.Args())
	if def {
		pflag.PrintDefaults()
	}
	fmt.Println(out)
}

func initFlags() {
	pflag.BoolVarP(&FLGS.Verbose, "verbose", "v", false, "Turn on/off debug mode.")
	pflag.StringVarP(&FLGS.Method, "X", "X", http.MethodGet, "Set request method.")
	pflag.StringVarP(&FLGS.Data, "data", "d", "", "Pass request data.")
	pflag.StringVarP(&FLGS.Headers, "Header", "H", "", "Pass in custom request headers.")
	pflag.BoolVarP(&FLGS.UnixSocket, "abstract-unix-socket", "aus", false, "(HTTP) Connect through an abstract Unix domain socket, instead of using the network. Note: netstat shows the path of an abstract socket prefixed with '@', however the <path> argument should not have this leading character.")
	pflag.Parse()
}

func _main(args []string) (help bool, output string) {
	if len(args) == 0 {
		log.Println("Please pass at least one argument in the format: scour [--X|--v] <url>")
		return true, ""
	}
	url := args[0]
	instanceCtx := context.WithValue(context.Background(), parser.KeyV, VERBOSE)
	p, err := parser.NewUrl(instanceCtx, url)
	if err != nil {
		return false, ""
	}
	var headers *invoke.RespHeaders
	var resp []byte

	switch FLGS.Method {
	case http.MethodGet:
		headers, resp = invoke.Get(instanceCtx, p)
	case http.MethodPost:
		headers, resp = invoke.Post(instanceCtx, p, []byte(DATA))
	case http.MethodDelete:
		headers, resp = invoke.Delete(instanceCtx, p)
	case http.MethodPut:
		headers, resp = invoke.Put(instanceCtx, p, []byte(DATA))
	}

	if FLGS.Verbose {
		output += fmt.Sprintf(ParsedUrlOutput, p.Host(), p.Path(), p.UProtocol(), p.Host()) + "\n"
		output += fmt.Sprintf(InvokeOutput, headers.Protocol, headers.RespCode, headers.Date, headers.ContentType, headers.ContentLength, headers.Connection, headers.Server, headers.AccessControlAllowOrigin, headers.AccessControlAllowCredentials) + "\n"
	}
	output += string(resp)
	return
}
