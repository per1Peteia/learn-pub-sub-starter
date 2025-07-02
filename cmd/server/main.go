package main

import (
	"fmt"
	"github.com/per1Peteia/learn-pub-sub-starter/internal/gamelogic"
	"github.com/per1Peteia/learn-pub-sub-starter/internal/pubsub"
	"github.com/per1Peteia/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

const CONN_URL = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril server...")

	gamelogic.PrintServerHelp()

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

	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		"game_logs.*",
		0,
		handlerLog(),
	)
	if err != nil {
		log.Fatalf("Error subscribing to log queue: %v", err)
	}

	for {
		in := gamelogic.GetInput()
		if len(in) == 0 {
			continue
		}
		switch in[0] {
		case "pause":
			fmt.Println("Broadcasting pause message...")
			err = pubsub.PublishJSON(ch, string(routing.ExchangePerilDirect), string(routing.PauseKey), routing.PlayingState{IsPaused: true})
			if err != nil {
				log.Fatalf("Error publishing message: %v", err)
			}
		case "resume":
			fmt.Println("Broadcasting resume message...")
			err = pubsub.PublishJSON(ch, string(routing.ExchangePerilDirect), string(routing.PauseKey), routing.PlayingState{IsPaused: false})
			if err != nil {
				log.Fatalf("Error publishing message: %v", err)
			}
		case "help":
			gamelogic.PrintServerHelp()
		case "quit":
			fmt.Println("Quitting...")
			return
		default:
			fmt.Println("Unknown input")
		}
	}

}
