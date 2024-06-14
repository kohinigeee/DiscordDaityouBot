package botmanager

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kohinigeee/DiscordDaityouBot/daityou"
	"github.com/kohinigeee/DiscordDaityouBot/lib"
	"github.com/kohinigeee/DiscordDaityouBot/mylogger"
)

const (
	ResetHandlerID = "reset_handler"
)

type resetHandlerCalcFunc func(s *discordgo.Session, i *discordgo.InteractionCreate, manager *BotManager, message *discordgo.Message)

var (
	funcMap map[string]resetHandlerCalcFunc = map[string]resetHandlerCalcFunc{
		easyPayMessageTitle:    calcEasyPayMessage,
		privatePayMessageTitle: calcPrivatePayMessage,
	}
)

// 計算対象のメッセージかどうかを判定する
func isResetHandlerTargetMessage(message *discordgo.Message) bool {
	embeds := message.Embeds

	if len(embeds) == 0 {
		return false
	}

	embed := embeds[0]
	title := embed.Title

	_, exist := funcMap[title]

	return exist
}

func GetAllTargetMessages(s *discordgo.Session, manager *BotManager) ([]*discordgo.Message, error) {
	afterID := ""
	const limit = 100
	targetMessages := make([]*discordgo.Message, 0)

	for {
		messages, err := s.ChannelMessages(manager.RestoreThreadID, limit, "", afterID, "")

		if err != nil {
			return targetMessages, err
		}

		if len(messages) == 0 {
			break
		}

		for _, message := range messages {
			if isResetHandlerTargetMessage(message) {
				targetMessages = append(targetMessages, message)
			}
		}

		if len(messages) < limit {
			break
		}

		afterID = messages[len(messages)-1].ID
	}

	return targetMessages, nil
}

func ResetHandler(s *discordgo.Session, i *discordgo.InteractionCreate, manager *BotManager) {
	logger := mylogger.L()
	const commandName = "Reset"

	logger.Debug("ResetHandler started")

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	if err != nil {
		logger.Error("Failed to send first response", slog.String("error", err.Error()))
		return
	}

	targetMeesages, err := GetAllTargetMessages(s, manager)

	//メッセージを取得できなかった場合のエラー処理
	if err != nil {
		logger.Error("Error getting all messages", err)

		// embed := manager.MakeErrorMessageEmbed(commandName, "エラーが発生しました", nil)
		// err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// 	Type: discordgo.InteractionResponseChannelMessageWithSource,
		// 	Data: &discordgo.InteractionResponseData{
		// 		Embeds: []*discordgo.MessageEmbed{embed},
		// 	},
		// })

		manager.SendErrorMessageInteractionEdit(i.Interaction, commandName, "エラーが発生しました", nil)

		return
	}

	newDaityouManager := daityou.NewDaityouManager()
	manager.daityouManager = newDaityouManager

	for _, message := range targetMeesages {
		embed := message.Embeds[0]
		calcFunction := funcMap[embed.Title]
		calcFunction(s, i, manager, message)
	}

	manager.SendNormalMessageInteractionEdit(i.Interaction, commandName, "リセットが完了しました", nil)

	logger.Debug("ResetHandler finished")
}

func strToAmount(str string) uint {
	words := strings.Fields(str)
	amountInt, _ := strconv.Atoi(words[0])
	return uint(amountInt)
}

func calcEasyPayMessage(s *discordgo.Session, i *discordgo.InteractionCreate, manager *BotManager, message *discordgo.Message) {
	logger := mylogger.L()
	logger.Debug("calcEasyPayMessage started")

	embed := message.Embeds[0]

	fileds := embed.Fields

	getIDs := func(fileds []*discordgo.MessageEmbedField) (authorID string, rentaledUserIDs []string, err error) {
		err = nil
		authorID, err = lib.GetIDFromMention(fileds[0].Value)
		if err != nil {
			return
		}

		rentaledUserMentions := strings.Fields(fileds[3].Value)
		rentaledUserIDs = make([]string, 0)
		for _, rentaledUserMention := range rentaledUserMentions {
			var id string
			id, err = lib.GetIDFromMention(rentaledUserMention)
			if err != nil {
				return
			}
			rentaledUserIDs = append(rentaledUserIDs, id)
		}
		return
	}

	authorID, rentaledUserIDs, err := getIDs(fileds)
	amount := strToAmount(fileds[1].Value)
	desc := fileds[2].Value
	if err != nil {
		logger.Error("Error getting IDs", slog.String("error", err.Error()))
		return
	}

	logger.Debug("calcEasyPayMessage 整形後",
		slog.String("authorID", authorID),
		slog.String("rentaledUserIDs", fmt.Sprintf("%v", rentaledUserIDs)),
		slog.Int("amount", int(amount)),
	)

	data, err := NewEasyPayDataFromIDs(manager, authorID, rentaledUserIDs, desc, amount)

	if err != nil {
		logger.Error("Error creating EasyPayData", slog.String("error", err.Error()))
		return
	}

	EasyPayApplyManager(manager, data)

	logger.Debug("calcEasyPayMessage finished")
}

func calcPrivatePayMessage(s *discordgo.Session, i *discordgo.InteractionCreate, manager *BotManager, message *discordgo.Message) {

	logger := mylogger.L()
	logger.Debug("calcPrivatePayMessage started")

	embed := message.Embeds[0]
	fileds := embed.Fields

	getIDs := func(fileds []*discordgo.MessageEmbedField) (authorID string, rentaledUserID string, err error) {
		err = nil
		authorID, err = lib.GetIDFromMention(fileds[0].Value)
		if err != nil {
			return
		}

		rentaledUserID, err = lib.GetIDFromMention(fileds[1].Value)
		if err != nil {
			return
		}
		return
	}

	authorID, rentaledUserID, err := getIDs(fileds)
	amount := strToAmount(fileds[2].Value)
	desc := fileds[3].Value
	if err != nil {
		logger.Error("Error getting IDs", slog.String("error", err.Error()))
		return
	}

	logger.Debug("calcPrivatePayMessage 整形後",
		slog.String("authorID", authorID),
		slog.String("rentaledUserID", rentaledUserID),
		slog.Int("amount", int(amount)),
	)

	data, err := NewPrivatePayDateFromID(manager, authorID, rentaledUserID, desc, amount)

	if err != nil {
		logger.Error("Error creating PrivatePayData", slog.String("error", err.Error()))
		return
	}

	PrivatePayApplyManager(manager, data)

	logger.Debug("calcPrivatePayMessage finished")
}
