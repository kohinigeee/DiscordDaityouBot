package botmanager

import (
	"github.com/kohinigeee/DiscordDaityouBot/slashapi"
)

type SlashCommand struct {
	Command slashapi.SlashCommandJson
	Handler DiscorBotdHandler
}

type InteractCommand struct {
	Name    string
	Handler DiscorBotdHandler
}

type DiscordModalCommand struct {
	Name    string
	Handler DiscorBotdHandler
}

var (
	SlashCommands        []SlashCommand
	InteractCommands     []InteractCommand
	DiscordModalCommands []DiscordModalCommand
)

func init() {

	SlashCommands = []SlashCommand{
		{
			Command: slashapi.SlashCommandJson{
				Name:        "ping",
				Description: "ping pong",
			},
			Handler: pingPong,
		},
		{
			Command: slashapi.SlashCommandJson{
				Name:        "easypay",
				Description: "お金の支払いを記録します",
			},
			Handler: EasyPayHandler,
		},
		{
			Command: slashapi.SlashCommandJson{
				Name:        "setthread",
				Description: "履歴を投稿するスレッドを設定します",
			},
			Handler: SetThreadHandler,
		},
		{
			Command: slashapi.SlashCommandJson{
				Name:        "debt",
				Description: "お金の状況を表示します",
			},
			Handler: DebtHandler,
		},
		{
			Command: slashapi.SlashCommandJson{
				Name:        "reset",
				Description: "スレッドの履歴を参照して賃借を再計算します",
			},
			Handler: ResetHandler,
		},
	}

	InteractCommands = []InteractCommand{
		{
			Name:    EasyPaySelectUsersHandlerID,
			Handler: EasyPaySelectUsersHandler,
		},
	}

	DiscordModalCommands = []DiscordModalCommand{
		{
			Name:    EasyPayModalName,
			Handler: EasyPayModalHandler,
		},
	}

}

func InitialSlashCommands() []SlashCommand {
	return SlashCommands
}

func InitialInteractCommands() []InteractCommand {
	return InteractCommands
}

func InitialDiscordModalCommands() []DiscordModalCommand {
	return DiscordModalCommands
}
