package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"jetstream/stream"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	defer nc.Close()

	jc, err := nc.JetStream()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	select {
	case <-jc.PublishAsyncComplete():
	}

	err = stream.CreateStreamIfNotExists(stream.StreamName, jc)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	go stream.PublishMessages(jc)

	locSub, err := stream.GetLocationsSubscription(jc)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	go gracefulShutdown(locSub)
	forever := make(chan int)
	<-forever
}

func gracefulShutdown(sub *nats.Subscription) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	signal.Notify(s, syscall.SIGTERM)
	go func() {
		<-s
		sub.Unsubscribe()
		sub.Drain()
		fmt.Println("Shutting down gracefully.")
		// clean up here
		os.Exit(0)
	}()
}
