package main

import (
	"api/repositories"
	"api/servers"
	"api/services"
	"api/util"
	"database/sql"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	log.Println("Starting server")

	connection, err := sql.Open("mysql", "sitoo:test@database/sitoo_test_assignment")

	if err != nil {
		log.Fatal("Could not connect to database")
	}

	var requestId uint32

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
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
