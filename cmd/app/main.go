package main

import (
	"database/sql"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

var tablename string
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
	TGBotConnect()
	if gBot, err = tgbotapi.NewBotAPI(gToken); err != nil {
		log.Panic(err)
	}
}

func main() {

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = UPDATE_CONFIG_TIMEOUT
	updates := gBot.GetUpdatesChan(updateConfig)

expectation:
	for update := range updates {

		switch update.Message.Text {
		case "/start":

			gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Начали"))

			db, err = DatabaseConnect()
			if err == nil {
			} else {
				log.Panic(err)
			}

			tablename = fmt.Sprintf("InventoryTable_%d", update.Message.Chat.ID)
			err = CreateLegoTable(db, tablename)
			if err == nil {
			} else {
				log.Panic(err)
			}

			tablename = fmt.Sprintf("SetHistoryTable_%d", update.Message.Chat.ID)
			err = CreateSetHistory(db, tablename)
			if err == nil {
			} else {
				log.Panic(err)
			}

			goto MainLoop

		default:
			// gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Я не съебался"))
			continue
		}
	}

MainLoop:
	for update := range updates {

		switch update.Message.Text {
		case "/stop":
			gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Закончили"))
			goto expectation

		case "/inventory":
			gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Добавить набор или убрать имеющийся?"))
			goto CmdInventory

		case "/compare":
			gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Введите номер набора для сравнения"))
			goto CmdCompare

		default:
			// gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Я съебался"))

		}
	}

CmdInventory:
	for update := range updates {
		switch update.Message.Text {

		case "/back":
			gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Вернулся"))
			goto MainLoop

		case "/addset":
			UpdateSetWindow(&update, updates, "добавления", "добавлено", "add", db)
			goto MainLoop

		case "/deleteset":
			UpdateSetWindow(&update, updates, "удаления", "удалено", "delete", db)
			goto MainLoop
		default:
			gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестный запрос"))
		}

	}

CmdCompare:
	for update := range updates {

		switch update.Message.Text {

		case "/back":
			gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Вернулся"))
			goto MainLoop

		default:
			tablename = fmt.Sprintf("InventoryTable_%d", update.Message.Chat.ID)
			Compare(update.Message.Text, "add", db, tablename, &update)
		}
	}

}
