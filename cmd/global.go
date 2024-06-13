package cmd

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/kohinigeee/DiscordDaityouBot/mylogger"
)

var (
	logger       *slog.Logger
	discordToken string
	appId        string
)

func init() {
	logger = mylogger.L()

	err := godotenv.Load()
	if err != nil {
		logger.Warn("Error loading .env file")
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "Info" {
		mylogger.SetLevel(slog.LevelInfo)
	}

	discordToken = "Bot " + os.Getenv("DISCORD_TOKEN")
	appId = os.Getenv("APPLICATION_ID")

	logger.Info("Imported env values", slog.String("DISCORD_TOKEN", discordToken), slog.String("APPLICATION_ID", appId))
}
