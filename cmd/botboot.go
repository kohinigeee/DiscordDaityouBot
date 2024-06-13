package cmd

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/kohinigeee/DiscordDaityouBot/botmanager"
)

func BotBoot() {
	discordSess, err := discordgo.New(discordToken)
	if err != nil {
		logger.Error("Error creating Discord session", slog.String("err", err.Error()))
		return
	}

	manager := botmanager.NewBotManager(discordSess)
	discordSess.Open()
	defer func() {
		err := discordSess.Close()
		if err != nil {
			logger.Error("Error closing Discord session", slog.String("err", err.Error()))
		}
	}()

	manager.Start()

	logger.Info("Bot is now running. Press CTRL+C to exit.")
	stopBot := make(chan os.Signal, 1)
	signal.Notify(stopBot, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stopBot
}
