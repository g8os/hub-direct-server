package main

import (
    "log"
    "net/http"
    "time"
    "fmt"
    "bytes"

    "github.com/Jumpscale/go-raml/nbd-ardb-gateway/goraml"

    "github.com/gorilla/mux"
    "gopkg.in/validator.v2"

    "github.com/go-redis/redis"
    "github.com/satori/go.uuid"
)

func getArdb() (*redis.Client, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:7777",
        Password: "", // no password set
        DB:       0,  // use default DB
    })

    _, err := client.Ping().Result()

    return client, err
}

func locallog(event string, transaction string, username string) {
    var line bytes.Buffer
    t := time.Now()

    line.WriteString(fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d ", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()))
    line.WriteString(fmt.Sprintf("%s %s %s\n", event, transaction, username))
    fmt.Print(line.String())
}

func transactionSingle() string {
    uid := uuid.NewV4()
    return uid.String()
}

func transactionBegin(username string) string {
    tid := transactionSingle()

    locallog("transaction-begin", tid, username)

    return tid
}

func transactionEnd(reason string, transaction string, username string) {
    locallog("transaction-end:" + reason, transaction, username)
}

func keyExists(client *redis.Client, key string, rootkey string) bool {
    if rootkey != "" {
        existsBool, _ := client.HExists(rootkey, key).Result()
        return existsBool
    }

    exists, _ := client.Exists(key).Result()
    return (exists == 1)
}

func main() {
    // input validator
    validator.SetValidationFunc("multipleOf", goraml.MultipleOf)

    r := mux.NewRouter()

    // home page
    r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "index.html")
    })

    // apidocs
    r.PathPrefix("/apidocs/").Handler(http.StripPrefix("/apidocs/", http.FileServer(http.Dir("./apidocs/"))))

    ExistsInterfaceRoutes(r, ExistsAPI{})

    InsertInterfaceRoutes(r, InsertAPI{})

    log.Println("starting server")
    locallog("server-started", "initializing", "root")

    s := &http.Server{
        Addr:           ":5000",
        Handler:        r,
        ReadTimeout:    3600 * time.Second,
        WriteTimeout:   3600 * time.Second,
        IdleTimeout:    3600 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }

    s.ListenAndServe()
}
