package main

import (
	"fmt"
	"lightcloud/db"
)

func main() {
	db.DBInit()
	fmt.Print("fin")
}
