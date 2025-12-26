package main

import (
	"log"

	"github.com/huynguyenanh2000/url-shorterner/internal/db"
	"github.com/huynguyenanh2000/url-shorterner/internal/env"
	"github.com/huynguyenanh2000/url-shorterner/internal/idgen"
	"github.com/huynguyenanh2000/url-shorterner/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "admin:adminpassword@tcp(localhost:3306)/url_shorterner?parseTime=true")
	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	store := store.NewStorage(conn)

	machineID := env.GetInt("MACHINE_ID", 1)

	snowflakeIDGenerator, err := idgen.NewSnowflakeClient(int64(machineID))
	if err != nil {
		log.Fatal(err)
	}

	db.SeedURLs(store, conn, snowflakeIDGenerator)
}
