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
	"os"
)

var (
	FLGS       = config.NewFlags()
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

func main() {
	if err := initFlags(); err != nil {
		log.Println(fmt.Errorf("errors encountered while validating flags: %w\n%s", err, config.Help))
		os.Exit(1)
	}
	fmt.Printf(ScourASCII, FLGS.Verbose)
	def, out := _main(pflag.Args())
	if def {
		pflag.PrintDefaults()
	}
	fmt.Println(out)
}

func initFlags() error {
	pflag.BoolVarP(&FLGS.Verbose, "verbose", "v", false, "Turn on/off debug mode.")
	pflag.StringVarP(&FLGS.Method, "X", "X", http.MethodGet, "Set request method.")
	pflag.StringVarP(&FLGS.Data, "data", "d", "", "Pass request data.")
	pflag.StringVarP(&FLGS.Headers, "Header", "H", "", "Pass in custom request headers.")
	pflag.BoolVarP(&FLGS.UnixSocket, "abstract-unix-socket", "aus", false, "(HTTP) Connect through an abstract Unix domain socket, instead of using the network. Note: netstat shows the path of an abstract socket prefixed with '@', however the <path> argument should not have this leading character.\nIf --abstract-unix-socket is provided several times, the last set value is used.\n")
	pflag.BoolVarP(&FLGS.UnixSocket, "unix-socket", "us", false, "(HTTP) Connect through this Unix domain socket, instead of using the network.\nIf --unix-socket is provided several times, the last set value is used.")
	pflag.BoolVarP(&FLGS.InteractiveMode, "it", "it", false, "Toggles console mode for socket connection. Only supported when using '--abstract-unix-socket'.")
	pflag.Parse()
	return FLGS.ValidateAll()
}

func _main(args []string) (help bool, output string) {
	if len(args) == 0 {
		log.Println("Please pass at least one argument in the format: scour [--X|--v] <url>")
		return true, ""
	}
	url := args[0]
	instanceCtx := context.WithValue(context.Background(), parser.KeyV, FLGS.Verbose)
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
		headers, resp = invoke.Post(instanceCtx, p, []byte(FLGS.Data))
	case http.MethodDelete:
		headers, resp = invoke.Delete(instanceCtx, p)
	case http.MethodPut:
		headers, resp = invoke.Put(instanceCtx, p, []byte(FLGS.Data))
	case config.MethodSocket:
		headers, resp = invoke.UnixSock(instanceCtx, p, []byte(FLGS.Data))
	}

	if FLGS.Verbose {
		output += fmt.Sprintf(ParsedUrlOutput, p.Host(), p.Path(), p.Protocol().MustUpper(), p.Host()) + "\n"
		output += fmt.Sprintf(InvokeOutput, headers.Protocol, headers.RespCode, headers.Date, headers.ContentType, headers.ContentLength, headers.Connection, headers.Server, headers.AccessControlAllowOrigin, headers.AccessControlAllowCredentials) + "\n"
	}
	output += string(resp)
	return
}
