package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const dblogfilepath = "./log/dblog.log"
const dbfilepath = "./data/lightcloud.db"

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


	createTableQueries := []string{
	`
	CREATE TABLE IF NOT EXISTS users(
		ID TEXT PRIMARY KEY,
		Username TEXT NOT NULL,
		Role TEXT NOT NULL,
		PasswordHash TEXT NOT NULL,
		CreatedAt TEXT NOT NULL
	)`,
	`
	CREATE TABLE IF NOT EXISTS files(
	ID TEXT PRIMARY KEY,
	OwnerID TEXT NOT NULL,
	OriginalName TEXT NOT NULL,
	StoredName TEXT NOT NULL,
	Size INTEGER NOT NULL,
	MimeType TEXT NOT NULL,
	CreatedAt TEXT NOT NULL
	)`,
	`
	CREATE TABLE IF NOT EXISTS sessions(
	Token TEXT PRIMARY KEY,
	UserID TEXT NOT NULL,
	ExpiresAt TEXT NOT NULL,
	CreatedAt TEXT NOT NULL
	)`,
	`
	CREATE TABLE IF NOT EXISTS share_links(
	Token TEXT PRIMARY KEY,
	FileID TEXT NOT NULL,
	CreatedBy TEXT NOT NULL,
	CreatedAt TEXT NOT NULL,
	ExpiresAt TEXT NOT NULL,
	PasswordHash TEXT
	)`,
	}
	
	for _,q := range createTableQueries{
		_, err = DB.Exec(q)
		if err != nil {
		log.Fatalf("테이블 생성 실패 [%s]: %v\n", q[:30], err)
	}
	}

	
	

	log.Println("DB 초기 설정 완료")
}
