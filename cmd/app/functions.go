package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

func IsCompare(update *tgbotapi.Update) bool {
	return update.Message != nil && update.Message.Text == "/Compare"
}

func Compare(setnum string, mode string, db *sql.DB, tablename string) {
	setTable := "set_table"
	compareTable := "compare_table"
	DeleteLegoTable(db, setTable)
	DeleteLegoTable(db, compareTable)
	CreateLegoTable(db, setTable)
	API_Connect(setnum, mode, db, setTable)
	query := fmt.Sprintf(
		`	CREATE TABLE IF NOT EXISTS %s(
		compare_part_num character varying(50) COLLATE pg_catalog."default" NOT NULL,
		compare_part_name character varying(2000) COLLATE pg_catalog."default" NOT NULL,
		compare_color_id smallint NOT NULL,
		compare_color_name character varying(50) COLLATE pg_catalog."default" NOT NULL
	)
	TABLESPACE pg_default;

	INSERT INTO %s(compare_part_num, compare_part_name, compare_color_id, compare_color_name)
	(
		SELECT part_num, part_name, color_id, color_name 
		FROM %s 
		intersect 
		SELECT part_num, part_name, color_id, color_name 
		FROM %s
	);
	INSERT INTO %s(part_num, part_name, color_id, color_name, quantity)
	(
		SELECT compare_part_num, compare_part_name, compare_color_id, compare_color_name, -(quantity) 
		FROM %s JOIN %s 
		ON compare_part_num = part_num AND compare_part_name = part_name 
		AND compare_color_id = color_id AND compare_color_name = color_name
	);`, compareTable, compareTable, tablename, setTable, setTable, compareTable, tablename)
	_, err := db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}
	PartMerger(db, setTable)
}

func API_Connect(setnum string, mode string, db *sql.DB, tablename string) {
	url := fmt.Sprintf("https://rebrickable.com/api/v3/lego/sets/%s/parts/", setnum)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil) // middleware

	// брать значения из переменной окружения

	req.Header.Set("Authorization", "key 67062c7b14264aedb5c8e9966c83df02")
	req.Header.Set("Accept", "application/json")
	data, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	bodyBytes, err := io.ReadAll(data.Body)

	var set Sets
	json.Unmarshal(bodyBytes, &set)
	for _, count := range set.Results {
		fmt.Printf("деталь: %s, цвет: %s ID цвета: %d, кол-во %d\n", count.Part.PartNum, count.Color.Name, count.Color.ID, count.Quantity)
		UpadteInventory(db, count.Part.PartNum, count.Part.Name, count.Color.ID, count.Color.Name, count.Quantity, tablename, mode)

	}
	if err != nil {
		fmt.Println(err)
	}
}

func IsBack(update *tgbotapi.Update) bool {
	return update.Message != nil && update.Message.Text == "/back"
}

func Inventory(update *tgbotapi.Update) bool {
	return update.Message != nil && update.Message.Text == "/Inventory"
}

func AddSetInventory(update *tgbotapi.Update) bool {
	return update.Message != nil && update.Message.Text == "/addset"
}

func DeleteSetInventory(update *tgbotapi.Update) bool {
	return update.Message != nil && update.Message.Text == "/deleteset"
}

func PartMerger(db *sql.DB, tablename string) {
	query := fmt.Sprintf(`DROP TABLE IF EXISTS public.part_switch;
	CREATE TABLE IF NOT EXISTS public.part_switch 
	(
    part_num character varying(50) COLLATE pg_catalog."default" NOT NULL,
    part_name character varying(2000) COLLATE pg_catalog."default" NOT NULL,
    color_id smallint NOT NULL,
    color_name character varying(50) COLLATE pg_catalog."default" NOT NULL,
    quantity smallint NOT NULL
	)
	TABLESPACE pg_default;
	INSERT INTO part_switch(part_num, part_name, color_id, color_name, quantity)
	(
	SELECT 
	  part_num,
	  part_name,
	  color_id,
	  color_name,
	  SUM (quantity) AS quantity
	FROM 
	  %s	 
	GROUP BY 
	  part_num,
	  part_name,
	  color_id,
	  color_name 
	ORDER BY 
  	  part_num);
	truncate %s;
	INSERT INTO %s(part_num, part_name, color_id, color_name, quantity)(select * from part_switch);
	DROP TABLE IF EXISTS public.part_switch;`, tablename, tablename, tablename)

	_, err := db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}
}

func CreateLegoTable(db *sql.DB, tablename string) {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(
		part_num character varying(50) COLLATE pg_catalog."default" NOT NULL,
		part_name character varying(2000) COLLATE pg_catalog."default" NOT NULL,
		color_id smallint NOT NULL,
		color_name character varying(50) COLLATE pg_catalog."default" NOT NULL,
		quantity smallint NOT NULL
	)`, tablename)

	_, err := db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}
}

func UpadteInventory(db *sql.DB, part_num string, part_name string, color_id int, color_name string, quantity int, tablename string, mode string) {

	query := fmt.Sprintf(`INSERT INTO %s(part_num, part_name, color_id, color_name, quantity)
		VALUES ($1, $2, $3, $4, $5) RETURNING part_num`, tablename)

	if mode == "delete" {
		quantity = -quantity
	}
	_, err := db.Exec(query, part_num, part_name, color_id, color_name, quantity)

	if err != nil {
		log.Fatal(err)
	}
}

func DeleteLegoTable(db *sql.DB, tablename string) {
	query := fmt.Sprintf(`DROP TABLE IF EXISTS public.%s;`, tablename)
	_, err := db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}
}

func AddOrDeleteSet(update *tgbotapi.Update, updates tgbotapi.UpdatesChannel, move1 string, move2 string, mode string, db *sql.DB) {
	msg := fmt.Sprintf("Введите номер набора для %s", move1)
	gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
	for update := range updates {
		switch update.Message.Text {
		case "/back":
			return
		default:
			fmt.Print(update.Message.Text)
			API_Connect(update.Message.Text, mode, db, tablename)
			PartMerger(db, tablename)
			msg = fmt.Sprintf("Введите номер набора для %s", move2)
			gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
			return
		}
	}
}
