package main

import (
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"math/rand"
	"net/http"
	"log"
	"os"
	"strconv"
)

const (
	DbName string = "db1"
	CollectionName string = "collection"
	PoolLimit int = 10000
	LogPath string = "/var/logs/app.log"
)

var (
	logFileHandler, _ = os.OpenFile(LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	appLog = log.New(logFileHandler, "[app]", log.Lshortfile|log.LstdFlags)
)

type DbData struct {
	Key string `bson:"key"`
	Value string `bson:"value"`
}

func initCollection(session *mgo.Session) error {
	collection := session.DB(DbName).C(CollectionName)
	return collection.EnsureIndex(mgo.Index{
		Key: []string{"key"},
		Unique: true,
	})
}

func main() {
	session, errorConnect := mgo.Dial("mongodb://root:example@mongo:27017/")
	if errorConnect != nil {
		appLog.Fatalln(errorConnect.Error())
		panic(errorConnect)
	}
	defer session.Close()

	mgo.SetLogger(log.New(logFileHandler, "[db]", log.LstdFlags|log.Lshortfile))

	session.SetPoolLimit(PoolLimit)

	//mgo.SetStats(true)
	//mgo.SetDebug(true)

	if errorInit := initCollection(session); errorInit != nil {
		appLog.Fatalln(errorInit.Error())
		panic(errorInit)
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, "Start Page");
		appLog.Println("Start Page")
	})

	http.HandleFunc("/read", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			fmt.Fprintln(writer, "Method Not Allowed")
			writer.WriteHeader(http.StatusMethodNotAllowed)
			appLog.Println("Method not allowed")
			return
		}

		key := request.URL.Query().Get("key")
		if len(key) < 1 {
			fmt.Fprintln(writer, "Empty param `key`");
			writer.WriteHeader(http.StatusBadRequest)
			appLog.Println("Empty param `key`")
			return
		}

		collection := session.DB(DbName).C(CollectionName)

		var record DbData
		errorFind := collection.Find(bson.M{"key": key}).Limit(1).One(&record)
		if errorFind != nil {
			fmt.Fprintln(writer, "Not find record to key = " + key)
			writer.WriteHeader(http.StatusNotFound)
			appLog.Println(errorFind.Error() + "| Not find record to key = " + key)
			return
		}

		fmt.Fprintln(writer, record.Value)
	})

	http.HandleFunc("/write", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			fmt.Fprintln(writer, "Method Not Allowed");
			writer.WriteHeader(http.StatusMethodNotAllowed)
			appLog.Println("Method Not Allowed")
			return
		}
		key := request.FormValue("key")
		value := request.FormValue("value")

		if len(key) < 1 || len(value) < 1 {
			fmt.Fprintln(writer, "Empty `key` or `value`")
			writer.WriteHeader(http.StatusBadRequest)
			appLog.Println("Empty `key` or `value`")
			return
		}

		collection := session.DB(DbName).C(CollectionName)

		_, errorUpsert := collection.Upsert(bson.M{"key": key}, &DbData{Key: key, Value: value})
		if errorUpsert != nil {
			fmt.Fprint(writer, "Error upsert to db")
			writer.WriteHeader(http.StatusInternalServerError)
			appLog.Println(errorUpsert.Error() + "| Error upsert to db")
			return
		}
	})


	http.HandleFunc("/randomRead", func(writer http.ResponseWriter, request *http.Request) {

		key := "autoKey" + strconv.Itoa(rand.Intn(10000))
		collection := session.DB(DbName).C(CollectionName)

		var record DbData
		errorFind := collection.Find(bson.M{"key": key}).Limit(1).One(&record)
		if errorFind != nil {
			appLog.Println(errorFind.Error() + "| Not find record to key = " + key)
			fmt.Fprintln(writer, "Empty record [" + key + "]")
			return
		}

		fmt.Fprintln(writer, record.Value)
	})

	http.HandleFunc("/randomWrite", func(writer http.ResponseWriter, request *http.Request) {

		value := strconv.Itoa(rand.Intn(10000))
		key := "autoKey" + value

		collection := session.DB(DbName).C(CollectionName)

		_, errorUpsert := collection.Upsert(bson.M{"key": key}, &DbData{Key: key, Value: value})
		if errorUpsert != nil {
			fmt.Fprintln(writer, "Error upsert to db")
			appLog.Println(errorUpsert.Error() + "| Error upsert to db")
			return
		}

		fmt.Fprintln(writer, "Upsert key [" + key + "]")
	})

	errListen := http.ListenAndServe(":8088", nil)
	if errListen != nil {
		appLog.Panicln("cannot listen: " + errListen.Error())
		panic("cannot listen: " + errListen.Error())
	}
}
