package main

import (
	"api/repositories"
	"api/servers"
	"api/services"
	"api/util"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	log.Println("Starting server")

	var connectionString string

	if len(os.Args) > 1 && os.Args[1] == "docker" {
		log.Printf("Using docker")
		connectionString = "sitoo:test@tcp(database)/sitoo_test_assignment"
	} else {
		log.Printf("Running standalone")
		connectionString = "root@/sitoo_test_assignment"
	}

	connection, err := sql.Open("mysql", connectionString)

	if err != nil {
		log.Fatal("Could not open database connection")
	}

	retryCount := 0

	for {
		err := connection.Ping()

		if retryCount > 10 {
			log.Fatal("Unable to ping database 10 times, giving up")
		}

		if err != nil {
			log.Printf("Could not ping database with error %s", err.Error())
			retryCount++

			time.Sleep(2 * time.Second)
		} else {
			log.Println("Database ping successful!")
			break
		}
	}

	var requestId uint32

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		requestId++

		repo := repositories.ProductRepositoryImpl{
			DB: connection,
		}

		service := services.ProductServiceImpl{
			Repo: repo,
			Metadata: util.Metadata{
				RequestID: requestId,
			},
		}

		server := servers.Server{
			Service: service,
		}

		server.HandleRequest(writer, request)
	})

	http.ListenAndServe(":8080", nil)
}
