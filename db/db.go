package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const dblogfilepath = "./log/dblog.log"
const dbfilepath = "./data/user.db"

var DB *sql.DB

func DBInit() {

	fmt.Println("DBInit 시작")
	os.MkdirAll("./log", 0755)
	os.MkdirAll("./data", 0755)
	logFile, err := os.OpenFile(dblogfilepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("로그 파일 열기 실패 : %v\n", err)
	}

	log.SetOutput(logFile)
	log.Println("서버 초기화 시작")

	DB, err = sql.Open("sqlite3", dbfilepath)
	if err != nil {
		log.Fatalf("DB 연결 실패: %v\n", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatalf("DB 연결 실패: %v\n", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS users(
		ID TEXT PRIMARY KEY,
		Username TEXT NOT NULL,
		Role TEXT NOT NULL,
		PasswordHash TEXT NOT NULL,
		CreatedAt TEXT NOT NULL
	);`

	_, err = DB.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("테이블 생성 실패: %v\n", err)
	}

	log.Println("DB 초기 설정 완료")
}
