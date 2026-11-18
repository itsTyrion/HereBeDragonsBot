package feat

import (
	"HereBeDragons/common"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var cooldown = common.NewCooldown(30 * time.Second)

var eightBallAnswers = [20]string{
	// affirmative
	"It is certain", "It is decidedly so", "Without a doubt", "Yes definitely", "You may rely on it",
	"As I see it, yes", "Most likely", "Outlook good", "Yes", "Signs point to yes",
	// non-committal
	"Reply hazy, try again", "Ask again later", "Better not tell you now",
	"Cannot predict now", "Concentrate and ask again",
	// negative
	"Don’t count on it", "My reply is no", "My sources say no", "Outlook not so good", "Very doubtful",
}
var sourceGifs = [3]string{
	"https://tenor.com/view/youre-just-gonna-have-to-trust-me-okay-billy-mcbride-goliath-have-faith-trust-me-on-this-gif-23202887",
	"https://tenor.com/view/my-source-is-when-the-source-is-reliable-reliable-senator-armstrong-metal-gear-rising-gif-27116155",
	"https://tenor.com/view/metal-gear-rising-metal-gear-rising-revengeance-senator-armstrong-revengeance-i-made-it-the-fuck-up-gif-25029602",
}

var funnyPattern1 *regexp.Regexp
var funnyPattern2 *regexp.Regexp

func readyChat(e *events.Ready) {
	cid := e.Client().ID().String()
	funnyPattern1 = regexp.MustCompile(`(<@!?` + cid + `>|@gork|@grok) +(is) (this|that) (true|real|cap)\??`)
	funnyPattern2 = regexp.MustCompile(`(<@!?` + cid + `>|@gork|@grok) +source\\??`)
}

func handleChatInteract(e *events.MessageCreate) {
	if e.Message.Author.Bot || e.GuildID == nil {
		return
	}
	m := e.Message

	content := strings.ToLower(m.Content)

	if funnyPattern1.MatchString(content) && cooldown.Add(m.Author.ID) {
		reply := eightBallAnswers[rand.Intn(20)]
		e.Client().Rest.CreateMessage(e.ChannelID, discord.MessageCreate{
			Content:          reply,
			MessageReference: m.MessageReference,
		})
	} else if funnyPattern2.MatchString(content) && cooldown.Add(m.Author.ID) {
		reply := sourceGifs[rand.Intn(3)]
		e.Client().Rest.CreateMessage(e.ChannelID, discord.MessageCreate{
			Content:          reply,
			MessageReference: m.MessageReference,
		})
	}
}
