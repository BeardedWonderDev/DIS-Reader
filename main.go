package main

import (
	"os"

	"github.com/BeardedWonderDev/DIS-Reader/cmd"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := &types.Config{
		AppName:    os.Getenv("APP_NAME"),
		AppVersion: os.Getenv("APP_VERSION"),
		Host:       os.Getenv("DIS_HOST"),
	}

	cmd.StartUI(cfg)
}
