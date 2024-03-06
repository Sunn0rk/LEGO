package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

func Compare(setnum string, mode string, db *sql.DB, tablename string, update *tgbotapi.Update) {
	setTable := fmt.Sprintf("%s_set", tablename)
	compareTable := fmt.Sprintf("%s_compare", tablename)
	DeleteLegoTable(db, setTable)
	DeleteLegoTable(db, compareTable)
	CreateLegoTable(db, setTable)
	check := API_Connect(setnum, mode, db, setTable)
	switch check {
	case 0:
		gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный номер набора"))
		return
	default:
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
			log.Panic(err)
		}
		PartMerger(db, setTable)
	}

}

func API_Connect(setnum string, mode string, db *sql.DB, tablename string) (check int) {
	url := fmt.Sprintf("https://rebrickable.com/api/v3/lego/sets/%s/parts/", setnum)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("Authorization", "key 67062c7b14264aedb5c8e9966c83df02")
	req.Header.Set("Accept", "application/json")
	data, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	bodyBytes, err := io.ReadAll(data.Body)

	var set Sets
	json.Unmarshal(bodyBytes, &set)
	fmt.Println(set.Results)
	check = len(set.Results)
	if check == 0 {

	} else {
		for _, count := range set.Results {
			// fmt.Printf("деталь: %s, цвет: %s ID цвета: %d, кол-во %d\n", count.Part.PartNum, count.Color.Name, count.Color.ID, count.Quantity)
			UpadteInventory(db, count.Part.PartNum, count.Part.Name, count.Color.ID, count.Color.Name, count.Quantity, tablename, mode)
		}
	}

	if err != nil {
		fmt.Println(err)
	}
	return check
}

func PartMerger(db *sql.DB, tablename string) {
	query := fmt.Sprintf(`DROP TABLE IF EXISTS public.%s_switch;
	CREATE TABLE IF NOT EXISTS public.%s_switch 
	(
    part_num character varying(50) COLLATE pg_catalog."default" NOT NULL,
    part_name character varying(2000) COLLATE pg_catalog."default" NOT NULL,
    color_id smallint NOT NULL,
    color_name character varying(50) COLLATE pg_catalog."default" NOT NULL,
    quantity smallint NOT NULL
	)
	TABLESPACE pg_default;
	INSERT INTO %s_switch(part_num, part_name, color_id, color_name, quantity)
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
	INSERT INTO %s(part_num, part_name, color_id, color_name, quantity)(select * from %s_switch);
	DROP TABLE IF EXISTS public.%s_switch;`, tablename, tablename, tablename, tablename, tablename, tablename, tablename, tablename)

	_, err := db.Exec(query)

	if err != nil {
		log.Panic(err)
	}
}

func CreateLegoTable(db *sql.DB, tablename string) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(
		part_num character varying(50) COLLATE pg_catalog."default" NOT NULL,
		part_name character varying(2000) COLLATE pg_catalog."default",
		color_id smallint NOT NULL,
		color_name character varying(50) COLLATE pg_catalog."default",
		quantity smallint NOT NULL
	)`, tablename)

	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func UpadteInventory(db *sql.DB, part_num string, part_name string, color_id int, color_name string, quantity int, tablename string, mode string) {

	query := fmt.Sprintf(`INSERT INTO %s(part_num, part_name, color_id, color_name, quantity)
		VALUES ($1, $2, $3, $4, $5) RETURNING part_num`, tablename)

	if mode == "delete" {
		quantity = -quantity
	}
	_, err := db.Exec(query, part_num, part_name, color_id, color_name, quantity)

	if err != nil {
		log.Panic(err)
	}
}

func DeleteLegoTable(db *sql.DB, tablename string) {
	query := fmt.Sprintf(`DROP TABLE IF EXISTS public.%s;`, tablename)
	_, err := db.Exec(query)

	if err != nil {
		log.Panic(err)
	}
}

func UpdateSetWindow(update *tgbotapi.Update, updates tgbotapi.UpdatesChannel, move1 string, move2 string, mode string, db *sql.DB) {

	msg := fmt.Sprintf("Введите номер набора для %s", move1)
	gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))

	for update := range updates {
		switch update.Message.Text {

		case "/back":
			return

		default:
			UpdateSetCommand(&update, updates, move1, move2, mode, db)
			return
		}
	}
}

func UpdateSetCommand(update *tgbotapi.Update, updates tgbotapi.UpdatesChannel, move1 string, move2 string, mode string, db *sql.DB) {
	if mode == "delete" && !BeforeDeleteCheck(update) {
		gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Такого набора в инвентаре нет"))
		return
	} else {
		tablename = fmt.Sprintf("InventoryTable_%d", update.Message.Chat.ID)
		check := API_Connect(update.Message.Text, mode, db, tablename)
		switch check {
		case 0:
			gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный номер набора"))
			return
		default:
			PartMerger(db, tablename)
			tablename = fmt.Sprintf("SetHistoryTable_%d", update.Message.Chat.ID)
			UpdateSetWindowHistory(db, tablename, update.Message.Text, 1, mode)
			SetHistoryMerger(db, tablename)
			gBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, move2))
			return
		}

	}
}

func BeforeDeleteCheck(update *tgbotapi.Update) bool {
	tablename = fmt.Sprintf("SetHistoryTable_%d", update.Message.Chat.ID)
	query := fmt.Sprintf("SELECT quantity FROM %s WHERE setnum = '%s'", tablename, update.Message.Text)
	var quantity int
	err = db.QueryRow(query).Scan(&quantity)
	if quantity > 0 {
		return true
	} else {
		return false
	}
}

func DatabaseConnect() (*sql.DB, error) {
	fmt.Println(database, username, password, server, port)
	connStr := fmt.Sprintf("%s://%s:%s@%s:%s/Lego?sslmode=disable", database, username, password, server, port)
	db, err = sql.Open(driverName, connStr)
	fmt.Println(db)

	if err != nil {
		return db, err
	}
	return db, nil
}

func TGBotConnect() error {
	_ = os.Setenv(TOKEN_NAME_IN_OS, "6842123718:AAGAhkDOdqUMTLuCzo4CkzPxXzpNil4VMj8")
	gToken = os.Getenv(TOKEN_NAME_IN_OS)
	if err != nil {
		return err
	}
	return nil
}

func CreateSetHistory(db *sql.DB, tablename string) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(
		setnum character varying(50) COLLATE pg_catalog."default" NOT NULL,
		quantity smallint NOT NULL
	)`, tablename)

	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func UpdateSetWindowHistory(db *sql.DB, tablename string, setnum string, quantity int, mode string) {
	query := fmt.Sprintf(`INSERT INTO %s(setnum, quantity)
		VALUES ($1, $2) RETURNING setnum`, tablename)

	if mode == "delete" {
		quantity = -quantity
	}
	_, err := db.Exec(query, setnum, quantity)

	if err != nil {
		log.Panic(err)
	}
}

func SetHistoryMerger(db *sql.DB, tablename string) {
	query := fmt.Sprintf(`DROP TABLE IF EXISTS public.%s_switch;
	CREATE TABLE IF NOT EXISTS public.%s_switch 
	(
    setnum character varying(50) COLLATE pg_catalog."default" NOT NULL,
    quantity smallint NOT NULL
	)
	TABLESPACE pg_default;
	INSERT INTO %s_switch(setnum, quantity)
	(
	SELECT 
	  setnum,
	  SUM (quantity) AS quantity
	FROM 
	  %s	 
	GROUP BY 
	  setnum 
	ORDER BY 
  	  setnum);
	truncate %s;
	INSERT INTO %s(setnum, quantity)(select * from %s_switch);
	DROP TABLE IF EXISTS public.%s_switch;`, tablename, tablename, tablename, tablename, tablename, tablename, tablename, tablename)

	_, err := db.Exec(query)

	if err != nil {
		log.Panic(err)
	}
}
