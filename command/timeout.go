package command

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/thunder33345/diskoi"
	"time"
)

type timeoutArgs struct {
	User     discordgo.User `diskoi:"description:The user to timeout;required"`
	Duration string         `diskoi:"description:How long should they be timed out for;required"`
}

func Timeout() diskoi.Command {
	cmd := diskoi.MustNewExecutor("timeout", "timeout a user", timeout)
	_ = cmd.SetChain(diskoi.NewChain(func(next diskoi.Middleware) diskoi.Middleware {
		return func(r diskoi.Request) error {
			member := r.Interaction().Member
			if member.Permissions&discordgo.PermissionModerateMembers > 0 {
				return next(r)
			}

			return r.Session().InteractionRespond(r.Interaction().Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "You do not have the required permissions to run this command", Flags: 1 << 6},
			})
		}
	}))
	return cmd
}

func timeout(s *discordgo.Session, i *discordgo.InteractionCreate, args timeoutArgs) error {
	d, err := time.ParseDuration(args.Duration)
	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: fmt.Sprintf(`Failed to parse "%s" as duration: %v`, args.Duration, err), Flags: 1 << 6},
		})
	}
	t := time.Now().Add(d)

	err = s.GuildMemberTimeout(i.GuildID, args.User.ID, &t)
	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: fmt.Sprintf(`Failed timeout %s for "%s": %v`, args.User.Username, d.String(), err), Flags: 1 << 6},
		})
	}
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: fmt.Sprintf(`Timed out %s for "%s"`, args.User.Username, d.String()), Flags: 1 << 6},
	})
}
