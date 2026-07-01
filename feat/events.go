package feat

import (
	. "HereBeDragons/common"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var botStartTime time.Time = time.Now()
var state BotState
var config BotConfig

func InitEvents(bs BotState, bc BotConfig) *events.ListenerAdapter {
	state = bs
	config = bc
	return &events.ListenerAdapter{
		OnMessageCreate: messageCreate,
		OnMessageDelete: messageDelete,
	}
}

func messageCreate(e *events.MessageCreate) {
	if e.Message.Author.Bot || e.GuildID == nil {
		return
	}
	client := e.Client()
	content := e.Message.Content
	var response string
	if strings.HasPrefix(content, "d!") {
		response = handleCommand(e)
	} else {
		if e.Message.Author.ID == state.LastPerson || state.ChannelID != e.Message.ChannelID {
			return
		}
		// check if content is a number
		if number, err := strconv.Atoi(content); err == nil && number == state.LastNumber+1 {
			state.LastNumber = state.LastNumber + 1
			state.LastPerson = e.Message.Author.ID
			if err = client.Rest.AddReaction(e.ChannelID, e.Message.ID, "✅"); err != nil {
				slog.Error("failed to add reaction", "error", err)
			}
			if err := SaveState(state); err != nil {
				slog.Warn("messageCreate: failed to persist state", "error", err)
			}
		} else {
			slog.Info("messageCreate: content is not a number")
			state.LastNumber = 0
			response = fmt.Sprintf("%s hat die Strähne unterbrochen. :(", e.Message.Author.Mention())
			slog.Info("messageCreate: lastNumber", "lastNumber", state.LastNumber, "lastPerson", state.LastPerson)
			if err := SaveState(state); err != nil {
				slog.Warn("messageCreate: failed to persist state", "error", err)
			}
		}
	}

	if response != "" {
		_, _ = client.Rest.CreateMessage(e.ChannelID, discord.MessageCreate{Content: response})
	}
}

func messageDelete(e *events.MessageDelete) {
	msg := e.Message
	if msg.Content == strconv.Itoa(state.LastNumber) {
		_, _ = e.Client().Rest.CreateMessage(
			state.ChannelID,
			discord.MessageCreate{
				Content: fmt.Sprintf("%s hat die letzte Nachricht gelöscht. Zählerstand: %d", msg.Author.Mention(), state.LastNumber),
			},
		)
	}
}
