package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/per1Peteia/learn-pub-sub-starter/internal/gamelogic"

	"github.com/per1Peteia/learn-pub-sub-starter/internal/pubsub"
	"github.com/per1Peteia/learn-pub-sub-starter/internal/routing"

	"github.com/rabbitmq/amqp091-go"
)

const CONN_URL = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril client...")

	conn, err := amqp091.Dial(CONN_URL)
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}

	defer conn.Close()
	fmt.Println("Connection successful")

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error getting username: %v", err)
	}

	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, userName), routing.PauseKey, 1)
	if err != nil {
		log.Fatalf("Error declaring and binding queue to channel: %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("RabbitMQ connection closed.")
}
