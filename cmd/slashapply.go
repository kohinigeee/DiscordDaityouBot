package cmd

import (
	"log/slog"
	"time"

	"github.com/kohinigeee/DiscordDaityouBot/botmanager"
	"github.com/kohinigeee/DiscordDaityouBot/mylogger"
	"github.com/kohinigeee/DiscordDaityouBot/slashapi"
)

func SlashApply() {

	slashapi.InitEnvs()

	commands := botmanager.InitialSlashCommands()

	for _, command := range commands {
		err := slashapi.ApplySlashCommand(&command.Command)
		if err != nil {
			mylogger.L().Error("Error applying slash command", slog.String("err", err.Error()))
		}
		time.Sleep(2 * time.Second)
	}
}
