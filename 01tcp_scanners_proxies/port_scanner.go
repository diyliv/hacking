package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
)

const NUM_JOBS = 2526

func scan(wId int, jobs <-chan int, resp chan<- int) {
	for j := range jobs {
		// fmt.Printf("Worker Id: %d started scanning: %v\n", wId, j)
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", j))
		if err != nil {
			resp <- 0
			continue
		}
		conn.Close()
		resp <- j
	}
}

func main() {
	jobs := make(chan int, NUM_JOBS)
	resp := make(chan int)

	openPorts := make([]int, 0)

	for w := 0; w < cap(jobs); w++ {
		go scan(w, jobs, resp)
	}

	go func() {
		for i := 1; i <= NUM_JOBS; i++ {
			jobs <- i
		}
	}()

	for a := 1; a < NUM_JOBS; a++ {
		port := <-resp
		if port != 0 {
			openPorts = append(openPorts, port)
		}
	}

	defer func() {
		close(jobs)
		close(resp)
	}()

	log.Printf("starting http server on port :8080")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var data Port
		w.Header().Add("Content-Type", "application/json")
		for _, p := range openPorts {
			fmt.Printf("[%d] open\n", p)
			data.Ports = openPorts
		}

		if err := json.NewEncoder(w).Encode(&data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			panic(err)
		}
		w.WriteHeader(http.StatusOK)
	})

	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			panic(err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)
	<-done
}
