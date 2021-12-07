package bedrockgopher

import (
	"github.com/thunder33345/diskoi"
)

// AddCommand registers the slash command to the bot. If the bot is not ready yet, it will be added to a queue which
// gets processed once the bot is ready.
func (b *Bot) AddCommand(cmd diskoi.Command) {
	if b.session.State.User == nil {
		b.commands = append(b.commands, cmd)
		return
	}
	b.diskoi.AddGuildCommand(b.guildID, cmd)
}

// SyncCommands unregisters existing commands and registers all commands in the queue to ensure there is no dead
// commands left on the guild.
func (b *Bot) SyncCommands() error {
	if err := b.unregisterCommands(); err != nil {
		return err
	}

	for _, cmd := range b.commands {
		b.logger.Debugf("registering command %s", cmd.Name())
		b.diskoi.AddGuildCommand(b.guildID, cmd)
	}
	b.commands = make([]diskoi.Command, 0)

	return b.diskoi.SyncCommands()
}

// unregisterCommands unregisters all the existing slash commands in the guild that were registered by the bot.
func (b *Bot) unregisterCommands() error {
	cmds, err := b.session.ApplicationCommands(b.session.State.User.ID, b.guildID)
	if err != nil {
		return err
	}
	for _, cmd := range cmds {
		b.logger.Debugf("unregistering command %s", cmd.Name)
		err = b.session.ApplicationCommandDelete(b.session.State.User.ID, b.guildID, cmd.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
