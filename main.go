package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/nu7hatch/gouuid"
	"log"
	"net/http"
)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var ctx = context.Background()

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	u, err := uuid.NewV4()
	log.Println("Client Connected")

	reader(ws, rdb, u)
}

func reader(conn *websocket.Conn, rdb *redis.Client, uuid *uuid.UUID) {
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(string(p))
		errSet := rdb.Set(ctx, uuid.String(), string(p), 60).Err()
		if errSet != nil {
			panic(err)
		}
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}

	}
}

func setupRoutes() {
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	fmt.Println("Online")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":1234", nil))
}
