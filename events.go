package main

import (
	"fmt"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

var (
	previousCounter    int          = 0
	countingChannelID  snowflake.ID = 0
	lastPersonCounting snowflake.ID = 0
	botStartTime       time.Time    = time.Now()
)

func messageCreate(e *events.MessageCreate) {
	if e.Message.Author.Bot || e.GuildID == nil {
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
			countingChannelID = e.Message.ChannelID
			response = "Kanal wurde auf <#" + countingChannelID.String() + "> gesetzt"
			if err := saveState(); err != nil {
				slog.Error("failed to persist state", "error", err)
			}
		case "help":
			response = "d!ping, d!setchannel, d!help, d!about, d!memberlist, d!purge <Anzahl>\n" +
				"Eine Person kann nicht zwei Nachrichten hintereinander senden."
		case "about":
			self, _ := client.Caches().SelfUser()
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			uptime := time.Since(botStartTime)
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
				SetAuthor("HereBeDragons", "https://youtu.be/qWNQUvIk954", *self.AvatarURL()).
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
				if messages, err := rest.GetMessages(e.ChannelID, 0, e.Message.ID, 0, number); err == nil {
					messageIDs := make([]snowflake.ID, 0, len(messages))
					for _, msg := range messages {
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
		case "memberlist":
			if client.Caches().MemberPermissions(*e.Message.Member).Has(discord.PermissionAdministrator) {
				updateMemberList(client, *e.GuildID)
				rest.AddReaction(e.ChannelID, e.Message.ID, "✅")
			}
		default:
			response = "Unbekannter Befehl. d!help"
		}
	} else {
		if e.Message.Author.ID == lastPersonCounting || countingChannelID != e.Message.ChannelID {
			return
		}
		// check if content is a number
		if number, err := strconv.Atoi(content); err == nil && number == previousCounter+1 {
			previousCounter = previousCounter + 1
			lastPersonCounting = e.Message.Author.ID
			if err = rest.AddReaction(e.ChannelID, e.Message.ID, "✅"); err != nil {
				slog.Error("failed to add reaction", "error", err)
			}
			if err := saveState(); err != nil {
				slog.Warn("messageCreate: failed to persist state", "error", err)
			}
		} else {
			slog.Info("messageCreate: content is not a number")
			previousCounter = 0
			response = fmt.Sprintf("%s hat die Strähne unterbrochen. :(", e.Message.Author.Mention())
			slog.Info("messageCreate: lastNumber", "lastNumber", previousCounter, "lastPerson", lastPersonCounting)
			if err := saveState(); err != nil {
				slog.Warn("messageCreate: failed to persist state", "error", err)
			}
		}
	}

	if response != "" {
		_, _ = client.Rest().CreateMessage(e.ChannelID, discord.MessageCreate{Content: response})
	}
}

func messageDelete(e *events.MessageDelete) {
	msg := e.Message
	if msg.Content == strconv.Itoa(previousCounter) {
		_, _ = e.Client().Rest().CreateMessage(
			countingChannelID,
			discord.MessageCreate{
				Content: fmt.Sprintf("%s hat die letzte Nachricht gelöscht. Zählerstand: %d", msg.Author.Mention(), previousCounter),
			},
		)
	}
}
