package bedrockgopher

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"github.com/thunder33345/diskoi"
	"os"
	"os/signal"
	"syscall"
)

// Bot represents the Bedrock Gopher discord bot.
type Bot struct {
	logger *logrus.Logger

	diskoi   *diskoi.Diskoi
	commands []diskoi.Command

	session *discordgo.Session
	guildID string
}

// New creates a new Bot with the provided token, and creates a discord session.
func New(logger *logrus.Logger, token, guildID string) (*Bot, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord session: %s", err)
	}

	d := diskoi.NewDiskoi()
	d.SetErrorHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate, cmd diskoi.Command, err error) {
		logger.Errorf("%s#%s failed to run command '%s': %v", i.User.Username, i.User.Discriminator, cmd.Name(), err)
	})
	d.RegisterSession(s)

	b := &Bot{
		logger: logger,

		diskoi:  d,
		session: s,

		guildID: guildID,
	}
	return b, nil
}

// Logger returns the logger provided for the bot.
func (b *Bot) Logger() *logrus.Logger {
	return b.logger
}

// Session returns the discord session for the bot.
func (b *Bot) Session() *discordgo.Session {
	return b.session
}

// GuildID returns the designated guild ID for the bot to use as the main guild.
func (b *Bot) GuildID() string {
	return b.guildID
}

// Run attempts to open the discord session and run the bot until a signal interrupt is received.
func (b *Bot) Run() error {
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("failed to open discord session: %s", err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	if err := b.session.Close(); err != nil {
		return fmt.Errorf("failed to close discord session: %s", err)
	}
	return nil
}

// Intents registers all the provided intents for the bot.
func (b *Bot) Intents(intents ...discordgo.Intent) {
	for _, i := range intents {
		b.session.Identify.Intents |= i
	}
}
