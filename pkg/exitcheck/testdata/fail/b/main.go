package main

import "os"

func main() {
	for _, v := range []string{"foo", "buzz", "exit", "go"} {
		if v == "exit" {
			os.Exit(1) // want "os.Exit call in main function of main package"
		}
	}
}
