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
	EasyPaySelectUsersHandlerID = "easypay_select_users"

	EasyPayModalName = "easypay_modal"

	easyPayCommandName = "EasyPay"

	easyPayModalAmountInputID = "easypay_input_money"
	easyPayModalDescInputID   = "easypay_input_desc"

	easyPayMessageTitle = "Easy Pay"
)

type EasyPayData struct {
	authorMember  *discordgo.Member
	targetMembers []*discordgo.Member
	amount        uint
	desc          string
	amountUnit    uint
}

func NewEasyPayDataFromIDs(manager *BotManager, authorID string, targetIDs []string, desc string, amount uint) (*EasyPayData, error) {
	data := &EasyPayData{}

	authorMember, err := lib.GetMemberFromID(manager.Session, manager.GuildID, authorID)
	if err != nil {
		return nil, err
	}

	targetMembers, err := lib.GetMembersFromIDs(manager.Session, manager.GuildID, targetIDs)
	if err != nil {
		return nil, err
	}

	data.targetMembers = targetMembers
	data.authorMember = authorMember
	data.desc = desc

	data.SetAmount(amount)
	return data, nil
}

func (data *EasyPayData) SetAmount(amount uint) {
	amountUnit := amount / uint(len(data.targetMembers)+1)
	data.amount = amount
	data.amountUnit = amountUnit
}

func deferEasyPay(userID string, manager *BotManager) {
	manager.DeleteHandlerMemory(easyPayCommandName, userID)
}

// ユーザーを選択するハンドラ
func EasyPayHandler(s *discordgo.Session, i *discordgo.InteractionCreate, manager *BotManager) {
	logger := mylogger.L()
	logger.Debug("EasyPayHandler started", slog.String("user", fmt.Sprintf("%+v", i.Member.User.Username)), slog.String("ID", i.ID), slog.String("InteractionID", i.Interaction.ID))

	authorID := i.Member.User.ID
	const handlerName = "Easy Pay"

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
		// if user.User.Bot {
		// 	continue
		// }

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
		CustomID:    EasyPaySelectUsersHandlerID,
		Options:     options,
		MaxValues:   len(options),
		MinValues:   &minValues,
		Placeholder: "ユーザーを選択してください",
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

		deferEasyPay(authorID, manager)
		return
	}

	logger.Debug("EasyPayHandler finished")
}

// ユーザーを選択した後のハンドラ
// 金額と使途を入力するモーダルを表示
func EasyPaySelectUsersHandler(s *discordgo.Session, i *discordgo.InteractionCreate, manager *BotManager) {
	logger := mylogger.L()

	logger.Debug("EasyPaySelectUsersHandler started", slog.String("user", fmt.Sprintf("%+v", i.Member.User.Username)), slog.String("ID", i.ID), slog.String("InteractionID", i.Interaction.ID))

	var selectedUsers []*discordgo.Member

	//--------セレクトメニューの処理---------
	for _, userID := range i.MessageComponentData().Values {
		member, err := s.GuildMember(i.GuildID, userID)

		if err != nil {
			logger.Error("Error getting user", slog.String("err", err.Error()))
			continue
		}

		selectedUsers = append(selectedUsers, member)
	}

	if len(selectedUsers) == 0 {
		logger.Error("Error getting user", slog.String("err", "no user selected"))
		errmsg := "ユーザーが選択されていません"
		manager.SendErrorMessage(i.ChannelID, "EasyPay", errmsg, nil)

		deferEasyPay(i.Member.User.ID, manager)
		return
	}

	authorID := i.Member.User.ID
	data := EasyPayData{
		authorMember:  i.Member,
		targetMembers: selectedUsers,
	}

	manager.SetHandlerMemory(easyPayCommandName, authorID, data)

	//-----------モーダルの生成-----------
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: EasyPayModalName,
			Title:    "Easy Pay",
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    easyPayModalAmountInputID,
							Label:       "金額の総額を入力してください (例: 1000)",
							Placeholder: "15000",
							Style:       discordgo.TextInputShort,
							Required:    true,
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    easyPayModalDescInputID,
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
		logger.Error("Error responding to interaction", slog.String("err", err.Error()))

		deferEasyPay(authorID, manager)
		return
	}

	logger.Debug("EasyPaySelectUsersHandler finished")
}

// 金額と使途を入力後のハンドラ
// 台帳処理を行う
func EasyPayModalHandler(s *discordgo.Session, i *discordgo.InteractionCreate, manager *BotManager) {

	const commandName = "Easy Pay"

	logger := mylogger.L()
	logger.Debug("EasyPayModalHandler started", slog.String("user", fmt.Sprintf("%+v", i.Member.User.Username)), slog.String("ID", i.ID), slog.String("InteractionID", i.Interaction.ID))

	authorID := i.Member.User.ID

	// -------メモリのデータを取得--------
	data, ok := manager.GetHandlerMemory(easyPayCommandName, authorID).(EasyPayData)

	if !ok {
		logger.Error("Error getting handler memory", slog.String("err", "failed to assert handler memory"))
		errmsg := "ハンドラメモリの取得に失敗しました"
		manager.SendErrorMessage(i.ChannelID, "EasyPay", errmsg, nil)

		deferEasyPay(authorID, manager)
		return
	}

	logger.Debug("EasyPayModalHandler GetMemoryData")
	fmt.Printf("%+v\n", data)

	// -------モーダルのデータを取得--------

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

	logger.Debug("EasyPayModalHandler GetModalData",
		slog.Int("amount", int(amount)),
		slog.String("desc", desc),
	)

	data.desc = desc

	// -------台帳処理--------
	data.SetAmount(amount)
	EasyPayApplyManager(manager, &data)

	logger.Debug("EasyPay Calculation start")

	link := fmt.Sprintf("https://discord.com/channels/%s/%s", i.GuildID, manager.RestoreThreadID)

	logger.Debug("EasyPay Calculation finished",
		slog.String("link", link))

	notionEmbed := manager.MakeNormalMessageEmbed("Easy Pay", fmt.Sprintf("お金のやりとりが完了したぶぅ\n(%s)", link), nil)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{notionEmbed},
		},
	})

	nick := lib.GetGuildNick(s, i.GuildID, authorID)
	embed := makeEmbedRentHistory(nick, data.authorMember.User, manager.BotUserInfo, &data)

	s.ChannelMessageSendEmbed(manager.RestoreThreadID, embed)

	deferEasyPay(authorID, manager)
	logger.Debug("EasyPayModalHandler finished")
}

func EasyPayApplyManager(manager *BotManager, data *EasyPayData) {

	daityouManager := manager.GetDaityouManager()
	for _, targetMamber := range data.targetMembers {
		daityouManager.EasyPay(data.authorMember.User.ID, targetMamber.User.ID, data.amountUnit)
	}
}

func makeEmbedRentHistory(authorNick string, authorUser *discordgo.User, botUser *discordgo.User, data *EasyPayData) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title: easyPayMessageTitle,
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
		Name:   "金額",
		Value:  fmt.Sprintf("%d ¥\n(%d¥/1人)", data.amount, data.amountUnit),
		Inline: true,
	})

	fileds = append(fileds, &discordgo.MessageEmbedField{
		Name:   "使途",
		Value:  data.desc,
		Inline: true,
	})

	//貸し先のユーザーを表示
	membersStr := ""
	for _, member := range data.targetMembers {
		membersStr += member.User.Mention() + "  "
	}

	fileds = append(fileds, &discordgo.MessageEmbedField{
		Name:   "借り主",
		Value:  membersStr,
		Inline: false,
	})

	embed.Fields = fileds

	return embed
}
