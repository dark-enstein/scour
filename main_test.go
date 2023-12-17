package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"os/exec"
	"strings"
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

func _gocmd(s string) (string, int, string, error) {
	sArr := strings.Split(s, " ")
	cmd := exec.Command("/opt/homebrew/opt/go/libexec/bin/go", sArr...)
	output, err := cmd.Output()
	return cmd.String(), 0, string(output), err
}

func _cmd(s string, a ...string) (string, int, string, error) {
	cmd := exec.Command(s, a...)
	output, err := cmd.Output()
	return cmd.String(), 0, string(output), err
}

func _build() int {
	_, _, out, err := _cmd("cp", "main.go main_cp.go")
	if err != nil {
		log.Println("error with copying file:", err)
	}
	log.Println("copy output:", out)
	str, code, output, err := _gocmd("build main_cp.go -o scour_test")
	if err != nil {
		log.Printf("Command %s failed with error: %s\n", str, err)
		return code
	}
	log.Printf("Output: %s\n", output)
	return code
}

func TestBuild(t *testing.T) {
	code := _build()
	if code > 0 {
		_, err := os.Stat(testBin)
		if os.IsExist(err) {
			log.Println("bin exists despite build command says its unhealthy")
		}
	}
	_, err := os.Stat(testBin)
	if os.IsNotExist(err) {
		_, _, out, _ := _cmd("ls")
		fmt.Println(out)
		log.Println("bin does not exists despite build command successfull")
	}
}
