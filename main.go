package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
)

func main() {
	slog.Info("starting up...")
	initData()
	config := loadConfig()
	token := config.Token
	if token == "" {
		slog.Error("Discord token not set")
		os.Exit(1)
	}

	if err := loadState(); err != nil {
		slog.Error("failed to load state", "error", err)
	}

	client, err := disgo.New(token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildMessages, gateway.IntentMessageContent),
		),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagGuilds, cache.FlagChannels, cache.FlagMessages),
		),
		bot.WithEventListenerFunc(messageCreate),
		bot.WithEventListenerFunc(messageDelete),
		bot.WithEventListenerFunc(func(e *events.Ready) {
			slog.Info("Ready!", "user", e.User.Username, "id", e.User.ID)
		}),
	)
	if err != nil {
		slog.Error("failed to create client", "error", err)
		os.Exit(1)
	}
	// connect to the gateway
	if err = client.OpenGateway(context.TODO()); err != nil {
		slog.Error("failed to connect to gateway", "error", err)
		os.Exit(1)
	}

	defer func() {
		slog.Info("shutting down...")
		client.Close(context.TODO())
	}()

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
}

var (
	lastNumber int          = 0
	channelID  snowflake.ID = 0
	lastPerson snowflake.ID = 0
)

func messageCreate(e *events.MessageCreate) {
	if e.Message.Author.Bot {
		return
	}
	content := e.Message.Content
	var response string
	if strings.HasPrefix(content, "d!") {
		switch content[2:] {
		case "ping":
			response = "pong"
		case "setchannel":
			channelID = e.Message.ChannelID
			response = "Kanal wurde auf <#" + channelID.String() + "> gesetzt"
			if err := saveState(); err != nil {
				slog.Error("failed to persist state", "error", err)
			}
		case "help":
			response = "d!ping, d!setchannel, d!help\nEine Person kann nicht zwei Nachrichten hintereinander senden."
		default:
			response = "Unbekannter Befehl. d!help"
		}
	} else {
		if e.Message.Author.ID == lastPerson || channelID != e.Message.ChannelID {
			return
		}
		// check if content is a number
		if number, err := strconv.Atoi(content); err == nil {
			if number == lastNumber+1 {
				lastNumber = lastNumber + 1
				lastPerson = e.Message.Author.ID
				if err := saveState(); err != nil {
					slog.Warn("failed to persist state", "error", err)
				}
			} else {
				lastNumber = 0
				response = fmt.Sprintf("%s hat die Strähne unterbrochen. :(", e.Message.Author.Mention())
				if err := saveState(); err != nil {
					slog.Warn("failed to persist state", "error", err)
				}
			}
		}
	}

	if response != "" {
		_, _ = e.Client().Rest().CreateMessage(e.ChannelID, discord.NewMessageCreateBuilder().SetContent(response).Build())
	}
}

func messageDelete(e *events.MessageDelete) {
	msg := e.Message
	if msg.Content == strconv.Itoa(lastNumber) {
		_, _ = e.Client().Rest().CreateMessage(
			channelID,
			discord.NewMessageCreateBuilder().
				SetContent(fmt.Sprintf("%s hat die letzte Nachricht gelöscht. Zählerstand: %d", msg.Author.Mention(), lastNumber)).
				Build(),
		)
	}
}
