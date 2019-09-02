package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

const snippetValidCharacters = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-"
const snippetLength = 6

var redisClient *redis.Client

func StringWithCharset(length int, charset string) string {
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	parameters := mux.Vars(r)

	p, e := redisClient.Get(parameters["snippet"]).Bytes()

	if e != nil {
		http.NotFound(w, r)
		return
	}
	w.Write(p)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(5368709120) // 5MB
	file, _, err := r.FormFile("body")

	if err != nil {
		http.Error(w, "Must include payload, curl -F body=@foo.txt", 400)
		return
	}

	body, _ := ioutil.ReadAll(file)
	snippet := StringWithCharset(snippetLength, snippetValidCharacters)

	redisClient.Set(snippet, body, 0)
	w.WriteHeader(http.StatusCreated)
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	s := fmt.Sprintf("%v://%v/%v \n", scheme, r.Host, snippet)
	w.Write([]byte(s))
}

func main() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	r := mux.NewRouter()
	r.HandleFunc("/{snippet:[a-zA-Z0-9-]+$}", viewHandler).Methods("GET")
	r.HandleFunc("/", saveHandler).Methods("POST")

	http.ListenAndServe(":8080", r)
}
