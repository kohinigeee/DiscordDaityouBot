package botmanager

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/kohinigeee/DiscordDaityouBot/mylogger"
)

const (
	PingPoingSelectUsersID = "select_users"
)

func pingPong(s *discordgo.Session, i *discordgo.InteractionCreate, manager *BotManager) {
	logger := mylogger.L()

	threadID := i.ChannelID
	logger.Debug("pingPong started",
		slog.String("ChannelID", threadID),
	)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "pong",
		},
	})

	logger.Debug("pingPong BotUser",
		slog.String("BotUser", fmt.Sprintf("%+v", manager.BotUserInfo)),
		slog.String("BotUser Abar URL ", manager.BotUserInfo.AvatarURL("20")),
	)

	logger.Debug("pingPong finished")
}

// func pingPong(s *discordgo.Session, i *discordgo.InteractionCreate, manager *BotManager) {
// 	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
// 		Type: discordgo.InteractionResponseChannelMessageWithSource,
// 		Data: &discordgo.InteractionResponseData{
// 			Content: "pong",
// 		},
// 	})
// }
