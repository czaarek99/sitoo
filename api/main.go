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

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	log.Println("Starting server")

	connectionString := "root@/sitoo_test_assignment"

	if os.Args[2] == "docker" {
		connectionString = "sitoo:test@database/sitoo_test_assignment"
	}

	connection, err := sql.Open("mysql", connectionString)

	if err != nil {
		log.Fatal("Could not connect to database")
	}

	err = connection.Ping()

	if err != nil {
		log.Print(err.Error())
		log.Fatal("Could not ping database")
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
