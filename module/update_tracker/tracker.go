package update_tracker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
	"golang.org/x/net/context"
)

const updateURL = "https://itunes.apple.com/lookup?bundleId=com.mojang.minecraftpe&time=%v"

func New(logger *slog.Logger, ctx context.Context, client bot.Client, channel snowflake.ID) *Tracker {
	return &Tracker{
		knownVersions: make([]string, 0, 4),
		log:           logger,
		context:       ctx,
		c:             client,
		channel:       channel,
	}
}

type Tracker struct {
	// knownVersions is a slice of all known versions of Minecraft.
	knownVersions []string
	log           *slog.Logger
	context       context.Context
	c             bot.Client
	channel       snowflake.ID
}

func (t *Tracker) StartUpdateTicking(interval time.Duration) {
	ti := time.NewTicker(interval)
	defer ti.Stop()
	t.log.Log(slog.LevelDebug, "Starting update ticker")
	t.check()
	for {
		select {
		case <-ti.C:
			t.check()
		case <-t.context.Done():
			t.log.Log(slog.LevelDebug, "Got context cancel signal")
			return
		}
	}
}

func (t *Tracker) check() {
	t.log.Log(slog.LevelDebug, "Checking for updates")
	resp, err := http.Get(fmt.Sprintf(updateURL, time.Now().UnixMilli()))
	if err != nil {
		t.log.Error("failed to check for updates", err)
		return
	}
	var m map[string]any
	if err = json.NewDecoder(resp.Body).Decode(&m); err != nil {
		t.log.Error("failed to decode response", err)
		_ = resp.Body.Close()
		return
	}
	_ = resp.Body.Close()
	if m["resultCount"].(float64) < 1 {
		// we got 0 results, this should not happen
		t.log.Log(slog.LevelWarn, "JSON result[resultCount] < 1, expecting at least 1")
		t.log.LogAttrs(slog.LevelDebug, "JSON result dump", slog.Attr{Key: "json", Value: slog.AnyValue(m)})
		return
	}
	version := m["results"].([]any)[0].(map[string]any)["version"].(string)
	releaseDate := m["results"].([]any)[0].(map[string]any)["currentVersionReleaseDate"].(string)
	releaseTime, _ := time.Parse(time.RFC3339, releaseDate)
	if len(t.knownVersions) == 0 {
		if releaseTime.IsZero() {
			t.log.LogAttrs(slog.LevelDebug, "Cant parse time, assuming version is old",
				slog.String("version", version))
			// Since we cannot get the time, we assume that the latest version is not the first query,
			// so set the current version to this version.
			t.knownVersions = append(t.knownVersions, version)
			return
		}
		timeSub := time.Now().Sub(releaseTime)
		if timeSub > time.Minute*15 {
			t.log.LogAttrs(slog.LevelDebug, "Parsed time, assuming version is old",
				slog.String("version", version), slog.Time("time", releaseTime), slog.Duration("duration", timeSub))
			// We assume this is an old version if it has been released for over 15 min
			t.knownVersions = append(t.knownVersions, version)
			return
		}
	}

	if slices.Contains(t.knownVersions, version) {
		t.log.LogAttrs(slog.LevelDebug, "Got old version",
			slog.String("version", version))
		// We don't care about the current version.
		return
	}

	t.log.LogAttrs(slog.LevelInfo, "New version available", slog.String("version", version))
	t.knownVersions = append(t.knownVersions, version)

	// Notify the new version
	_, err = t.c.Rest().CreateMessage(t.channel, discord.MessageCreate{
		Content: fmt.Sprintf("Minecraft v%s is now available at %s! @here", version, releaseTime.Format(time.RFC3339)),
	})
	if err != nil {
		t.log.Error("error sending update message", err)
	}
}
