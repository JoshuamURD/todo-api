package main

import (
	"joshuamURD/go-auth-api/pkgs/controllers"
	"joshuamURD/go-auth-api/pkgs/db"
	"joshuamURD/go-auth-api/pkgs/hash"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite" // Import with blank identifier to register the driver
)

func main() {
	// The database is initialised with the path to the database file
	db.Initialize(db.Config{Path: "test.db"})
	database := db.GetInstance()

	//Initialises a hasher with the default cost of bcrypt
	hasher := hash.NewBcryptHasher(bcrypt.DefaultCost)

	//Intialise the controllers with the hasher and the database
	//The controller is used to handle the requests and responses
	registerController := controllers.NewController(hasher, &database)

	//Initialises the mux and add the routes to it
	mux := http.NewServeMux()
	mux.HandleFunc("/register", registerController.Register)
	mux.HandleFunc("/login", registerController.Login)

	//Initialises the server with the mux and the port and the error log
	server := http.Server{
		Addr:     "127.0.0.1:8080",
		Handler:  mux,
		ErrorLog: log.New(os.Stderr, "ErrorLog: ", log.Lshortfile),
	}

	log.Printf("Server is running on %s", server.Addr)

	//Starts the server
	server.ListenAndServe()
}
