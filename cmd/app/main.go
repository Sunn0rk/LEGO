package main

import (
	telegram "github.com/Sunn0rk/LEGO/pkg"
)

func main() {
	var bot = telegram.Telegram_bot{}
	bot.Start_polling()
}
