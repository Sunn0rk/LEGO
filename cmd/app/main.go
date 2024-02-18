package main

import (
	"database/sql"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

var gBot *tgbotapi.BotAPI
var gToken string
var db *sql.DB
var err error

type Sets struct {
	Results []struct {
		Part struct {
			PartNum string `json:"part_num"`
			Name    string `json:"name"`
		} `json:"part"`
		Color struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"color"`
		Quantity int `json:"quantity"`
	} `json:"results"`
}

func init() {
	fmt.Println("1")
	TGBotConnect()
	if gBot, err = tgbotapi.NewBotAPI(gToken); err != nil {
		log.Panic(err)
	}
	db, err = DatabaseConnect()
	if err == nil {
		// defer db.Close()
	} else {
		log.Fatal(err)
	}
	err = CreateLegoTable(db, tablename)
}

func main() {

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT
	updates := gBot.GetUpdatesChan(updateConfig)

Loop:
	for update := range updates {

		switch update.Message.Text {

		case "/inventory":

			gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Добавить набор или убрать имеющийся?"))

			for update := range updates {
				switch update.Message.Text {

				case "/back":
					gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Вернулся"))
					continue Loop

				case "/addset":
					AddOrDeleteSet(&update, updates, "добавления", "добавлено", "add", db)
					continue Loop

				case "/deleteset":
					AddOrDeleteSet(&update, updates, "удалено", "удалено", "delete", db)
					continue Loop
				}
			}

		case "/compare":

			gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Введите номер набора для сравнения"))

			for update := range updates {

				switch update.Message.Text {

				case "/back":
					gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Вернулся"))
					continue Loop

				default:
					Compare(update.Message.Text, "add", db, tablename)
				}
			}
		default:

			gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Я съебался"))

		}

	}

}
