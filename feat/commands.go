package feat

import (
	. "HereBeDragons/common"
	"fmt"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func handleCommand(e *events.MessageCreate) (response string) {
	content := e.Message.Content
	client := e.Client()
	rest := client.Rest
	raw := strings.Split(content, " ")

	slog.Info("messageCreate: ", "raw", raw)
	args := raw[1:]
	switch raw[0][2:] {
	case "ping":
		response = "pong"
	case "setchannel":
		state.ChannelID = e.Message.ChannelID
		response = "Kanal wurde auf <#" + state.ChannelID.String() + "> gesetzt"
		if err := SaveState(state); err != nil {
			slog.Error("failed to persist state", "error", err)
		}
	case "help":
		response = "d!ping, d!setchannel, d!help, d!about, d!memberlist, d!purge <Anzahl>\n" +
			"Eine Person kann nicht zwei Nachrichten hintereinander senden."
	case "about":
		self, _ := client.Caches.SelfUser()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		uptime := time.Since(botStartTime)
		uptimeFmt := fmt.Sprintf("%d:%02d:%02d:%02d",
			int(uptime.Hours()/24), int(uptime.Hours())%24, int(uptime.Minutes())%60, int(uptime.Seconds())%60)

		embed := discord.NewEmbed().
			WithColor(0x9B4F96).
			AddField("**Dev**: ", "itsTyrion (<@!265038515375570944>)", false).
			AddField("**Ping**: ", fmt.Sprintf("%dms", client.Gateway.Latency().Milliseconds()), false).
			AddField("**Uptime**: ", uptimeFmt, false).
			AddField("**RAM**: ", fmt.Sprintf("%dMB alloc/%dMB sys", m.HeapAlloc/1024/1024, m.HeapSys/1024/1024), false).
			AddField("**Powered by**: ", runtime.Version()+", disgo "+disgo.Version+" and Deez Nuts", false).
			WithFooter("Made (with love) in Germany", "https://itstyrion.de/random/MC-Heart.png").
			WithAuthor("HereBeDragons", "https://youtu.be/qWNQUvIk954", *self.AvatarURL())
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
		if client.Caches.MemberPermissions(*e.Message.Member).Has(discord.PermissionAdministrator) {
			UpdateMemberList(client, *e.GuildID, snowflake.ID(config.MemberListChannelID))
			rest.AddReaction(e.ChannelID, e.Message.ID, "✅")
		}
	default:
		response = "Unbekannter Befehl. d!help"
	}
	return
}
