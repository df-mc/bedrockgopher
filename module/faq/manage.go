package faq

import (
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
	"github.com/disgoorg/snowflake/v2"
)

func CreateSlashManage(name string) discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:        name,
		Description: "Manage FAQ",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "create",
				Description: "Create FAQ",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "title",
						Description: "faq title",
						Required:    true,
					}, discord.ApplicationCommandOptionString{
						Name:        "description",
						Description: "faq description",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "edit",
				Description: "Edit FAQ",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "id",
						Description: "message id",
						Required:    true,
					},
					discord.ApplicationCommandOptionString{
						Name:        "title",
						Description: "faq title",
						Required:    true,
					},
					discord.ApplicationCommandOptionString{
						Name:        "description",
						Description: "faq description",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "delete",
				Description: "Delete FAQ",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "id",
						Description: "message id",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "toc",
				Description: "Create TOC button",
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "index",
				Description: "Invoke manual reindexing",
			},
		},
		DefaultMemberPermissions: ptrOf(json.NewNullable[discord.Permissions](discord.PermissionManageChannels + discord.PermissionManageMessages)),
		DMPermission:             ptrOf(false),
	}
}

func (f *Faq) ManageHandler() handler.Router {
	r := handler.New()
	r.HandleCommand("create", f.HandleCreate)
	r.HandleCommand("edit", f.HandleEdit)
	r.HandleCommand("delete", f.HandleDelete)
	return r
}

func (f *Faq) HandleCreate(e *handler.CommandEvent) error {
	// todo add custom embed color? lol
	_, err := e.Client().Rest().CreateMessage(e.ChannelID(), discord.NewMessageCreateBuilder().SetEmbeds(
		createEmbed(e.SlashCommandInteractionData().String("title"), formatDescription(e.SlashCommandInteractionData(), "description")).Embed,
	).MessageCreate)
	if err != nil {
		return respondErr(e.Respond, "Failed to create message", err)
	}
	f.requestIndex()
	return e.CreateMessage(discord.NewMessageCreateBuilder().SetContent("Message created.").SetEphemeral(true).MessageCreate)
}

func (f *Faq) HandleEdit(e *handler.CommandEvent) error {
	idStr := e.SlashCommandInteractionData().String("id")
	id, err := snowflake.Parse(idStr)
	if err != nil {
		return respondErr(e.Respond, "Failed to parse messageID", err)
	}

	_, err = e.Client().Rest().UpdateMessage(e.ChannelID(), id, discord.NewMessageUpdateBuilder().SetEmbeds(
		createEmbed(e.SlashCommandInteractionData().String("title"), formatDescription(e.SlashCommandInteractionData(), "description")).Embed,
	).MessageUpdate)

	if err != nil {
		return respondErr(e.Respond, "Failed to update  messageID", err)
	}
	f.requestIndex()
	return e.CreateMessage(discord.NewMessageCreateBuilder().SetContent("Message updated.").SetEphemeral(true).MessageCreate)
}

func formatDescription(data discord.SlashCommandInteractionData, key string) string {
	str := data.String(key)
	return strings.Replace(str, "\\n", "\n", -1)
}

func (f *Faq) HandleDelete(e *handler.CommandEvent) error {
	idStr := e.SlashCommandInteractionData().String("id")
	id, err := snowflake.Parse(idStr)
	if err != nil {
		return respondErr(e.Respond, "Failed to parse messageID", err)
	}
	err = e.Client().Rest().DeleteMessage(e.ChannelID(), id)
	if err != nil {
		return respondErr(e.Respond, "Failed to delete messageID", err)
	}
	f.requestIndex()
	return e.CreateMessage(discord.NewMessageCreateBuilder().SetContent("Message deleted.").SetEphemeral(true).MessageCreate)
}

func (f *Faq) HandleCreateTOC(e *handler.CommandEvent) error {
	b := discord.NewMessageCreateBuilder().SetContent("View TOC")
	b.AddContainerComponents(discord.ActionRowComponent{
		discord.ButtonComponent{
			Style:    1,
			Label:    "View TOC",
			CustomID: "request-toc",
		},
	})
	_, err := e.Client().Rest().CreateMessage(e.ChannelID(), b.MessageCreate)
	if err != nil {
		return respondErr(e.Respond, "Failed to create TOC button", err)
	}
	return e.Respond(discord.InteractionResponseTypeCreateMessage,
		discord.NewMessageCreateBuilder().SetContent("Created TOC button").SetEphemeral(true).MessageCreate)
}

func (f *Faq) HandleIndex(e *handler.CommandEvent) error {
	err := e.DeferCreateMessage(true)
	if err != nil {
		return err
	}

	err = f.Index(e.Client())
	if err != nil {
		return respondErr(e.Respond, "Error indexing channel", err)
	}
	_, err = e.CreateFollowupMessage(discord.NewMessageCreateBuilder().SetEphemeral(true).
		SetContent("Successfully indexed faq channel.").MessageCreate)
	return err
}
