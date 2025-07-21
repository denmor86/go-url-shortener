package main

import "os"

func foo() {
	os.Exit(1) // разрешено, так как не в main
}

func main() {
	// нет вызова os.Exit
}
