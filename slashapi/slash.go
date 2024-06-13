package slashapi

//DiscordAPIのギルドコマンドを登録するためのパッケージ

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/kohinigeee/DiscordDaityouBot/mylogger"
)

var (
	appID        string
	guildId      string
	discordToken string
	logger       *slog.Logger
	client       *http.Client
)

type SlashCommandOptionJson struct {
	Type        discordgo.ApplicationCommandOptionType `json:"type"`
	Name        string                                 `json:"name"`
	Description string                                 `json:"description"`
	Required    bool                                   `json:"required"`
}

type SlashCommandJson struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Options     []SlashCommandOptionJson `json:"options"`
}

func InitEnvs() {
	appID = os.Getenv("APPLICATION_ID")
	guildId = os.Getenv("GUILD_ID")
	discordToken = "Bot " + os.Getenv("DISCORD_TOKEN")

	logger = mylogger.L()

	logger.Info("[slash.go init func] Imported env values", slog.String("APPLICATION_ID", appID), slog.String("GUILD_ID", guildId), slog.String("DISCORD_TOKEN", discordToken))

	client = &http.Client{}
}

func setHeader(req *http.Request) {
	req.Header.Set("Authorization", discordToken)
	req.Header.Set("Content-Type", "application/json")
}

func ApplySlashCommand(command *SlashCommandJson) error {
	endPoint := fmt.Sprintf("https://discord.com/api/v8/applications/%s/guilds/%s/commands", appID, guildId)

	requestBody, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("[ApplySlashCommand] Error marshaling command: %w", err)
	}

	req, err := http.NewRequest("POST", endPoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("[ApplySlashCommand] Error creating request: %w", err)
	}

	setHeader(req)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("[ApplySlashCommand] Error sending request: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("[ApplySlashCommand] Error reading response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("[ApplySlashCommand] Error unmarshaling response: %w", err)
	}

	statusCode := resp.StatusCode
	logger.Info(fmt.Sprintf("[Slash command results : %s]", command.Name), slog.Int("Status", statusCode), slog.String("command", command.Name), slog.String("response", string(body)))

	allowStatus := map[int]interface{}{
		http.StatusOK:      nil,
		http.StatusCreated: nil,
	}

	if _, ok := allowStatus[statusCode]; !ok {
		return fmt.Errorf("[ ApplySlashCommand ] Error applying slash command: %s", string(body))
	}

	return nil
}
