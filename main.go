package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

var conn *sql.DB

func main() {
	var err error

	conn, err = sql.Open("postgres", os.Getenv("CONN_STR"))
	if err != nil {
		log.Fatal("connecting to db with err: ", err)
	}

	http.HandleFunc("/v1/words", wordsHandleFunc)

	log.Println("starting server")
	log.Fatal(http.ListenAndServe(":8090", nil))

}

func wordsHandleFunc(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		response []byte
	)

	defer func() {
		writeResponse(w, response, err)
	}()

	switch r.Method {
	case "POST":
		err = putWords(r.Body)
	}
}

func putWords(body io.ReadCloser) error {
	var (
		words []string
	)

	defer body.Close()

	raw, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, &words)
	if err != nil {
		return err
	}
	letters := getLettersMany(words)
	err = insertWords(words, letters)

	return err
}

func getLettersMany(words []string) (letters []string) {

	for i := range words {
		letters = append(letters, getLetters(words[i]))
	}
	return
}

func getLetters(word string) string {
	letters := strings.Split(word, "")
	sort.Strings(letters)
	return strings.Join(letters, "")
}

func writeResponse(w http.ResponseWriter, result []byte, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err == nil {
		w.WriteHeader(http.StatusOK)
		w.Write(result)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"%s"}`, err.Error())
	}
}
