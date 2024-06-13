package lib

import "github.com/bwmarrin/discordgo"

func GetModalDataValue(data *discordgo.ModalSubmitInteractionData, componentIndex int, inputIndex int) string {
	return data.Components[componentIndex].(*discordgo.ActionsRow).Components[inputIndex].(*discordgo.TextInput).Value
}

func GetOptionByName(options []*discordgo.ApplicationCommandInteractionDataOption, name string) *discordgo.ApplicationCommandInteractionDataOption {
	for _, option := range options {
		if option.Name == name {
			return option
		}
	}
	return nil
}

func SendEmptyInteractionResponse(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
	})
}

func GetAllGuildMembers(s *discordgo.Session, guildID string) ([]*discordgo.Member, error) {

	var allMembers []*discordgo.Member
	limit := 100
	after := ""

	for {
		members, err := s.GuildMembers(guildID, after, limit)
		if err != nil {
			return nil, err
		}

		if len(members) == 0 {
			break
		}

		allMembers = append(allMembers, members...)
		after = members[len(members)-1].User.ID

		if len(members) < limit {
			break
		}
	}

	return allMembers, nil
}

func GetGuildNick(s *discordgo.Session, guildID string, userID string) string {
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		return ""
	}

	if member.Nick == "" {
		return member.User.Username
	}
	return member.Nick
}

func GetFirstMessage(s *discordgo.Session, channelID string) (*discordgo.Message, error) {
	messages, err := s.ChannelMessages(channelID, 10, "", "", "")
	if err != nil {
		return nil, err
	}

	if len(messages) == 0 {
		return nil, nil
	}

	return messages[0], nil
}
