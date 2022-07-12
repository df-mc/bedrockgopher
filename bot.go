package bedrockgopher

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"github.com/thunder33345/diskoi"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Bot represents the Bedrock Gopher discord bot.
type Bot struct {
	logger *logrus.Logger

	diskoi   *diskoi.Diskoi
	commands []diskoi.Command

	session *discordgo.Session
	guildID string

	c chan struct{}
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
	go b.startUpdateTicking()
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
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	b.logger.Info("bedrock gopher is shutting down...")
	if err := b.session.Close(); err != nil {
		return fmt.Errorf("failed to close discord session: %s", err)
	}
	b.c <- struct{}{}
	b.logger.Info("bye!")
	return nil
}

// Intents registers all the provided intents for the bot.
func (b *Bot) Intents(intents ...discordgo.Intent) {
	for _, i := range intents {
		b.session.Identify.Intents |= i
	}
}

// updateURL is the URL to check for updates.
const updateURL = "https://itunes.apple.com/lookup?bundleId=com.mojang.minecraftpe&time=%v"

// currentVersion is the current version of Minecraft.
var currentVersion string

// startUpdateTicking starts a ticker which checks for new Minecraft updates every minute.
func (b *Bot) startUpdateTicking() {
	t := time.NewTicker(time.Second * 5)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			resp, err := http.Get(fmt.Sprintf(updateURL, time.Now().UnixMilli()))
			if err != nil {
				b.logger.Errorf("failed to check for updates: %s", err)
				continue
			}
			var m map[string]interface{}
			if err = json.NewDecoder(resp.Body).Decode(&m); err != nil {
				b.logger.Errorf("failed to decode response: %s", err)
				_ = resp.Body.Close()
				continue
			}
			_ = resp.Body.Close()
			if m["resultCount"].(float64) > 0 {
				version := m["results"].([]interface{})[0].(map[string]interface{})["version"].(string)
				if len(currentVersion) == 0 {
					// We can assume that the latest version is not the first query, so set the current version to this
					// version.
					currentVersion = version
				}
				if version == currentVersion {
					// We don't care about the current version.
					continue
				}
				_, err := b.session.ChannelMessageSend("671024455979630620", "Minecraft v"+version+" is now available! @here")
				if err != nil {
					b.logger.Errorf("failed to send update message: %s", err)
				}
				b.logger.Infof("new version available: v%s", version)
				currentVersion = version
			}
		case <-b.c:
			return
		}
	}
}
