package main

import (
	"fmt"
	"os"
	"time"

	"github.com/theckman/yacspin"
)

type SpinnerUpdater struct {
	S       *yacspin.Spinner
	Success int32
	Failed  int32
	Output  string
}

func CreateSpinnerFromMethods() *yacspin.Spinner {
	cfg := yacspin.Config{
		Frequency:         100 * time.Millisecond,
		CharSet:           yacspin.CharSets[40],
		SuffixAutoColon:   true,
		Message:           "Initializing Download",
		StopMessage:       "Done downloading subtitle",
		StopCharacter:     "✓",
		StopColors:        []string{"fgGreen"},
		StopFailCharacter: "✗",
		StopFailMessage:   "Failed while downloading subtitle",
		StopFailColors:    []string{"fgRed"},
	}

	s, err := yacspin.New(cfg)
	if err != nil {
		exitf("failed to generate spinner from methods: %v", err)
	}

	if err := s.CharSet(yacspin.CharSets[11]); err != nil {
		exitf("failed to set charset: %v", err)
	}

	if err := s.Colors("fgYellow"); err != nil {
		exitf("failed to set color: %v", err)
	}

	if err := s.StopColors("fgGreen"); err != nil {
		exitf("failed to set stop colors: %v", err)
	}

	if err := s.StopFailColors("fgRed"); err != nil {
		exitf("failed to set stop fail colors: %v", err)
	}

	s.Suffix(" ")
	s.StopCharacter("✓")
	s.StopMessage("done")
	s.StopFailCharacter("✗")
	s.StopFailMessage("failed")

	return s
}

func exitf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
	os.Exit(1)
}
