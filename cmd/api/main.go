package main

import (
	"log"

	"github.com/benshields/messagebox/internal/api"
)

func main() {
	err := api.Start("")
	log.Println(err)
}
