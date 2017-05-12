package main

import (
    "net/http"
    "fmt"
    "io/ioutil"
    "encoding/base64"
)

// InsertAPI is API implementation of /insert root endpoint
type InsertAPI struct {
}

// Put is the handler for PUT /insert
func (api InsertAPI) Put(w http.ResponseWriter, r *http.Request) {
	// uncomment below line to add header
	// w.Header().Set("key","value")

    usernames, usernamefound := r.Header["X-Iyo-Username"]

    if ! usernamefound {
        fmt.Println("Access denied")
        w.WriteHeader(401)
        return
    }

    username := usernames[0]

    rootkey := r.FormValue("rootkey")
    if rootkey != "" {
        fmt.Println("Using HSET:", rootkey)
    }

    err := r.ParseMultipartForm(8192)
    if err != nil {
        fmt.Println(err)
		w.WriteHeader(500)
		return
	}

    // connect ardb
    client, err := getArdb()
    if err != nil {
        fmt.Println(err)
		w.WriteHeader(500)
		return
	}

    m := r.MultipartForm
    files := m.File["files[]"]

    fmt.Printf("Inserting %d entries\n", len(files))
    transaction := transactionBegin("username")

    for i, _ := range files {
        locallog("key-insert:" + files[i].Filename, transaction, username)

        file, err := files[i].Open()
        defer file.Close()

        if err != nil {
            w.WriteHeader(400)
            transactionEnd("file-read-error", transaction, username)
            fmt.Println(err)
            return
        }

        decoded, err := base64.StdEncoding.DecodeString(files[i].Filename)
        if err != nil {
            fmt.Println(err)
            transactionEnd("base64-error", transaction, username)
            w.WriteHeader(500)
            return
        }

        // fmt.Println("Inserting:", files[i].Filename)

        if keyExists(client, string(decoded), rootkey) == true {
            w.WriteHeader(401)
            transactionEnd("key-exists-denied", transaction, username)
            fmt.Println("Key exists, overwrite denied:", files[i].Filename)
            return
        }

        buffer, err := ioutil.ReadAll(file)
        if err != nil {
            transactionEnd("read-error", transaction, username)
            w.WriteHeader(500)
            fmt.Println(err)
            return
        }

        if rootkey != "" {
            client.HSet(rootkey, string(decoded), buffer)

        } else {
            client.Set(string(decoded), buffer, 0)
        }
    }

    transactionEnd("success", transaction, username)
    client.Close()
}
