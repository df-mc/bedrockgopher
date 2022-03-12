package command

import (
	"github.com/bwmarrin/discordgo"
	"github.com/thunder33345/diskoi"
)

func checkPerms(flag int64) diskoi.Chainer {
	return func(next diskoi.Middleware) diskoi.Middleware {
		return func(r diskoi.Request) error {
			member := r.Interaction().Member
			if member.Permissions&flag > 0 {
				return next(r)
			}

			return r.Session().InteractionRespond(r.Interaction().Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "You do not have the required permissions to run this command", Flags: 1 << 6},
			})
		}
	}
}
