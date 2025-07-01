package main

import (
	"fmt"
	"log"

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

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error channeling: %v", err)
	}

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error getting username: %v", err)
	}

	gameState := gamelogic.NewGameState(userName)

	// detect pauses
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("pause.%s", userName),
		routing.PauseKey,
		1,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatalf("error subscribing: %v", err)
	}

	// detect player moves
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, userName),
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		1,
		handlerMove(gameState),
	)
	if err != nil {
		log.Fatalf("error subscribing: %v", err)
	}

	for {
		in := gamelogic.GetInput()
		if len(in) == 0 {
			continue
		}
		switch in[0] {
		case "spawn":
			err := gameState.CommandSpawn(in)
			if err != nil {
				log.Printf("%v", err)
			}
		case "move":
			move, err := gameState.CommandMove(in)
			if err != nil {
				log.Printf("%v", err)
			}

			err = pubsub.PublishJSON(ch, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, userName), move)
			if err != nil {
				log.Printf("%v", err)
			}
			fmt.Println("Army moved successfully...")
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unknown input")
			continue
		}
	}
}
