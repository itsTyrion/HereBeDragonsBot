package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
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
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMessages,
				gateway.IntentMessageContent,
				gateway.IntentGuildMembers,
				gateway.IntentGuildPresences,
			),
		),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagGuilds, cache.FlagChannels, cache.FlagMessages, cache.FlagMembers, cache.FlagRoles),
		),
		bot.WithEventListenerFunc(func(e *events.MessageCreate) {
			go messageCreate(e)
		}),
		bot.WithEventListenerFunc(messageDelete),
		bot.WithEventListenerFunc(func(e *events.Ready) {
			slog.Info("Ready!", "user", e.User.Username, "id", e.User.ID)
		}),
		bot.WithEventListenerFunc(func(e *events.GuildReady) {
			go updateMemberList(e.Client(), e.GuildID)
		}),
	)
	if err != nil {
		slog.Error("failed to create client", "error", err)
		os.Exit(1)
	}

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
