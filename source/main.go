package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"openriot/commands"
)

// Injected at build time via Makefile ldflags:
//
//	-X main.version=$(OPENRIOT_VERSION)
//	-X main.openbsdVersion=$(OPENBSD_VERSION)
//
// Do NOT hardcode these here - change Makefile instead.
var version = "dev"
var openbsdVersion = "7.9"

var testMode bool

// logDebugCall logs each binary invocation to /tmp/openriot_calls.log
func logDebugCall() {
	if os.Getenv("OPENRIOT_DEBUG") != "1" {
		return
	}
	f, err := os.OpenFile("/tmp/openriot_calls.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: cannot open log: %v\n", err)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s %s\n", time.Now().Format("15:04:05.000"), strings.Join(os.Args[1:], " "))
}

func main() {
	logDebugCall()

	// Handle --test/-t flag first (affects other commands)
	for _, arg := range os.Args[1:] {
		if arg == "--test" || arg == "-t" {
			testMode = true
		}
	}

	registry := commands.NewRegistry()
	commands.RegisterAll(registry, &testMode)

	if len(os.Args) >= 2 {
		cmdName := os.Args[1]
		if cmdName == "--version" {
			fmt.Println("openriot", version)
			os.Exit(0)
		}
		var args []string
		if len(os.Args) > 2 {
			args = os.Args[2:]
		}
		code := registry.Dispatch(cmdName, args)
		if code >= 0 {
			os.Exit(code)
		}
	}

	fmt.Fprint(os.Stderr, registry.Usage(version))
	os.Exit(1)
}
