package main

import (
  "net/http"
  "io/ioutil"
  "regexp"
  "math/rand"
  "fmt"
  "time"
  "github.com/go-redis/redis"
)

const validCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-"
const length = 6

func createSnippetName() string {
  r := rand.New(rand.NewSource(time.Now().UnixNano()))
  b := make([]byte, length)
  for i := range b {
    b[i] = validCharacters[r.Intn(len(validCharacters))]
  }
  return string(b)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Print("GET ")
  var validPath = regexp.MustCompile("^/v/([a-zA-Z0-9-]+)$")
  m := validPath.FindStringSubmatch(r.URL.Path)
  if m == nil {
    http.NotFound(w, r)
    fmt.Println("[invalid-url]")
    return
  }
  snippet := m[1]


  p, e := client.Get(snippet).Bytes()

  if e != nil {
    http.NotFound(w, r)
    fmt.Println("[not-found]")
    return
  }
  fmt.Println(snippet, "[OK]")
  w.Write(p)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
  r.ParseMultipartForm(5368709120) // 5MB
  file, _, err := r.FormFile("body")


  if err != nil {
    fmt.Println("[invalid]")
    http.Error(w, "Must include file upload", 400)
    return
  }

  body, _ := ioutil.ReadAll(file)

  snippet := createSnippetName()

  client.Set(snippet, body, 0)
  fmt.Println("SET ", snippet)
  w.WriteHeader(http.StatusCreated)
  s := fmt.Sprintf("See http://%v/v/%v \n", r.Host, snippet)
  w.Write([]byte(s))
}

var client *redis.Client

func main() {
  client = redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "", // no password set
    DB:       0,  // use default DB
  })

  http.HandleFunc("/v/", viewHandler)
  http.HandleFunc("/s",  saveHandler)

  http.ListenAndServe(":8080", nil)
}