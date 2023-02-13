package main

import (
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/paginator"
	"golang.org/x/exp/slog"
	"golang.org/x/net/context"
)

func New(logger *slog.Logger, token string) *Bot {
	return &Bot{
		Logger:    logger,
		Paginator: paginator.New(),
		Token:     token,
		Router:    handler.New(),
	}
}

type Bot struct {
	Logger    *slog.Logger
	Client    bot.Client
	Paginator *paginator.Manager
	Router    handler.Router
	Token     string
}

func (b *Bot) SetupBot(listeners ...bot.EventListener) {
	var err error
	b.Client, err = disgo.New(b.Token,
		bot.WithLogger(&shimLogger{log: b.Logger}),
		bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildMessages, gateway.IntentMessageContent)),
		bot.WithCacheConfigOpts(cache.WithCaches(cache.FlagGuilds, cache.FlagMessages)),
		bot.WithEventListeners(b.Paginator, b.Router),
		bot.WithEventListeners(bot.NewListenerFunc(b.OnReady)),
		bot.WithEventListeners(listeners...),
	)

	if err != nil {
		b.Logger.Error("Failed to setup bot: ", err)
		panic(err)
	}
}

func (b *Bot) OnReady(_ *events.Ready) {
	b.Logger.Info("Bot connected")
	if err := b.Client.SetPresence(context.TODO(), gateway.WithListeningActivity("you"), gateway.WithOnlineStatus(discord.OnlineStatusOnline)); err != nil {
		b.Logger.Error("Failed to set presence", err)
	}
}
