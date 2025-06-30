package main

import (
	"log"
	"os"

	"github.com/BeardedWonderDev/DIS-Reader/cmd/ui"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := types.Config{
		JavaPath: os.Getenv("JAVA_PATH"),
		JarPath:  os.Getenv("JAR_PATH"),
		Host:     os.Getenv("DIS_HOST"),
		JDBCPort: os.Getenv("JDBC_PORT"),
	}

	err := ui.NewUI(&cfg).Run()
	if err != nil {
		log.Printf("fatal: %v", err)
		os.Exit(1)
	}
}
