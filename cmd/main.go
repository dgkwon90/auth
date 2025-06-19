// Package main is the entry point for the authentication service application.
package main

import (
	"auth/internal/config"
	"auth/internal/server"
)

func main() {
	cfg := config.LoadConfig()
	server := server.NewServer(cfg)
	defer server.Close()
	if err := server.App.Listen(":" + cfg.Port); err != nil {
		panic(err)
	}
}
