package main

import (
	"log"
	"time"
)

func main() {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("hello world")
	}
}
