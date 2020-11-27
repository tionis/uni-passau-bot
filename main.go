package main

import "os"

func main() {
	// Start UniPassauBot with environment variable
	UniPassauBot(os.Getenv("UNIPASSAUBOT_TOKEN"))
}
