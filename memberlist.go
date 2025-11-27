package main

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/samber/lo"
)

var roleHierarchy = []snowflake.ID{
	snowflake.ID(1430892242440159342), // Guild GM
	snowflake.ID(1430892337323704392), // Rep. Guild GM
	snowflake.ID(1430892712290422866), // Chief
	snowflake.ID(1430894199393353861), // Sergeant
	snowflake.ID(1430899443586171023), // Cadet
}

// Spezial-Rollen für Div-Leader Mapping (Rolle -> Titel im Chat)
var divLeaderRoles = map[snowflake.ID]string{
	snowflake.ID(1430925917278044250): "[Div-Leader] Farming",
	snowflake.ID(1430925981236854937): "[Div-Leader] Building",
	snowflake.ID(1430925991529943221): "[Div-Leader] Mining",
	snowflake.ID(1430925984055693422): "[Div-Leader] Combat",
	snowflake.ID(1430925986395979776): "[Div-Leader] Redstone",
}

const memberListChannelID = 1443475483839955035

func updateMemberList(client bot.Client, guildID snowflake.ID) {
	rest := client.Rest()
	messageID := snowflake.ID(0)
	if messages, err := rest.GetMessages(memberListChannelID, 0, 0, 0, 20); err == nil && len(messages) > 0 {
		if message, found := lo.Find(messages, func(msg discord.Message) bool {
			return msg.Author.ID == client.ID()
		}); found {
			messageID = message.ID
			slog.Info("[MemberList] Found existing message, updating...", "messageID", messageID)
		} else {
			message, _ := rest.CreateMessage(memberListChannelID, discord.MessageCreate{Content: "."})
			messageID = message.ID
			slog.Info("[MemberList] Created new message", "messageID", messageID)
		}
	} else {
		message, _ := rest.CreateMessage(memberListChannelID, discord.MessageCreate{Content: "."})
		messageID = message.ID
		slog.Info("[MemberList] Created new message", "messageID", messageID)
	}
	var sb strings.Builder
	sb.WriteString("Mitgliederliste von DragonsGuild\n\n")

	for _, hierarchyRoleID := range roleHierarchy {
		roleMembers := make([]discord.Member, 0)
		client.Caches().MemberCache().ForEach(func(groupID snowflake.ID, member discord.Member) {
			if groupID == guildID && lo.Contains(member.RoleIDs, hierarchyRoleID) {
				roleMembers = append(roleMembers, member)
			}
		})

		sb.WriteString(fmt.Sprintf("----- <@&%s> -----\n\n", hierarchyRoleID))

		if len(roleMembers) == 0 {
			sb.WriteString("Keine Mitglieder\n\n")
			continue
		}

		sort.Slice(roleMembers, func(i, j int) bool { // sortieren (nach Nickname/Username)
			return strings.ToLower(roleMembers[i].EffectiveName()) < strings.ToLower(roleMembers[j].EffectiveName())
		})

		for _, member := range roleMembers {
			line := fmt.Sprintf("<@%s>", member.User.ID)

			foundDivRole, hasDivRole := lo.Find(member.RoleIDs, func(rid snowflake.ID) bool {
				_, exists := divLeaderRoles[rid]
				return exists
			})

			if hasDivRole {
				line = fmt.Sprintf("<@&%s> > <@%s>", foundDivRole, member.User.ID)
			}

			sb.WriteString(line + "\n\n")
		}
	}

	sb.WriteString("-----")

	content := sb.String()
	_, err := client.Rest().UpdateMessage(memberListChannelID, messageID, discord.MessageUpdate{Content: &content})
	if err != nil {
		slog.Error("failed to update message", "error", err)
	} else {
		slog.Info("Member list updated successfully")
	}
}
