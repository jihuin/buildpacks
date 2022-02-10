// Package main is a mock process thats behavior can be configured by
// by setting certain environment variables.
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/GoogleCloudPlatform/buildpacks/internal/buildpacktestenv"
)

func main() {
	mocksJSON := os.Getenv(buildpacktestenv.EnvHelperMockProcessMap)
	if mocksJSON == "" {
		log.Fatalf("%q env var must be set", buildpacktestenv.EnvHelperMockProcessMap)
	}
	mockProcesses, err := buildpacktestenv.UnmarshalMockProcessMap(mocksJSON)
	if err != nil {
		log.Fatalf("unable to unmarshal mock process map from JSON '%s': %v", mocksJSON, err)
	}

	fullCommand := strings.Join(os.Args[1:], " ")
	var mockMatch *buildpacktestenv.MockProcess = nil
	for shortCommand, mock := range mockProcesses {
		if strings.Contains(fullCommand, shortCommand) {
			mockMatch = mock
			break
		}
	}
	if mockMatch == nil {
		// To avoid needing to mock every call to Exec, assume
		// the process should pass if it wasn't specified by the test.
		os.Exit(0)
	}

	for dest, src := range mockMatch.MovePaths {
		os.Rename(src, dest)
	}

	if mockMatch.Stdout != "" {
		fmt.Fprint(os.Stdout, mockMatch.Stdout)
	}

	if mockMatch.Stderr != "" {
		fmt.Fprint(os.Stderr, mockMatch.Stderr)
	}

	os.Exit(mockMatch.ExitCode)
}
