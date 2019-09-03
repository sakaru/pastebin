package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const snippetValidCharacters = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-"
const snippetLength = 6

var redisClient *redis.Client
var baseURL string

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
		http.Error(w, "Must include payload; curl -F body=@foo.txt", 400)
		return
	}

	body, _ := ioutil.ReadAll(file)
	snippet := StringWithCharset(snippetLength, snippetValidCharacters)

	redisClient.Set(snippet, body, 0)
	w.WriteHeader(http.StatusCreated)
	s := ""
	if baseURL != "" {
		s = fmt.Sprintf("%v/%v\n", baseURL, snippet)
	} else {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		s = fmt.Sprintf("%v://%v/%v\n", scheme, r.Host, snippet)
	}
	w.Write([]byte(s))
}

func main() {
	baseURL = os.Getenv("BASE_URL")
	redisAddr, ok := os.LookupEnv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	if !ok {
		redisAddr = "localhost:6379"
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	r := mux.NewRouter()
	r.HandleFunc("/{snippet:[a-zA-Z0-9-]+$}", viewHandler).Methods("GET")
	r.HandleFunc("/", saveHandler).Methods("POST")

	listenOn, ok := os.LookupEnv("LISTEN_ON")
	if !ok {
		listenOn = ":8080"
	}
	http.ListenAndServe(listenOn, r)
}
