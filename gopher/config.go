package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/disgoorg/snowflake/v2"
)

func ReadConfig() (Config, error) {
	by, err := os.ReadFile("./config.json")
	if err != nil {
		return Config{}, err
	}
	var c Config
	err = json.Unmarshal(by, &c)
	return c, err
}

type Config struct {
	Token   string        `json:"token"`
	Faq     FaqConfig     `json:"faq"`
	Tracker TrackerConfig `json:"tracker"`
}

type FaqConfig struct {
	GuildID   snowflake.ID `json:"guildID"`
	ChannelID snowflake.ID `json:"channelID"`
}

type TrackerConfig struct {
	ChannelID snowflake.ID  `json:"channelID"`
	Interval  time.Duration `json:"interval"`
}
