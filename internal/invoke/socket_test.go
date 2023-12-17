package invoke

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"net"
	"os"
	"testing"
)

type SocketTestSuite struct {
	suite.Suite
}

var (
	testSockets = map[string]bool{
		"file1.txt":      false,
		"socket1.sock":   true,
		"image1.png":     false,
		"socket2.sock":   true,
		"document1.docx": false,
		"archive.tar.gz": false,
		"script.sh":      false,
		"socket3.sock":   true,
		"data.csv":       false,
		"note.md":        false,
		"config.conf":    false,
		"backup.zip":     false,
	}
	Order = `
-------------------------------
Test Order: %d
-------------------------------
`
	TESTDIR = "./sock"
)

func (suite *SocketTestSuite) SetupTest() {
	err := os.MkdirAll(TESTDIR, 0777)
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *SocketTestSuite) TestIsSocket() {
	i := 1
	for k, v := range testSockets {
		if v {
			l, err := net.Listen("unix", k)
			if err != nil {
				log.Fatal("listen error:", err)
			}
			fileInfo, _ := os.Stat(k)
			log.Printf("FilePath: %s\n", k)
			log.Printf("Filemode: %s\n", fileInfo.Mode().Type())
			what := IsSocket(k)
			assert.Equal(suite.T(), v, what)
			defer func(l net.Listener) {
				_ = l.Close()
			}(l)
		} else {
			fmt.Printf(Order, i)
			tmpFile, err := os.Create(k)
			if err != nil {
				log.Fatalf("Error creating file: %s\n", err.Error())
			}
			filePath := tmpFile.Name()
			fileInfo, _ := os.Stat(filePath)
			log.Printf("FilePath: %s\n", filePath)
			log.Printf("Filemode: %s\n", fileInfo.Mode().Type())

			what := IsSocket(filePath)
			assert.Equal(suite.T(), v, what)
			// Close and remove the temp file, so we can use its name for the socket
			_ = tmpFile.Close()
			_ = os.Remove(filePath)
		}
		i++
	}
}

func clean() {
	err := os.RemoveAll(TESTDIR)
	if err != nil {
		log.Printf("failed to delete test dir: %s\n", TESTDIR)
	}
}

func (suite *SocketTestSuite) TearDownTest() {
	clean()
}

func TestSocketSuite(t *testing.T) {
	suite.Run(t, new(SocketTestSuite))
}
