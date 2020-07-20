package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"shortener.julianvos.nl/lib"

	_ "github.com/go-sql-driver/mysql"

	"github.com/joho/godotenv"

	"github.com/gorilla/mux"
)

func handleRequest() {
	fs := http.FileServer(http.Dir("./static"))
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", lib.Redirect).
		Methods("GET").
		Queries("destination", "")

	router.Handle("/", fs)

	router.HandleFunc("/generate", generateURL).
		Methods("GET")

	log.Fatal(http.ListenAndServe(":3002", router))
}

func generateURL(res http.ResponseWriter, req *http.Request) {
	var result = fmt.Sprintf("https://shortener.julianvos.nl/?destination=%v&origin=%v", url.QueryEscape(req.URL.Query()["destination"][0]), url.QueryEscape(req.URL.Query()["origin"][0]))
	res.Write([]byte(result))
	defer req.Body.Close()
}

func main() {
	godotenv.Load()
	handleRequest()
}
