package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

// 
var gBot *tgbotapi.BotAPI
var gToken string
var db *sql.DB
var err error
// 

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
	_ = os.Setenv(TOKEN_NAME_IN_OS, "6842123718:AAGAhkDOdqUMTLuCzo4CkzPxXzpNil4VMj8")
	gToken = os.Getenv(TOKEN_NAME_IN_OS)
	var err error
	if gBot, err = tgbotapi.NewBotAPI(gToken); err != nil {
		log.Panic(err)
	}

	gBot.Debug = false
}

func main() {
	connStr := fmt.Sprintf("%s://%s:%s@%s:%s/lego?sslmode=disable", database, username, password, server, port)
	db, err = sql.Open(driverName, connStr)
	fmt.Println(db)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	CreateLegoTable(db, tablename)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT

	updates := gBot.GetUpdatesChan(updateConfig)



	
	for update := range updates {
		if Inventory(&update) {
			gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Добавить набор или убрать имеющийся?"))
			for update := range updates {
				if IsBack(&update) {
					break
				} else if AddSetInventory(&update) {
					gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Введите номер набора для добавления"))
					for update := range updates {
						if IsBack(&update) {
							break
						} else {
							fmt.Print(update.Message.Text)
							API_Connect(update.Message.Text, "add", db, tablename)
							PartMerger(db, tablename)
							gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "добавлено"))
						}
					}

				} else if DeleteSetInventory(&update) {
					gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Введите номер набора для удаления"))
					for update := range updates {
						if IsBack(&update) {
							break
						} else {
							fmt.Print(update.Message.Text)
							API_Connect(update.Message.Text, "delete", db, tablename)
							PartMerger(db, tablename)
							gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "удалено"))
						}
					}

				}
			}

		} else if IsCompare(&update) {
			gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Введите номер набора для сравнения"))
			for update := range updates {
				if IsBack(&update) {
					break
				} else {
					Compare(update.Message.Text, "add", db, tablename)
				}

			}
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			fmt.Println(update)
			fmt.Println(update.Message.Text)
			if _, err := gBot.Send(msg); err != nil {
				panic(err)
			}
		}

	}

}
