package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"jetstream/stream"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	sub, err := jc.Subscribe(stream.StreamName+"."+stream.LocationSubject,
		handleLocations,
		nats.MaxDeliver(3),
		nats.BackOff([]time.Duration{200 * time.Millisecond, 500 * time.Millisecond, time.Second}),
		nats.MaxAckPending(5),
		//nats.Durable("MONITOR"),
		//nats.StartTime(time.Date(2022, 7, 12, 15, 32, 12, 0, time.Local)), //2022-07-12T15:32:12+02:00
		nats.StartSequence(296),
	)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	go gracefulShutdown(sub)
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

func handleLocations(msg *nats.Msg) {
	fmt.Printf(">> Consumer: Location >> Payload:%s\n", string(msg.Data))
	msg.Ack()
}
