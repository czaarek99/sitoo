package main

import (
	"database/sql"
	"log"
	"net/http"
	"sitoo/repositories"
	"sitoo/servers"
	"sitoo/services"
)

func main() {
	log.Println("Starting server")

	connection, err := sql.Open("mysql", "root:@/sitoo_test_assignment")

	if err != nil {
		log.Fatal("Could not connect to database")
	}

	repo := repositories.ProductRepositoryImpl{
		DB: connection,
	}

	service := services.ProductServiceImpl{
		Repo: repo,
	}

	server := servers.Server{
		Service: service,
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		server.HandleRequest(writer, request)
	})

	http.ListenAndServe(":8080", nil)
}
