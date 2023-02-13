package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/df-mc/bedrockgopher/module/faq"
	"github.com/df-mc/bedrockgopher/module/update_tracker"
	"golang.org/x/exp/slog"
)

func main() {
	textHandler := slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelInfo,
		ReplaceAttr: nil,
	}.NewTextHandler(os.Stdout)

	logger := slog.New(textHandler)
	logger.Info("App started")

	cfg, err := ReadConfig()
	if err != nil {
		logger.Warn("Failed to load config", slog.String("err", err.Error()))
		return
	}
	b := New(NewLevelLogger(slog.LevelInfo, logger.Handler()).With(slog.String("module", "bot")), cfg.Token)

	// todo redo/improve this part
	// maybe condense everything into it somehow
	b.SetupBot()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := b.Client.OpenGateway(ctx); err != nil {
		b.Logger.Error("Failed to connect to gateway: %s", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10)
		defer cancel()
		b.Client.Close(ctx)
	}()

	faqTracker := faq.New(NewLevelLogger(slog.LevelWarn, logger.Handler()).With(slog.String("module", "tracker")),
		cfg.Faq.GuildID, cfg.Faq.ChannelID)
	cmds := faqTracker.MountHandler(b.Router, "faq", "manage-faq")
	_, err = b.Client.Rest().SetGuildCommands(b.Client.ApplicationID(), cfg.Faq.GuildID, cmds)
	if err != nil {
		b.Logger.Error("Failed to register commands", err)
	}
	go faqTracker.StartIndexing(ctx, b.Client, time.Minute*1)

	tracker := update_tracker.New(
		NewLevelLogger(slog.LevelWarn, logger.Handler()).With(slog.String("module", "tracker")),
		ctx, b.Client, cfg.Tracker.ChannelID)
	go tracker.StartUpdateTicking(time.Minute * cfg.Tracker.Interval)

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s
	b.Logger.Info("Got signal, shutting down...")
	cancel()
	time.Sleep(time.Millisecond * 250)
	b.Logger.Info("Bye!")
}
