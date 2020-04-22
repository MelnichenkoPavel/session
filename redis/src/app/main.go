package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

const (
	PoolLimit int = 10000
	LogPath string = "/var/logs/app.log"
)

var (
	logFileHandler, _ = os.OpenFile(LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	appLog = log.New(logFileHandler, "[app]", log.Lshortfile|log.LstdFlags)
)

func main() {
	redis.SetLogger(log.New(logFileHandler, "[db]", log.Lshortfile|log.LstdFlags))

	redis_client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
		MaxRetries: 5,
		PoolSize: PoolLimit,
	})

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, "Start page")
		appLog.Println("Start page")
	})

	http.HandleFunc("/read", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			fmt.Fprintln(writer, "Method not allowed")
			writer.WriteHeader(http.StatusMethodNotAllowed)
			appLog.Println("Method not allowed")
			return
		}

		key := request.URL.Query().Get("key")
		if len(key) < 1 {
			fmt.Fprintln(writer, "Empty param `key`")
			writer.WriteHeader(http.StatusBadRequest)
			appLog.Println("Empty param `key`")
			return
		}

		cmd := redis_client.Get(key)
		data, err := cmd.Result()
		if err != nil {
			fmt.Fprintln(writer, "Not find record key = " + key)
			writer.WriteHeader(http.StatusNotFound)
			appLog.Fatalln(err.Error() + "Not find record key = " + key)
			return
		}

		fmt.Fprintln(writer, data)
	})

	http.HandleFunc("/write", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			fmt.Fprintln(writer, "Method not allowed")
			writer.WriteHeader(http.StatusMethodNotAllowed)
			appLog.Println("Method not allowed")
			return
		}
		key := request.FormValue("key")
		value := request.FormValue("value")

		if len(key) < 1 || len(value) < 1 {
			fmt.Fprintln(writer, "Empty param `key` or `value`")
			writer.WriteHeader(http.StatusBadRequest)
			appLog.Println("Empty param `key` or `value`")
			return
		}

		status := redis_client.Set(key, value, 0)
		if status.Err() != nil {
			fmt.Fprintln(writer, "Error db set")
			writer.WriteHeader(http.StatusInternalServerError)
			appLog.Println(status.Err().Error() + "Error db set")
			return
		}
	})

	http.HandleFunc("/randomRead", func(writer http.ResponseWriter, request *http.Request) {

		key := "autoKey" + strconv.Itoa(rand.Intn(10000))

		cmd := redis_client.Get(key)
		data, err := cmd.Result()
		if err != nil {
			fmt.Fprintln(writer, "Empty record key = " + key)
			return
		}

		fmt.Fprintln(writer, data)
	})

	http.HandleFunc("/randomWrite", func(writer http.ResponseWriter, request *http.Request) {

		value := strconv.Itoa(rand.Intn(10000))
		key := "autoKey" + value

		status := redis_client.Set(key, value, 0)
		if status.Err() != nil {
			fmt.Fprintln(writer, "Error db set")
			appLog.Println(status.Err().Error() + "Error db set")
			return
		}

		fmt.Fprintln(writer, "Upsert key [" + key + "]")
	})

	errListen := http.ListenAndServe(":8088", nil)
	if errListen != nil {
		appLog.Panicln(errListen.Error())
		panic("cannot listen: " + errListen.Error())
	}
}
