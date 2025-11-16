package main

import (
	"context"
	"log"

	"PR/internal/app"
)

func main() {

	a, err := app.NewApp(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
