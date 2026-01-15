package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

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
						for i := range len(results){
							println(results[i])
						}

            // Echo the data back to the connection.
						_, err = conn.Write([]byte(strings.Join(results, "\n")))
            if err != nil {
                log.Fatal(err)
            }
        }(conn)
    }
}
