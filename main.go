package main

import (
	"context"
	"fmt"
	"github.com/dark-enstein/scour/internal/config"
	"github.com/dark-enstein/scour/internal/invoke"
	"github.com/dark-enstein/scour/internal/invoke/httpoke"
	"github.com/dark-enstein/scour/internal/invoke/socket"
	"github.com/dark-enstein/scour/internal/parser"
	"github.com/dark-enstein/scour/internal/parser/httparser"
	"github.com/dark-enstein/scour/internal/parser/socketparser"
	"github.com/spf13/pflag"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	SOCKET_TEST = iota + 1
	HTTP_TEST
)

var (
	// FLGS holds the flags values for every iteration
	FLGS = config.NewFlags()
	// ScourASCII holds the header output of Scour. TODO: This should be refactored to using go:embed via text files
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
	// ParsedUrlOutput holds the template for parsing url information in verbose mode. TODO: This should be refactored to using go:embed via text files
	ParsedUrlOutput = `
connecting to %s
*   Trying %s...
* Connected to %s (%s) port %s
> %s /%s %s/1.1
> Host: %s
> Accept: */
`
	// InvokeOutput returns the metadata from the response. Activated in verbose mode. TODO: This should be refactored to using go:embed via text files.
	InvokeOutput = `
< %s/1.1 %s
< Date: %s
< Content-Type: %s
< Content-Length: %s
< Connection: %s
< Server: %s
< Access-Control-Allow-Origin: %s
< Access-Control-Allow-Credentials: %v
`
)

func main() {
	// setDebug flag is used for toggling debug mode on or off
	var setDebug = false
	var help bool
	var out string

	// control flow for when Goland IDE is running in debug mode or not
	if setDebug {
		FLGS = debug(SOCKET_TEST, FLGS)
		fmt.Printf(ScourASCII, FLGS.Verbose)
		err := FLGS.ValidateAll()
		if err != nil {
			log.Println(fmt.Errorf("errors encountered while validating flags: %w\n%s", err, config.Help))
			pflag.PrintDefaults()
			os.Exit(1)
		}
		help, out = _main([]string{"t.sock http:/images/json"})
	} else {
		if err := initFlags(); err != nil {
			log.Println(fmt.Errorf("errors encountered while validating flags: %w\n%s", err, config.Help))
			pflag.PrintDefaults()
			os.Exit(1)
		}
		fmt.Printf(ScourASCII, FLGS.Verbose)
		fmt.Println("all args:", pflag.Args())
		if pflag.NArg() > 0 {
			// TODO: implement a smoother _main() function that properly handles the argslist as they are, without needing it as a string to process it
			if pflag.NArg() == 1 {
				help, out = _main(pflag.Args())
			} else {
				help, out = _main([]string{fmt.Sprintf("%s %s", pflag.Args()[0], pflag.Args()[1])})
			}
		} else {
			help, out = _main([]string{""})
		}
	}
	if help {
		pflag.PrintDefaults()
	}
	fmt.Println(out)
}

// initFlags parses in cmdline flags, and does validation on them
func initFlags() error {
	// TODO: Switch to using cobra or some more robust cli framework
	pflag.BoolVarP(&FLGS.Verbose, "verbose", "v", false, "Turn on/off debug mode.")
	pflag.StringVarP(&FLGS.Method, "X", "X", http.MethodGet, "Set request method.")
	pflag.StringVarP(&FLGS.Data, "data", "d", "", "Pass request data.")
	pflag.StringVarP(&FLGS.Headers, "Header", "H", "", "Pass in custom request headers.")
	//pflag.BoolVarP(&FLGS.UnixSocket, "abstract-unix-socket", "aus", false, "(HTTP) Connect through an abstract Unix domain socket, instead of using the network. Note: netstat shows the path of an abstract socket prefixed with '@', however the <path> argument should not have this leading character.\nIf --abstract-unix-socket is provided several times, the last set value is used.\n")
	pflag.BoolVarP(&FLGS.UnixSocket, "unix-socket", "u", false, "(HTTP) Connect through this Unix domain socket, instead of using the network.\nIf --unix-socket is provided several times, the last set value is used.")
	pflag.BoolVarP(&FLGS.InteractiveMode, "it", "i", false, "Toggles console mode for socket connection. Only supported when using '--abstract-unix-socket'.")
	pflag.StringVarP(&FLGS.SocketLoc, "create-socket", "c", "", "Creates a socket at the specified path")
	pflag.Parse()
	return FLGS.ValidateAll()
}

// _main is the lower level main function
func _main(args []string) (help bool, output string) {
	if len(args) == 0 {
		if FLGS.Method == config.MethodSocket {
			log.Println("Please pass at least one argument in the format: scour [--X|--v] <socket-path> <url>")
		} else {
			log.Println("Please pass at least one argument in the format: scour [--X|--v] <url>")
		}
		return true, ""
	} else if len(args) > 1 {
		if FLGS.Method == config.MethodSocket {
			log.Println("Please pass at least one argument in the format: scour [--X|--v] <socket-path> <url>")
		} else {
			log.Println("Too many arguments passed in. Only one argument required: scour [--X|--v] <url>")
		}
		return true, ""
	}
	instanceCtx := context.WithValue(context.Background(), httparser.KeyV, FLGS.Verbose)

	if len(FLGS.SocketLoc) > 1 {
		if err := socket.CreateSocketSubProc(FLGS.SocketLoc); err != nil {
			log.Println(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	url, err := parseUrl(instanceCtx, args[0], FLGS)
	if err != nil {
		log.Println(err)
		return false, ""
	}
	var headers *invoke.RespHeaders
	var resp []byte

	switch FLGS.Method {
	case http.MethodGet:
		headers, resp, err = httpoke.Get(instanceCtx, url)
	case http.MethodPost:
		headers, resp, err = httpoke.Post(instanceCtx, url, []byte(FLGS.Data))
	case http.MethodDelete:
		headers, resp, err = httpoke.Delete(instanceCtx, url)
	case http.MethodPut:
		headers, resp, err = httpoke.Put(instanceCtx, url, []byte(FLGS.Data))
	case http.MethodPatch:
		headers, resp, err = httpoke.Patch(instanceCtx, url, []byte(FLGS.Data))
	case config.MethodSocket:
		resp, err = socket.UnixSock(instanceCtx, url, FLGS.InteractiveMode)
	}

	if FLGS.Verbose {
		output += fmt.Sprintf(ParsedUrlOutput, url.Host(), url.Host(), url.Host(), url.Host(), url.Port(), strings.ToUpper(url.Path()), url.Path(), url.Protocol().MustUpper(), url.Host()) + "\n"
		output += fmt.Sprintf(InvokeOutput, headers.Protocol, headers.RespCode, headers.Date, headers.ContentType, headers.ContentLength, headers.Connection, headers.Server, headers.AccessControlAllowOrigin, headers.AccessControlAllowCredentials) + "\n"
	}
	output += string(resp)
	return
}

// parseUrl parses the right url from the request
func parseUrl(ctx context.Context, urlString string, flag *config.Flags) (url parser.Url, err error) {
	if len(urlString) < 1 {
		return nil, fmt.Errorf("url string empty")
	}
	switch flag.Resolve() {
	case config.MODE_HTTP:
		httpurl, _ := httparser.NewUrl(ctx, urlString)
		url = httpurl
	case config.MODE_SOCKET:
		socketUrl := socketparser.NewSocket(ctx, urlString)
		url = socketUrl
	}

	if url.Err() != nil {
		return nil, err
	}
	return url, nil
}

// debug sets some default flag values for Goland debugging
func debug(debugType int, flag *config.Flags) *config.Flags {
	switch debugType {
	case SOCKET_TEST:
		flag = &config.Flags{
			Verbose:         true,
			Method:          http.MethodGet,
			Data:            "",
			Headers:         "",
			UnixSocket:      true,
			InteractiveMode: false,
		}
	case HTTP_TEST:
		flag = &config.Flags{
			Verbose:         true,
			Method:          "GET",
			Data:            "",
			Headers:         "accept: application/json",
			UnixSocket:      false,
			InteractiveMode: false,
		}
	}
	return flag
}
