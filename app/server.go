package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type SearchResult struct{
	Title string `json:"title"`
	URL string `json:"url"`
}

func listen(db *sql.DB, dict Dictionary){
	socket, err := net.Listen("unix","/tmp/koudelka_socket")
	if err != nil {
		log.Fatal(err)
	}

	c:= make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func(){
		<-c 
		os.Remove("/tmp/koudelka_socket")
		os.Exit(1)
	}()

	println("listening...")

	for {
        // Accept an incoming connection.
        conn, err := socket.Accept()
        if err != nil {
            log.Fatal(err)
        }

        // Handle the connection in a separate goroutine.
        go func(conn net.Conn) {
            defer conn.Close()
            // Create a buffer for incoming data.
            buf := make([]byte, 4096)

            // Read data from the connection.
            n, err :=conn.Read(buf)
            if err != nil {
                log.Fatal(err)
            }

						query := string(buf[:n])
						results := search(db,query, dict)
						jsonData, err := json.Marshal(results)
						if err != nil {
							log.Println(err)
							return
						}
            // Echo the data back to the connection.
						_, err = conn.Write(jsonData)
            if err != nil {
                log.Fatal(err)
            }
        }(conn)
    }
}
