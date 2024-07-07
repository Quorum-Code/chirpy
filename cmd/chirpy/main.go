package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/Quorum-Code/chirpy/internal/webserver"
)

type RunConfig struct {
	Debug bool
	Test  bool
}

var HelpText = `
'q' -> Quit: close the program
'h' -> Help: prints out some helpful text
`

func main() {
	// Pass Server Config
	svrcfg := webserver.ServerConfig{
		IsDebug:   *flag.Bool("debug", false, "Enable debug mode"),
		IsTesting: *flag.Bool("testing", false, "Enable for testing"),
	}

	// Start server
	_ = webserver.StartServer(svrcfg)

	// Start cli reader
	readCommandLine()
}

func readCommandLine() {
	// Process input
	scanner := bufio.NewScanner(os.Stdin)
	ok := true
	for ok {
		ok = parseScanner(scanner)
	}
}

func parseScanner(s *bufio.Scanner) bool {
	if !s.Scan() {
		if s.Err() != nil {
			fmt.Printf("ERROR: %e\n", s.Err())
		} else {
			fmt.Println("reached end of input, closing server")
		}
		return false
	}

	text := s.Text()
	fmt.Printf("text: %s\n", text)

	switch text {
	case "h":
		fmt.Println(HelpText)
		return true
	case "q":
		fmt.Println("closing server")
		return false
	default:
		fmt.Println("command not recognized, use 'h' for help")
		return true
	}
}
