package main

import (
	"beta/db"
	"fmt"
)

func main() {
	fmt.Println("main 시작")
	db.DBInit()
	fmt.Println("db 생성완료")
}
