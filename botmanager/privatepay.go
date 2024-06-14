package botmanager

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/kohinigeee/DiscordDaityouBot/lib"
	"github.com/kohinigeee/DiscordDaityouBot/mylogger"
)

const (
	PrivatePaySelectUserHandlerID = "privatpay_select_users"

	PrivatePayModalName   = "privatepay_modal"
	privatePayCommandName = "PrivatePay"

	privatePayModalAmountInputID = "privatepay_input_money"
	privatePayModalDescInputID   = "privatepay_input_desc"

	privatePayMessageTitle = "Private Pay"
)

type PrivatePayData struct {
	authorMember *discordgo.Member
	targetMember *discordgo.Member
	amount       uint
	desc         string
}

func (data *PrivatePayData) SetAmount(amount uint) {
	data.amount = amount
}

func deferPrivatePay(userID string, manager *BotManager) {
	manager.DeleteHandlerMemory(privatePayCommandName, userID)
}

func NewPrivatePayDateFromID(manager *BotManager, authorID string, rentaledUserID string, desc string, amount uint) (*PrivatePayData, error) {
	authorMember, err := lib.GetMemberFromID(manager.Session, manager.GuildID, authorID)
	if err != nil {
		return nil, err
	}

	rentaledMember, err := lib.GetMemberFromID(manager.Session, manager.GuildID, rentaledUserID)
	if err != nil {
		return nil, err
	}

	return &PrivatePayData{
		authorMember: authorMember,
		targetMember: rentaledMember,
		amount:       amount,
		desc:         desc,
	}, nil
}

func PrivatePayHandler(s *discordgo.Session, i *discordgo.InteractionCreate, manager *BotManager) {
	logger := mylogger.L()
	logger.Debug("PrivatePayHandler started", slog.String("user", fmt.Sprintf("%+v", i.Member.User.Username)), slog.String("ID", i.ID), slog.String("InteractionID", i.Interaction.ID))

	const handlerName = "Private Pay"
	authorID := i.Member.User.ID

	users, err := lib.GetAllGuildMembers(s, i.GuildID)

	if err != nil {
		logger.Error("Error getting all users", slog.String("err", err.Error()))

		errmsg := "ユーザ一覧の取得に失敗しました"
		manager.SendErrorMessage(i.ChannelID, handlerName, errmsg, nil)

		deferEasyPay(authorID, manager)
		return
	}

	options := make([]discordgo.SelectMenuOption, 0)

	for _, user := range users {

		if user.User.ID == authorID {
			continue
		}

		// ボットは除外
		if isExcludeBOT && user.User.Bot {
			continue
		}

		var label string
		if user.Nick != "" {
			label = user.Nick
		} else if user.User.GlobalName != "" {
			label = user.User.GlobalName
		} else {
			label = user.User.Username
		}

		options = append(options, discordgo.SelectMenuOption{
			Label: label,
			Value: user.User.ID,
		})
	}

	var minValues = 1

	selectMenu := discordgo.SelectMenu{
		CustomID:    PrivatePaySelectUserHandlerID,
		Options:     options,
		MaxValues:   1,
		MinValues:   &minValues,
		Placeholder: "貸すユーザーを選択してください",
	}

	actionRow := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			&selectMenu,
		},
	}

	embed := manager.MakeNormalMessageEmbed(handlerName, "", nil)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{actionRow},
		},
	})

	if err != nil {
		logger.Error("Error sending message", slog.String("err", err.Error()))

		deferPrivatePay(authorID, manager)
		return
	}

	logger.Debug("PrivatePayHandler finished")
}

// ユーザーを選択した後のハンドラ
// 金額と使途を入力するモーダルを表示

func PrivatePaySelectUserHandler(s *discordgo.Session, i *discordgo.InteractionCreate, manager *BotManager) {
	logger := mylogger.L()

	logger.Debug("PrivatePaySelectUserHandler started", slog.String("user", fmt.Sprintf("%+v", i.Member.User.Username)), slog.String("ID", i.ID), slog.String("InteractionID", i.Interaction.ID))

	var selectedMember *discordgo.Member
	authorID := i.Member.User.ID

	// 選択されたユーザーを取得

	userID := i.MessageComponentData().Values[0]
	selectedMember, err := s.GuildMember(i.GuildID, userID)
	if err != nil {
		logger.Error("Error getting selected user", slog.String("err", err.Error()))

		errmsg := "ユーザーの取得に失敗しました"
		manager.SendErrorMessage(i.ChannelID, privatePayCommandName, errmsg, nil)

		deferPrivatePay(i.Member.User.ID, manager)
		return
	}

	data := &PrivatePayData{
		authorMember: i.Member,
		targetMember: selectedMember,
	}

	manager.SetHandlerMemory(privatePayCommandName, authorID, data)

	//-----------モーダルの生成-----------

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: PrivatePayModalName,
			Title:    "Private Pay",
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    privatePayModalAmountInputID,
							Label:       "相手に支払う金額を入力してください (例: 1000)",
							Placeholder: "15000",
							Style:       discordgo.TextInputShort,
							Required:    true,
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    privatePayModalDescInputID,
							Label:       "お金の使途を入力してください",
							Style:       discordgo.TextInputShort,
							Placeholder: "ランチ代",
							Required:    true,
						},
					},
				},
			},
		},
	})

	if err != nil {
		logger.Error("Error sending message", slog.String("err", err.Error()))

		deferPrivatePay(authorID, manager)
		return
	}

	logger.Debug("PrivatePaySelectUserHandler finished")
}

// 金額と使途を入力後のハンドラ
// 台帳処理を行う
func PrivatePayModalHandler(s *discordgo.Session, i *discordgo.InteractionCreate, manager *BotManager) {
	const commandName = "Private Pay"

	logger := mylogger.L()
	logger.Debug("PrivatePayModalHandler started", slog.String("user", fmt.Sprintf("%+v", i.Member.User.Username)), slog.String("ID", i.ID), slog.String("InteractionID", i.Interaction.ID))

	authorID := i.Member.User.ID

	//-------メモリのデータを取得-------
	data, ok := manager.GetHandlerMemory(privatePayCommandName, authorID).(*PrivatePayData)

	if !ok {
		logger.Error("Error getting handler memory", slog.String("err", "type assertion error"))

		errmsg := "データの取得に失敗しました"
		manager.SendErrorMessage(i.ChannelID, commandName, errmsg, nil)

		deferPrivatePay(authorID, manager)
		return
	}

	logger.Debug("PrivatePayMOdalHandler GetMemoryData")
	fmt.Printf("%+v\n", data)

	//-------モーダルのデータを取得-------

	modalData := i.ModalSubmitData()
	amountStr := lib.GetModalDataValue(&modalData, 0, 0)
	desc := lib.GetModalDataValue(&modalData, 1, 0)

	amountInt, err := strconv.Atoi(amountStr)
	if err != nil {
		logger.Error("Error converting amount", slog.String("err", err.Error()))
		errmsg := fmt.Sprintf(" '%s' は金額として不正です", amountStr)
		manager.SendErrorMessage(i.ChannelID, commandName, errmsg, nil)

		deferEasyPay(authorID, manager)
		return
	}

	if amountInt <= 0 {
		logger.Error("Error converting amount", slog.String("err", "amount is less than or equal to 0"))
		errmsg := fmt.Sprintf("金額は0以上で入力してください (入力値: %d)", amountInt)
		manager.SendErrorMessage(i.ChannelID, commandName, errmsg, nil)

		deferEasyPay(authorID, manager)
		return
	}

	amount := uint(amountInt)

	logger.Debug("PrivatePayModalHandler GetModalData",
		slog.Int("amount", int(amount)),
		slog.String("desc", desc),
	)

	data.desc = desc

	//-------台帳処理-------
	logger.Debug("PrivatePayCalculation start")

	data.SetAmount(amount)
	PrivatePayApplyManager(manager, data)

	link := fmt.Sprintf("https://discord.com/channels/%s/%s", i.GuildID, manager.RestoreThreadID)

	notionEmbed := manager.MakeNormalMessageEmbed("Easy Pay", fmt.Sprintf("お金のやりとりが完了したぶぅ\n(%s)", link), nil)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{notionEmbed},
		},
	})

	nick := lib.GetGuildNick(s, i.GuildID, authorID)
	embed := makePrivatePayRentHistoryEmbed(nick, i.Member.User, manager.BotUserInfo, data)

	s.ChannelMessageSendEmbed(manager.RestoreThreadID, embed)

	deferPrivatePay(authorID, manager)
	logger.Debug("PrivatePayModalHandler finished")
}

func PrivatePayApplyManager(manager *BotManager, data *PrivatePayData) {
	daityouManager := manager.GetDaityouManager()
	targetID := data.targetMember.User.ID
	authorID := data.authorMember.User.ID
	daityouManager.EasyPay(authorID, targetID, data.amount)
}

func makePrivatePayRentHistoryEmbed(authorNick string, authorUser *discordgo.User, botUser *discordgo.User, data *PrivatePayData) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title: privatePayMessageTitle,
		Color: ImageColorHex,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: botUser.AvatarURL("24"),
		},
		Author: &discordgo.MessageEmbedAuthor{
			Name:    authorNick,
			IconURL: authorUser.AvatarURL("48"),
		},
	}

	fileds := make([]*discordgo.MessageEmbedField, 0)

	fileds = append(fileds, &discordgo.MessageEmbedField{
		Name:   "貸し主",
		Value:  authorUser.Mention(),
		Inline: true,
	})
	fileds = append(fileds, &discordgo.MessageEmbedField{
		Name:   "借り主",
		Value:  data.targetMember.User.Mention(),
		Inline: true,
	})
	fileds = append(fileds, &discordgo.MessageEmbedField{
		Name:   "金額",
		Value:  fmt.Sprintf("%d ¥", data.amount),
		Inline: true,
	})

	fileds = append(fileds, &discordgo.MessageEmbedField{
		Name:   "使途",
		Value:  data.desc,
		Inline: true,
	})

	embed.Fields = fileds

	return embed
}
