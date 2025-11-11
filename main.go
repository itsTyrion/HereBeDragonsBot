package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

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
		bot.WithEventListenerFunc(func(e *events.MessageCreate) {
			go messageCreate(e)
		}),
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
	startTime  time.Time    = time.Now()
)

func messageCreate(e *events.MessageCreate) {
	if e.Message.Author.Bot {
		return
	}
	client := e.Client()
	rest := client.Rest()
	content := e.Message.Content
	var response string
	if strings.HasPrefix(content, "d!") {
		raw := strings.Split(content, " ")
		slog.Info("messageCreate: raw", "raw", raw)
		args := raw[1:]
		switch raw[0][2:] {
		case "ping":
			response = "pong"
		case "setchannel":
			channelID = e.Message.ChannelID
			response = "Kanal wurde auf <#" + channelID.String() + "> gesetzt"
			if err := saveState(); err != nil {
				slog.Error("failed to persist state", "error", err)
			}
		case "help":
			response = "d!ping, d!setchannel, d!help, d!about, d!purge <Anzahl>\n" +
				"Eine Person kann nicht zwei Nachrichten hintereinander senden."
		case "about":
			selfMember, _ := rest.GetUser(client.ID())
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			uptime := time.Since(startTime)
			uptimeFmt := fmt.Sprintf("%d:%02d:%02d:%02d",
				int(uptime.Hours()/24), int(uptime.Hours())%24, int(uptime.Minutes())%60, int(uptime.Seconds())%60)

			embed := discord.NewEmbedBuilder().
				SetColor(0x9B4F96).
				AddField("**Dev**: ", "itsTyrion (<@!265038515375570944>)", false).
				AddField("**Ping**: ", fmt.Sprintf("%dms", e.Client().Gateway().Latency().Milliseconds()), false).
				AddField("**Uptime**: ", uptimeFmt, false).
				AddField("**RAM**: ", fmt.Sprintf("%dMB alloc/%dMB sys", m.HeapAlloc/1024/1024, m.HeapSys/1024/1024), false).
				AddField("**Powered by**: ", runtime.Version()+", disgo and Deez Nuts", false).
				SetFooter("Made (with love) in Germany", "https://itstyrion.de/random/MC-Heart.png").
				SetAuthor("HereBeDragons", "https://youtu.be/qWNQUvIk954", *selfMember.AvatarURL()).
				Build()
			_, _ = rest.CreateMessage(e.ChannelID, discord.MessageCreate{Embeds: []discord.Embed{embed}})
		case "purge":
			if len(args) == 0 {
				response = "Bitte gib eine Anzahl an Nachrichten an."
				return
			}
			if number, err := strconv.Atoi(args[0]); err == nil {
				if number < 1 || number > 100 {
					response = "Bitte gib zwischen 1 und 100 an."
					return
				}
				if err = rest.AddReaction(e.ChannelID, e.Message.ID, "⏳"); err != nil {
					slog.Error("failed to add reaction", "error", err)
				}
				if page, err := rest.GetMessages(e.ChannelID, 0, e.Message.ID, 0, number); err == nil {
					messageIDs := make([]snowflake.ID, 0, len(page))
					for _, msg := range page {
						messageIDs = append(messageIDs, msg.ID)
					}
					if err := rest.BulkDeleteMessages(e.ChannelID, messageIDs); err != nil {
						slog.Error("failed to purge messages", "error", err)
						response = "Fehler beim Löschen der Nachrichten: " + err.Error()
					} else {
						rest.RemoveOwnReaction(e.ChannelID, e.Message.ID, "⏳")
						rest.AddReaction(e.ChannelID, e.Message.ID, "✅")
					}
				} else {
					slog.Error("failed to get messages", "error", err)
					response = "Fehler beim Löschen der Nachrichten: " + err.Error()
				}
			}
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
				if err = rest.AddReaction(e.ChannelID, e.Message.ID, "✅"); err != nil {
					slog.Error("failed to add reaction", "error", err)
				}
				if err := saveState(); err != nil {
					slog.Warn("messageCreate: failed to persist state", "error", err)
				}
			} else {
				slog.Info("messageCreate: number is incorrect")
				lastNumber = 0
				response = fmt.Sprintf("%s hat die Strähne unterbrochen. :(", e.Message.Author.Mention())
				slog.Info("messageCreate: lastNumber", "lastNumber", lastNumber, "lastPerson", lastPerson)
				if err := saveState(); err != nil {
					slog.Warn("messageCreate: failed to persist state", "error", err)
				}
			}
		}
	}

	if response != "" {
		_, _ = client.Rest().CreateMessage(e.ChannelID, discord.MessageCreate{Content: response})
	}
}

func messageDelete(e *events.MessageDelete) {
	msg := e.Message
	if msg.Content == strconv.Itoa(lastNumber) {
		_, _ = e.Client().Rest().CreateMessage(
			channelID,
			discord.MessageCreate{
				Content: fmt.Sprintf("%s hat die letzte Nachricht gelöscht. Zählerstand: %d", msg.Author.Mention(), lastNumber),
			},
		)
	}
}
