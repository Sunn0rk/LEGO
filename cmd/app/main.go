package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-telegram/bot"
	_ "github.com/lib/pq"
)

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

func main() {

	b, err := bot.New("6821580874:AAFidms6Y2RyW6TQUFBYmCgY8rTJArjRuFo")

	// table_name = "parttest"
	// database_connect()
	server := "localhost"
	database := "postgres"
	port := "5432"
	username := "postgres"
	password := "9201"

	// connStr := "postgres://postgres:9201@localhost:5432/pqgotest?sslmode=disable"
	connStr := fmt.Sprintf("%s://%s:%s@%s:%s/pqgotest?sslmode=disable", database, username, password, server, port)

	db, err := sql.Open(database, connStr) //?
	fmt.Println(db)
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	CreateLegoTable(db)

	api_connect()

	PartMerger(db)

	// fmt.Println(url)
}

// ? как возвращать значения
func database_connect() {
	// server := "localhost"
	// database := "postgres"
	// port := 5432
	// username := "postgres"
	// password := "9201"

	connStr := "postgres://postgres:9201@localhost:5432/pqgotest?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	fmt.Println(db)
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func api_connect() {
	setnum := "60115-1"
	url := fmt.Sprintf("https://rebrickable.com/api/v3/lego/sets/%s/parts/", setnum)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil) // middleware

	// брать значения из переменной окружения

	req.Header.Set("Authorization", "key 67062c7b14264aedb5c8e9966c83df02")
	req.Header.Set("Accept", "application/json")
	data, err := client.Do(req)
	bodyBytes, err := io.ReadAll(data.Body)

	var set Sets
	json.Unmarshal(bodyBytes, &set)
	for _, count := range set.Results {
		fmt.Printf("деталь: %s, цвет: %s ID цвета: %d, кол-во %d\n", count.Part.PartNum, count.Color.Name, count.Color.ID, count.Quantity)
		// DeleteSet(db, count.Part.PartNum, count.Part.Name, count.Color.ID, count.Color.Name, count.Quantity)
		// InsertSet(db, count.Part.PartNum, count.Part.Name, count.Color.ID, count.Color.Name, count.Quantity)
	}
	fmt.Println(err)
}

func PartMerger(db *sql.DB) {
	// ? имя таблицы
	query := `DROP TABLE IF EXISTS public.part_switch;
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
	  parttest	 
	GROUP BY 
	  part_num,
	  part_name,
	  color_id,
	  color_name 
	ORDER BY 
  	  part_num);
  
truncate parttest;
INSERT INTO parttest(part_num, part_name, color_id, color_name, quantity)(select * from part_switch);
DROP TABLE IF EXISTS public.part_switch;`

	_, err := db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}
}

func CreateLegoTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS partTest (
		part_num character varying(50) COLLATE pg_catalog."default" NOT NULL,
		part_name character varying(2000) COLLATE pg_catalog."default" NOT NULL,
		color_id smallint NOT NULL,
		color_name character varying(50) COLLATE pg_catalog."default" NOT NULL,
		quantity smallint NOT NULL
	)`

	_, err := db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}
}

func InsertSet(db *sql.DB, part_num string, part_name string, color_id int, color_name string, quantity int) {

	query := `INSERT INTO partTest(part_num, part_name, color_id, color_name, quantity)
		VALUES ($1, $2, $3, $4, $5) RETURNING part_num`

	_, err := db.Exec(query, part_num, part_name, color_id, color_name, quantity)

	if err != nil {
		log.Fatal(err)
	}
}

func DeleteSet(db *sql.DB, part_num string, part_name string, color_id int, color_name string, quantity int) {
	query := `INSERT INTO partTest(part_num, part_name, color_id, color_name, quantity)
		VALUES ($1, $2, $3, $4, $5) RETURNING part_num`

	quantity = -quantity
	_, err := db.Exec(query, part_num, part_name, color_id, color_name, quantity)

	if err != nil {
		log.Fatal(err)
	}
}
