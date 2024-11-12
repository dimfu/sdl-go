package main

import "fmt"

func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("  [flags] <command>")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --lang=<country alpha 2 code>   Override preferred language to use.")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  run          - Executes the main program.")
	fmt.Println("  config list  - See current config.")
	fmt.Println("  config reset - To reset the config.")
	fmt.Println("  help         - Displays this help message.")
}
