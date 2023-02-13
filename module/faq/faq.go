package faq

import (
	"context"
	"sync"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"golang.org/x/exp/slog"
)

type Faq struct {
	log          *slog.Logger
	faqGuildID   snowflake.ID
	faqChannelID snowflake.ID

	indexMu     sync.RWMutex
	index       []entry
	indexString []string
	indexReq    chan struct{}
}

func New(logger *slog.Logger, guild, channel snowflake.ID) *Faq {
	f := Faq{
		log:          logger,
		faqGuildID:   guild,
		faqChannelID: channel,
		indexReq:     make(chan struct{}, 1),
	}
	return &f
}

func (f *Faq) StartIndexing(ctx context.Context, client bot.Client, interval time.Duration) {
	f.log.LogAttrs(slog.LevelDebug, "Indexer started", slog.Duration("interval", interval))

	f.requestIndex()

	t := time.NewTicker(interval)
	j := func() {
		err := f.Index(client)
		if err != nil {
			f.log.Error("Error while indexing", err)
		}
	}
	for {
		select {
		case <-ctx.Done():
			f.log.LogAttrs(slog.LevelDebug, "Indexer exiting")
			return
		case <-f.indexReq:
			f.log.LogAttrs(slog.LevelDebug, "Received index request")
			j()
			t.Reset(interval)
		case <-t.C:
			f.log.LogAttrs(slog.LevelDebug, "Received indexer tick")
			j()
		}
	}
}

func (f *Faq) requestIndex() {
	select {
	case f.indexReq <- struct{}{}:
	default:
	}
}

func (f *Faq) MountHandler(h handler.Router, searchName, managePrefix string) []discord.ApplicationCommandCreate {
	h.HandleCommand("/"+searchName, f.HandleSearch)
	h.HandleAutocomplete("/"+searchName, f.HandleCompleteSearch)

	h.Route("/"+managePrefix, func(h handler.Router) {
		h.HandleCommand("/create", f.HandleCreate)
		h.HandleCommand("/edit", f.HandleEdit)
		h.HandleCommand("/delete", f.HandleDelete)
		h.HandleCommand("/toc", f.HandleCreateTOC)
		h.HandleCommand("/index", f.HandleIndex)
	})
	h.HandleComponent("request-toc", f.ComponentRespondTOC)

	return []discord.ApplicationCommandCreate{CreateSlashSearch(searchName), CreateSlashManage(managePrefix)}
}

func respondErr(r events.InteractionResponderFunc, desc string, err error) error {
	m := discord.NewMessageCreateBuilder()
	e := discord.NewEmbedBuilder().SetTitle("Error").SetDescription(desc).SetColor(15548997)
	if err != nil {
		e.SetField(0, "Error", err.Error(), false)
	}
	m.SetEmbeds(e.Embed)
	m.SetEphemeral(true)
	return r(discord.InteractionResponseTypeCreateMessage, m.MessageCreate)
}

func createEmbed(title, description string) *discord.EmbedBuilder {
	return discord.NewEmbedBuilder().SetTitle(title).SetDescription(description).SetColor(0x00add8)
}

func ptrOf[T any](t T) *T {
	return &t
}
