/*
The simple http server implements the table of counters.
to create counters :- set?name=N&val=V
to get counter :- get?name=N
to increment counter :- inc?name=N

*/
package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type CoutnerStore struct {
	sync.Mutex
	counters map[string]int
}

func (cs *CoutnerStore) get(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET request %v", r)
	name := r.URL.Query().Get("name")
	if val, ok := cs.counters[name]; ok {
		fmt.Fprintf(w, "%s,%d", name, val)
	} else {
		fmt.Fprint(w, "Page NOT FOUND")
	}
}

func (cs *CoutnerStore) set(w http.ResponseWriter, r *http.Request) {
	// Lock-UnLock is important to avoid data racing .
	cs.Lock()
	defer cs.Unlock()
	name := r.URL.Query().Get("name")
	val := r.URL.Query().Get("val")
	intval, err := strconv.Atoi(val)
	if err != nil {
		fmt.Fprintf(w, "Error %s", err)
	} else {
		cs.counters[name] = intval
		fmt.Fprintf(w, "OK!\n")
	}
}

func (cs *CoutnerStore) inc(w http.ResponseWriter, r *http.Request) {
	cs.Lock()
	defer cs.Unlock()
	name := r.URL.Query().Get("name")
	if _, ok := cs.counters[name]; ok {
		cs.counters[name]++
		fmt.Fprintf(w, "OK!\n")
	} else {
		fmt.Fprintf(w, "Data not FOUND!")
	}
}

// lets implement rate limiting. limit the degree of concurrency of http server.
func limitnumClients(f http.HandlerFunc, maxclients int) http.HandlerFunc {
	//couting sema here.
	sema := make(chan struct{}, maxclients)

	return func(w http.ResponseWriter, r *http.Request) {
		sema <- struct{}{}
		defer func() {
			<-sema
		}()
		f(w, r)
	}
}

func main() {
	store := CoutnerStore{counters: map[string]int{"ram": 1, "sham": 2}}
	http.HandleFunc("/get", store.get)
	http.HandleFunc("/set", store.set)
	http.HandleFunc("/inc", store.inc)
	log.Println("HTTP server is going to start.....")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
