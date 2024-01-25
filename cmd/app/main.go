package main

import (
	"log"
	"os"
	"regexp"

	telegram "github.com/Sunn0rk/LEGO/pkg"
	"github.com/joho/godotenv"
)

func main() {
	loadEnv()
	var bot_token = os.Getenv("TG_API_BOT_TOKEN")
	var bot = telegram.Telegram_bot{}
	bot.Start_polling(bot_token)
}

func loadEnv() {
	const projectDirName = "LEGO"
	projectName := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	currentWorkDirectory, _ := os.Getwd()
	rootPath := projectName.Find([]byte(currentWorkDirectory))

	err := godotenv.Load(string(rootPath) + `/.env`)

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}
