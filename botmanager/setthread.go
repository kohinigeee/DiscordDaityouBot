package botmanager

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/kohinigeee/DiscordDaityouBot/mylogger"
)

const (
	SetThreadHandlerID = "set_thread_handler"
)

func SetThreadHandler(s *discordgo.Session, i *discordgo.InteractionCreate, manager *BotManager) {
	logger := mylogger.L()
	const commandName = "Set Thread"

	threadID := i.ChannelID
	logger.Debug("SetThreadHandler started",
		slog.String("ChannelID", threadID),
	)

	logger.Debug("Set Thread Infos",
		slog.String("threadID", threadID),
	)

	manager.RestoreThreadID = threadID

	embed := manager.MakeNormalMessageEmbed(commandName, "スレッドを設定したぶぅ", nil)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})

	logger.Debug("SetThreadHandler finished")
}
