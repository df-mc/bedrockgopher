package main

import (
	"fmt"

	"golang.org/x/exp/slog"
)

type shimLogger struct {
	log *slog.Logger
}

func (s *shimLogger) Trace(args ...any) {
	s.log.Log(slog.LevelDebug-2, fmt.Sprint(args))
}

func (s *shimLogger) Debug(args ...any) {
	s.log.Log(slog.LevelDebug, fmt.Sprint(args))
}

func (s *shimLogger) Info(args ...any) {
	s.log.Log(slog.LevelInfo, fmt.Sprint(args))
}

func (s *shimLogger) Warn(args ...any) {
	s.log.Log(slog.LevelWarn, fmt.Sprint(args))
}

func (s *shimLogger) Error(args ...any) {
	s.log.Log(slog.LevelError, fmt.Sprint(args))
}

func (s *shimLogger) Fatal(args ...any) {
	s.log.Log(slog.LevelError+1, fmt.Sprint(args))
}

func (s *shimLogger) Panic(args ...any) {
	s.log.Log(slog.LevelError+2, fmt.Sprint(args))
}

func (s *shimLogger) Tracef(format string, args ...any) {
	s.log.Log(slog.LevelDebug-2, fmt.Sprintf(format, args))
}

func (s *shimLogger) Debugf(format string, args ...any) {
	s.log.Log(slog.LevelDebug, fmt.Sprintf(format, args))
}

func (s *shimLogger) Infof(format string, args ...any) {
	s.log.Log(slog.LevelInfo, fmt.Sprintf(format, args))
}

func (s *shimLogger) Warnf(format string, args ...any) {
	s.log.Log(slog.LevelWarn, fmt.Sprintf(format, args))
}

func (s *shimLogger) Errorf(format string, args ...any) {
	s.log.Log(slog.LevelError, fmt.Sprintf(format, args))
}

func (s *shimLogger) Fatalf(format string, args ...any) {
	s.log.Log(slog.LevelError+1, fmt.Sprintf(format, args))
}

func (s *shimLogger) Panicf(format string, args ...any) {
	s.log.Log(slog.LevelError+2, fmt.Sprintf(format, args))
}
