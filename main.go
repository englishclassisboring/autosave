package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/nu7hatch/gouuid"
	"log"
	"net/http"
	"os"
	"time"
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
		Addr:     os.Getenv("ADDR"),
		Password: os.Getenv("PASS"),
		DB:       0, // use default DB
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
		start := time.Now()
		errSet := rdb.Set(ctx, uuid.String(), string(p), 1*time.Minute).Err()
		t := time.Now()
		elapsed := t.Sub(start)
		val2, err := rdb.Get(ctx, uuid.String()).Result()

		var mes = "added " + val2 + " at " + uuid.String() + " - took " + elapsed.String() + "\n http://localhost:1234/id?uuid=" + uuid.String()
		fmt.Println(mes)

		if errSet != nil {
			panic(err)
		}
		if err := conn.WriteMessage(messageType, []byte(mes)); err != nil {
			log.Println(err)
			return
		}

	}
}

func idHandler(w http.ResponseWriter, r *http.Request) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("ADDR"),
		Password: os.Getenv("PASS"),
		DB:       0, // use default DB
	})

	keys, ok := r.URL.Query()["uuid"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return
	}
	key := keys[0]
	fmt.Println("getting uuid", key)
	val2, errget := rdb.Get(ctx, key).Result()
	if errget != nil {
		panic(errget)
	}
	fmt.Println("get val: ", val2)
	fmt.Fprintf(w, val2)
}
func setupRoutes() {
	http.HandleFunc("/ws", wsEndpoint)
	http.HandleFunc("/id", idHandler)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("Online")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":1234", nil))
}
