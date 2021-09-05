package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/lonord/sse"
)

func main() {
	addrFlag := flag.String("addr", ":8080", "addr to listen")
	flag.Parse()

	http.HandleFunc("/data", handleData)
	http.HandleFunc("/watch", handleWatch)

	log.Println("start server on", *addrFlag)
	log.Fatal(http.ListenAndServe(*addrFlag, nil))
}

var (
	dataLock   sync.RWMutex
	sseService *sse.Service

	dataBytes []byte
	dataType  string
)

func init() {
	sseService = sse.NewService()
	dataBytes = make([]byte, 0)
	dataType = "text/plain"
}

func handleData(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		dataLock.RLock()
		defer dataLock.RUnlock()

		w.Header().Set("Content-Type", dataType)
		w.WriteHeader(http.StatusOK)
		io.Copy(w, bytes.NewReader(dataBytes))
		log.Printf("get data %s (%d bytes)\n", dataType, len(dataBytes))

	} else if r.Method == "POST" {
		dataLock.Lock()
		defer dataLock.Unlock()

		w.Header().Set("Content-Type", "text/plain")

		// content size limit 50MiB
		if r.ContentLength > 50*1024*1024 {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			fmt.Fprintln(w, "content too large (max 50MiB)")
			log.Printf("content too large (put %d bytes)\n", r.ContentLength)
			return
		}

		dataType = r.Header.Get("Content-Type")
		var b bytes.Buffer
		io.Copy(&b, r.Body)
		dataBytes = b.Bytes()
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
		log.Printf("put data %s (%d bytes)\n", dataType, b.Len())
		sseService.Broadcast(sse.Event{Data: "NewData"})

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleWatch(w http.ResponseWriter, r *http.Request) {
	id := sse.GenerateClientID()
	ch, err := sseService.HandleClient(id, w)
	if err != nil {
		handleServerError(err, w)
		return
	}
	log.Println("watch from new client", id)
	<-ch
	log.Println("watch closed for client", id)
}

func handleServerError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Println("server error:", err)
	fmt.Fprintf(w, "server error: %s", err)
}
