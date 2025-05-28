package main

import (
	"log"

	"github.com/Abdurahmanit/GroupProject/order-service/internal/app"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/app/config"
)

func main() {
	cfg := config.MustLoad()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	application.Run()
}
