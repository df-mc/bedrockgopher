package command

import (
	"github.com/bwmarrin/discordgo"
	"github.com/df-mc/bedrockgopher"
	"github.com/thunder33345/diskoi"
)

type faqNew struct {
	Question string `diskoi:"description:Question"`
	Answer   string `diskoi:"description:Answer"`
}

type faqEdit struct {
	MessageID string `diskoi:"description:Message ID"`
	Question  string `diskoi:"description:Question"`
	Answer    string `diskoi:"description:Answer"`
}

// FAQ represents the /faq command. It allows a moderator to create and manage FAQs.
func FAQ(bot *bedrockgopher.Bot) diskoi.Command {
	newE := diskoi.MustNewExecutor("new", "Create a new FAQ entry", func(s *discordgo.Session, i *discordgo.InteractionCreate, args faqNew) error {
		embed := &discordgo.MessageEmbed{Color: 0x00add8, Title: args.Question, Description: args.Answer}
		if _, err := s.ChannelMessageSendEmbed(i.ChannelID, embed); err != nil {
			bot.Logger().Errorf("failed to send message: %v", err)
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "Failed to create a FAQ entry. Please try again later", Flags: 1 << 6},
			})
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "A new FAQ entry has been created!", Flags: 1 << 6},
		})
	})
	newE.MustSetRequired("Question", true)
	newE.MustSetRequired("Answer", true)

	editE := diskoi.MustNewExecutor("edit", "Edit a FAQ entry", func(s *discordgo.Session, i *discordgo.InteractionCreate, args faqEdit) error {
		message, err := s.ChannelMessage(i.ChannelID, args.MessageID)
		if err != nil || message.Author.ID != bot.Session().State.User.ID || len(message.Embeds) == 0 {
			bot.Logger().Errorf("failed to edit message: %v", err)
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "Message provided is not part of the FAQ", Flags: 1 << 6},
			})
		}

		embed := message.Embeds[0]
		if args.Question != "" {
			embed.Title = args.Question
		}
		if args.Answer != "" {
			embed.Description = args.Answer
		}

		if _, err := s.ChannelMessageEditEmbed(i.ChannelID, args.MessageID, embed); err != nil {
			bot.Logger().Errorf("failed to edit message: %v", err)
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "Failed to edit FAQ entry. Please try again later", Flags: 1 << 6},
			})
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "The requested faq entry has been updated!", Flags: 1 << 6},
		})
	})
	editE.MustSetRequired("MessageID", true)

	cg := diskoi.NewCommandGroup("faq", "Manage the guild's FAQ channel")
	cg.AddSubcommand(newE)
	cg.AddSubcommand(editE)

	cg.SetChain(diskoi.NewChain(func(next diskoi.Middleware) diskoi.Middleware {
		return func(r diskoi.Request) error {
			member := r.Interaction().Member
			if member.Permissions&discordgo.PermissionManageMessages > 0 {
				return next(r)
			}

			return r.Session().InteractionRespond(r.Interaction().Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "You do not have the required permissions to run this command", Flags: 1 << 6},
			})
		}
	}))

	return cg
}
