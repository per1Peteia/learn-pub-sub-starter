package main

import (
	"fmt"

	"github.com/per1Peteia/learn-pub-sub-starter/internal/gamelogic"
	"github.com/per1Peteia/learn-pub-sub-starter/internal/pubsub"
	"github.com/per1Peteia/learn-pub-sub-starter/internal/routing"
)

func handlerLog() func(gl routing.GameLog) pubsub.Acktype {
	return func(gl routing.GameLog) pubsub.Acktype {
		defer fmt.Print("> ")
		if err := gamelogic.WriteLog(gl); err != nil {
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
