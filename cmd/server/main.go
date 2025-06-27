package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/per1Peteia/learn-pub-sub-starter/internal/pubsub"
	"github.com/per1Peteia/learn-pub-sub-starter/internal/routing"
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

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error channeling: %v", err)
	}
	err = pubsub.PublishJSON(ch, string(routing.ExchangePerilDirect), string(routing.PauseKey), routing.PlayingState{IsPaused: true})
	if err != nil {
		log.Fatalf("Error publishing message: %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Setup work here

	<-signalChan

	fmt.Println("Received termination signal, shutting down...")
}
