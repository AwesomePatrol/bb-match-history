package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/awesomepatrol/bb-match-history/discord"
	"github.com/awesomepatrol/bb-match-history/server"
	"github.com/awesomepatrol/bb-match-history/stats"
)

var (
	token string
	DB    string
	addr  string
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.StringVar(&DB, "db", "", "Path to DB")
	flag.StringVar(&addr, "addr", ":8080", "Address of HTTP server")
	flag.Parse()
}

func main() {
	stats.OpenDB(DB)
	defer stats.CloseDB()

	if token == "" {
		log.Println("run without discord bot")
	} else {
		discord.OpenBot(token)
		defer discord.CloseBot()
	}
	update, err := stats.ShouldUpdateELO()
	if err != nil {
		log.Println("failed to check if ELO was updated:", err)
	} else if update {
		err := stats.UpdateELO()
		if err != nil {
			log.Println("elo update failed:", err)
		}
	}

	server.OpenHTTP(addr)

	log.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
