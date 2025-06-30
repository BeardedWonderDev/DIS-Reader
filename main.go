package main

import (
	"log"
	"os"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/BeardedWonderDev/DIS-Reader/ui"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := types.Config{
		Host: os.Getenv("DIS_HOST"),
	}

	err := ui.NewUI(&cfg).Run()
	if err != nil {
		log.Printf("fatal: %v", err)
		os.Exit(1)
	}
}
