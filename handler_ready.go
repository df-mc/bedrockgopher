package bedrockgopher

import "github.com/bwmarrin/discordgo"

// HandleReady handles the ready event.
func (b *Bot) HandleReady(*discordgo.Session, *discordgo.Ready) {
	if err := b.SyncCommands(); err != nil {
		b.logger.Fatalf("failed to sync commands: %v", err)
	}

	b.logger.Infof("Bot ready")
}
