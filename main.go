package main

import (
	"flag"
	"log/slog"

	"github.com/kohinigeee/DiscordDaityouBot/cmd"
	"github.com/kohinigeee/DiscordDaityouBot/mylogger"
)

//必要権限
//READ Messages/View Channels
//Send Messages
//Create Public Threads
//Send Messages in Threads
//Manage Messages
//Manage Threads
//Embed Links
//Read Message History
//Mention Everyone
//Add Reactions
//Use Slash Commands
//Use Emebedded Activities

func main() {
	logger := mylogger.L()

	var mode string

	flag.StringVar(&mode, "mode", "boot", "boot or slashapply")
	flag.Parse()

	switch mode {
	case "boot":
		logger.Info("Starting bot boot mode")
		cmd.BotBoot()
	case "slashapply":
		logger.Info("Starting slash apply mode")
		cmd.SlashApply()
	default:
		logger.Error("Invalid mode name", slog.String("mode", mode))
		flag.Usage()
	}

}
