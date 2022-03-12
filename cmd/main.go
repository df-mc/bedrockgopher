package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/df-mc/bedrockgopher"
	"github.com/df-mc/bedrockgopher/command"
	"github.com/sirupsen/logrus"
)

var (
	// Token is the token for the bot provided discord, and should be set during the building process.
	// For local development, manually set it to the token but remember to remove it before committing.
	Token string

	// guildID is the identifier of the Bedrock Gophers guild. The bot is designed to be used in one guild specifically
	// due to the limitations of registering slash commands globally. It also makes it easier to manage things such as
	// channels and roles as they can be controlled manually.
	guildID string = "623638955262345216"
)

func main() {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{ForceColors: true}
	log.Level = logrus.DebugLevel

	log.Infof("starting bot...")
	bot, err := bedrockgopher.New(log, Token, guildID)
	if err != nil {
		log.Fatal(err)
	}

	bot.Intents(discordgo.IntentsGuildIntegrations, discordgo.IntentsGuildMessages)

	bot.Session().AddHandlerOnce(bot.HandleReady)

	bot.AddCommand(command.FAQ(bot))
	bot.AddCommand(command.Timeout())

	if err := bot.Run(); err != nil {
		log.Fatal(err)
	}
	log.Infof("bot stopped")
}
