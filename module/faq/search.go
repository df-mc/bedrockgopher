package faq

import (
	"fmt"
	"sort"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/exp/slog"
)

type entry struct {
	title       string
	description string
	id          snowflake.ID
}

func CreateSlashSearch(name string) discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:        name,
		Description: "Search FAQ",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:         "search",
				Description:  "Text to search for",
				Required:     true,
				Autocomplete: true,
			},
			discord.ApplicationCommandOptionBool{
				Name:        "ephemeral",
				Description: "Use ephemeral lookup",
			},
		},
	}
}

func (f *Faq) Index(client bot.Client) error {
	f.log.Debug("Indexing...")
	all := make([]discord.Message, 0, 50)
	index := make([]entry, 0, 50)
	// indexing the cache by creating a pager for the faq channel
	pager := client.Rest().GetMessagesPage(f.faqChannelID, snowflake.ID(0), 50)
	// fetch everything into a slice
	for pager.Next() {
		all = append(all, pager.Items...)
	}
	if pager.Err != rest.ErrNoMorePages {
		return fmt.Errorf("error indexing: %v", pager.Err)
	}

	for i := len(all) - 1; i >= 0; i-- {
		msg := all[i]
		// ignore messages that aren't sent by the bot, or has no embed
		if msg.Author.ID != client.ID() {
			continue
		}
		if len(msg.Embeds) != 1 {
			continue
		}
		e := msg.Embeds[0]
		index = append(index, entry{
			title:       e.Title,
			description: e.Description,
			id:          msg.ID,
		})
	}

	f.log.Log(slog.LevelDebug, "Done indexing", slog.Int("total", len(all)), slog.Int("matching", len(index)))

	f.indexMu.Lock()
	defer f.indexMu.Unlock()

	f.index = index
	f.indexString = make([]string, 0, len(index))
	for _, v := range index {
		f.indexString = append(f.indexString, v.title+" "+v.description)
	}
	return nil
}

var errorNoResult = "error-no-result"

func (f *Faq) HandleSearch(e *handler.CommandEvent) error {
	search := e.SlashCommandInteractionData().String("search")
	if search == errorNoResult {
		return respondErr(e.Respond, "There's no result", nil)
	}
	f.indexMu.RLock()
	defer f.indexMu.RUnlock()

	var result entry
	var found bool
	for _, i := range f.index {
		if i.title == search {
			// in a realistic use case, there shouldn't be duplicated titles
			result = i
			found = true
		}
	}
	if !found {
		r := f.searchAndSort(search)
		if len(r) <= 0 {
			return respondErr(e.Respond, "There's no result", nil)
		}
		result = f.index[r[0].OriginalIndex]
	}
	embed := createEmbed(result.title, result.description)
	b := discord.NewMessageCreateBuilder()
	b.SetEmbeds(embed.Embed)
	b.SetEphemeral(e.SlashCommandInteractionData().Bool("ephemeral"))

	return e.CreateMessage(b.MessageCreate)
}

func (f *Faq) HandleCompleteSearch(e *handler.AutocompleteEvent) error {
	f.indexMu.RLock()
	defer f.indexMu.RUnlock()
	r := f.searchAndSort(e.Data.String("search"))
	if len(r) <= 0 {
		return e.Respond(
			discord.InteractionResponseTypeApplicationCommandAutocompleteResult,
			discord.AutocompleteResult{
				Choices: []discord.AutocompleteChoice{
					discord.AutocompleteChoiceString{
						Name:  "No result found!",
						Value: errorNoResult,
					},
				},
			},
		)
	}
	sorted := make([]discord.AutocompleteChoice, 0, len(r))

	for _, res := range r {
		choice := discord.AutocompleteChoiceString{
			Name:  f.index[res.OriginalIndex].title,
			Value: f.index[res.OriginalIndex].title,
		}
		sorted = append(sorted, choice)
	}

	return e.Respond(
		discord.InteractionResponseTypeApplicationCommandAutocompleteResult,
		discord.AutocompleteResult{
			Choices: sorted,
		},
	)
}

func (f *Faq) searchAndSort(keyword string) fuzzy.Ranks {
	r := fuzzy.RankFind(keyword, f.indexString)
	if len(r) <= 0 {
		return nil
	}
	sort.Sort(r)
	return r
}

func (f *Faq) ComponentRespondTOC(e *handler.ComponentEvent) error {
	err := e.DeferCreateMessage(true)
	if err != nil {
		f.log.Warn("failed at defer create")
		return err
	}
	var pieces []string
	var builder strings.Builder
	for i, v := range f.index {
		cur := fmt.Sprintf("%d. **[%s](https://discord.com/channels/%v/%v/%v)**\n", i+1, v.title, f.faqGuildID, f.faqChannelID, v.id)
		if builder.Len()+len(cur) < 2000 {
			builder.WriteString(cur)
			continue
		}
		pieces = append(pieces, builder.String())
		builder.Reset()
		builder.WriteString(cur)
	}
	pieces = append(pieces, builder.String())
	for _, p := range pieces {
		b := discord.NewMessageCreateBuilder().SetEphemeral(true)
		em := discord.Embed{}
		em.Title = "FAQ Table of Content"
		em.Description = p
		b.SetEmbeds(em)

		_, err = e.CreateFollowupMessage(b.MessageCreate)
		if err != nil {
			f.log.Warn("failed at create message")
			return err
		}
	}
	return nil
}
