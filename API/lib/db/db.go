package db

import (
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PCCSuite/PCCSamba/SambaAPI/lib"
	_ "github.com/lib/pq"
)

var db *sql.DB

const table = "samba_users"

func InitDB() {
	var err error
	options := []string{
		"host=" + os.Getenv("PCC_SAMBAAPI_DB_ADDR"),
		"dbname=" + os.Getenv("PCC_SAMBAAPI_DB_NAME"),
		"user=" + os.Getenv("PCC_SAMBAAPI_DB_USER"),
		"password=" + os.Getenv("PCC_SAMBAAPI_DB_PASSWORD"),
		"sslmode=disable",
	}
	db, err = sql.Open("postgres", strings.Join(options, " "))
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		log.Printf("DB Ping try: %d", i+1)
		err = db.Ping()
		if err != nil {
			log.Printf("DB Ping failed: %v", err)
			time.Sleep(1 * time.Second)
		} else {
			log.Println("DB Ping success")
			prepereDB()
			return
		}
	}
	log.Fatal("Failed to initialize DB")
}

func CloseDB() {
	db.Close()
}

var getStmt *sql.Stmt
var setStmt *sql.Stmt
var addStmt *sql.Stmt

func prepereDB() {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS " + table + " ( " +
		"id TEXT NOT NULL PRIMARY KEY, " +
		"mode INTEGER NOT NULL, " +
		"data TEXT)")
	if err != nil {
		log.Panic("Failed to CREATE table: ", err)
	}
	getStmt, err = db.Prepare("SELECT * FROM " + table + " WHERE id = $1")
	if err != nil {
		log.Panic("Failed to prepere getStmt: ", err)
	}
	setStmt, err = db.Prepare("UPDATE " + table + " SET mode = $1, data = $2 WHERE id = $3")
	if err != nil {
		log.Panic("Failed to prepere setStmt: ", err)
	}
	addStmt, err = db.Prepare("INSERT INTO " + table + " (id,mode,data) VALUES ($1,$2,$3)")
	if err != nil {
		log.Panic("Failed to prepere addStmt: ", err)
	}
}

type UserData struct {
	ID   string
	Mode lib.PasswordMode
	Data string
}

func GetData(id string) (*UserData, error) {
	data := UserData{}
	row := getStmt.QueryRow(id)
	err := row.Scan(&data.ID, &data.Mode, &data.Data)
	return &data, err
}

func SetData(data *UserData) error {
	_, err := setStmt.Exec(data.Mode, data.Data, data.ID)
	return err
}

func AddUser(data *UserData) error {
	_, err := addStmt.Exec(data.ID, data.Mode, data.Data)
	return err
}
