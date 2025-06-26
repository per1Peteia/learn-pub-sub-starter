package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

const CONN_URL = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril server...")

	conn, err := amqp.Dial(CONN_URL)
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}

	defer conn.Close()
	fmt.Println("Connection successful")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Setup work here

	<-signalChan

	fmt.Println("Received termination signal, shutting down...")
}
