package main

import (
	"os"

	UniPassauBot "github.com/tionis/uni-passau-bot/api"
)

func main() {
	// Start UniPassauBot with environment variable
	UniPassauBot.UniPassauBot(os.Getenv("UNIPASSAUBOT_TOKEN"))
}
