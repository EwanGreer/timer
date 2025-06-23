/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"log/slog"
	"os"

	"github.com/EwanGreer/timer/cmd"
)

func main() {
	level := slog.LevelInfo
	if env := os.Getenv("ENV"); env == "development" {
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level, AddSource: true})))

	cmd.Execute()
}
