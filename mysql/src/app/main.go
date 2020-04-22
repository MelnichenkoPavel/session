package main

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	DbName string = "db1"
	TableName string = "collection"
	PoolLimit int = 10000
	LogPath string = "/var/logs/app.log"
)

var (
	logFileHandler, _ = os.OpenFile(LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	appLog = log.New(logFileHandler, "[app]", log.Lshortfile|log.LstdFlags)
)

//TODO. Метод сделан, что бы дождаться пока mysql стартанет в контейнере и настоит пользователей
func waitingForMysql(db *sql.DB) {
	errorPing := db.Ping()
	for errorPing != nil {
		fmt.Println("Waiting mysql: ", time.Now())
		time.Sleep(time.Second)
		errorPing = db.Ping()
	}
	fmt.Println("MySql done")
}

func initDB(dbName string) error {

	db, errorConnect := sql.Open("mysql", "root:example@tcp(mysql:3306)/")
	if errorConnect != nil {
		return errorConnect
	}
	defer db.Close()

	// хак
	waitingForMysql(db)

	_, errorCreate := db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName)
	if errorCreate != nil {
		return errorCreate
	}
	return nil
}

func initTable(name string, db *sql.DB) error {
	sqlQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%v` (`key` varchar(255) NOT NULL, `value` JSON NOT NULL, PRIMARY KEY (`key`)) ENGINE=InnoDB DEFAULT CHARSET=utf8", name)
	_, errorCreate := db.Exec(sqlQuery)
	return errorCreate
}

func main() {
	errorCreateDbLog := mysql.SetLogger(log.New(logFileHandler, "[db]", log.Lshortfile|log.LstdFlags))
	if errorCreateDbLog != nil {
		appLog.Panicln(errorCreateDbLog.Error())
		panic(errorCreateDbLog)
	}

	if errorInitDB := initDB(DbName); errorInitDB != nil {
		appLog.Panicln(errorInitDB.Error())
		panic(errorInitDB)
	}

	db, errorConnect := sql.Open("mysql", "root:example@tcp(mysql:3306)/" + DbName)
	if errorConnect != nil {
		appLog.Panicln(errorConnect.Error())
		panic(errorConnect)
	}
	defer db.Close()

	db.SetMaxOpenConns(PoolLimit)

	if errorInitTable := initTable(TableName, db); errorInitTable != nil {
		appLog.Panicln(errorConnect.Error())
		panic(errorInitTable)
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, "Start page")
		appLog.Println("Start page")
	})

	http.HandleFunc("/read", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			appLog.Println("Method not allowed")
			return
		}

		key := request.URL.Query().Get("key")
		if len(key) < 1 {
			writer.WriteHeader(http.StatusBadRequest)
			appLog.Println("Empty param `key`")
			return
		}

		var result string
		errQuery := db.QueryRow("SELECT `value` FROM " + TableName + " WHERE `key` = ?", key).Scan(&result)
		if errQuery != nil {
			writer.WriteHeader(http.StatusNotFound)
			appLog.Println(errQuery.Error() + "Not find record key = " + key)
			return
		}

		fmt.Fprintf(writer, result)
	})

	http.HandleFunc("/write", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			appLog.Println("Method not allowed")
			return
		}
		key := request.FormValue("key")
		value := request.FormValue("value")

		if len(key) < 1 || len(value) < 1 {
			writer.WriteHeader(http.StatusBadRequest)
			appLog.Println("Empty param `key` or `value`")
			return
		}

		sqlQuery := fmt.Sprintf("INSERT INTO `%v` (`key`, `value`) VALUES ('%v', '%v') ON DUPLICATE KEY UPDATE `value` = '%v'", TableName, key, value, value)
		_, errorInsert := db.Exec(sqlQuery)
		if errorInsert != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			appLog.Println(errorInsert.Error() + "Error db upsert")
			return
		}
	})

	http.HandleFunc("/randomWrite", func(writer http.ResponseWriter, request *http.Request) {

		value := strconv.Itoa(rand.Intn(10000))
		key := "autoKey" + value

		sqlQuery := fmt.Sprintf("INSERT INTO `%v` (`key`, `value`) VALUES ('%v', '%v') ON DUPLICATE KEY UPDATE `value` = '%v'", TableName, key, value, value)
		_, errorInsert := db.Exec(sqlQuery)
		if errorInsert != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			appLog.Println(errorInsert.Error())
			return
		}

		fmt.Fprintln(writer, "Upsert key [" + key + "]")
	})

	http.HandleFunc("/randomRead", func(writer http.ResponseWriter, request *http.Request) {

		key := "autoKey" + strconv.Itoa(rand.Intn(10000))

		var result string
		errQuery := db.QueryRow("SELECT `value` FROM " + TableName + " WHERE `key` = ?", key).Scan(&result)
		if errQuery != nil {
			appLog.Println(errQuery.Error() + "Not find record key = " + key)
			fmt.Fprintln(writer, "Empty key [" + key + "]")
			return
		}

		fmt.Fprintf(writer, result)
	})

	errListen := http.ListenAndServe(":8088", nil)
	if errListen != nil {
		appLog.Panicln("cannot listen: " + errListen.Error())
		panic("cannot listen: " + errListen.Error())
	}
}
