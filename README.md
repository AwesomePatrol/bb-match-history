# Biter Battles Match History

Hosted on: http://bb.patrol.neutrino.re/recent/ and http://bb-raven.patrol.neutrino.re/recent/

## Run

```bash
$ cd assets
$ hugo
$ cd ..
$ go run main.go -addr :8080 -db path_to.db -t discord_bot_token
```

You can start it without Discord bot, by ommiting `-t` argument.
