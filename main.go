package main

import (
	"context"
	"os"
	"log"
	"log/slog"
	UniPassauBot "github.com/tionis/uni-passau-bot/api"
)

func main() {
	// Start UniPassauBot with environment variable
	// TODO add line numbers to logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Enabled(context.Background(), slog.LevelDebug)

	UniPassauBot.UniPassauBot(logger.WithGroup("UniPassauBot"),os.Getenv("UNIPASSAUBOT_TOKEN"))
}
