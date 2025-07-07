package main

import (
	"flag"
)

type Env struct {
	S Storage
}

func getArgs() *Env {
	addr := flag.String("host", "", "redis server host")
	if *addr == "" {
		*addr = "localhost:7379"
	}

	pass := flag.String("pass", "", "redis server password")
	if *pass == "" {
		*pass = "asdasdzxc"
	}

	db := flag.Int("db", 0, "redis server db")
	if *db == 0 {
		*db = 1
	}
	flag.Parse()

	r := NewRedisClient(*addr, *pass, *db)
	return &Env{S: r}
}
