package botmanager

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kohinigeee/DiscordDaityouBot/daityou"
	"github.com/kohinigeee/DiscordDaityouBot/lib"
	"github.com/kohinigeee/DiscordDaityouBot/mylogger"
)

const (
	DebtHandlerID = "debt_handler"
)

func formatWithCommna(num uint) string {
	in := fmt.Sprintf("%d", num)
	var out strings.Builder

	l := len(in)
	for i := 0; i < len(in); i++ {
		if i > 0 && (l-i)%3 == 0 {
			out.WriteByte(',')
		}
		out.WriteByte(in[i])
	}
	return out.String()
}

func DebtHandler(s *discordgo.Session, i *discordgo.InteractionCreate, manager *BotManager) {
	logger := mylogger.L()

	states := manager.GetAllStates(s, i)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	if err != nil {
		logger.Error("Failed to send first response", slog.String("error", err.Error()))
		return
	}

	makeStatusMsg := func(state daityou.DaityouCell) string {
		authorUser, err := s.GuildMember(manager.GuildID, state.AuthorUser)
		if err != nil {
			return ""
		}
		rentaledUser, err := s.GuildMember(manager.GuildID, state.RentaledUser)
		if err != nil {
			return ""
		}

		result := ""
		result += fmt.Sprintf("%s -> %s : %s¥", authorUser.Mention(), rentaledUser.Mention(), formatWithCommna(state.Amount))
		return result
	}

	contentStr := ""
	for _, state := range states {
		if state.Amount == 0 {
			continue
		}

		str := makeStatusMsg(state)
		if str == "" {
			continue
		}
		contentStr += str + "\n"
	}

	authorNick := lib.GetGuildNick(s, manager.GuildID, i.Member.User.ID)

	embed := &discordgo.MessageEmbed{
		Title: "貸し借り状況",
		Color: 0x00F1AA,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: manager.BotUserInfo.AvatarURL("24"),
		},
		Author: &discordgo.MessageEmbedAuthor{
			Name:    authorNick,
			IconURL: i.Member.User.AvatarURL("48"),
		},
		Description: contentStr,
	}

	logger.Debug("DebeHandler Made Embed",
		slog.String("contentStr", contentStr),
	)

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})

	if err != nil {
		logger.Error("Failed to send response", slog.String("error", err.Error()))
		return
	}

	logger.Debug("DebtHandler finished")
}
