package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wagslane/go-rabbitmq"
)

var consumerName = "notification-service"

func main() {
	// creating rabbitmq consumer
	consumer, err := rabbitmq.NewConsumer(
		"amqp://guest:guest@localhost",
		rabbitmq.Config{},
		rabbitmq.WithConsumerOptionsLogging,
	)
	if err != nil {
		log.Fatal(err)
	}

	// closing consumer after program termination
	defer func() {
		err := consumer.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// started consuming messages from purchases-completed queue through purchase-key binding key
	err = consumer.StartConsuming(
		// function to process message from queue
		func(d rabbitmq.Delivery) (action rabbitmq.Action) {
			log.Printf("message incoming: %v", string(d.Body))
			time.Sleep(5 * time.Second) // pretending we're doing something with it
			log.Printf("message processed successfully: %v", string(d.Body))

			// we can return rabbitmq.Ack | NackDiscard | NackRequeue
			return rabbitmq.Ack // message gets discarded after successfully processed
		},
		"purchases-completed",                      // queue name from which we want to consume
		[]string{"purchase-key"},                   // binding key (from exchange)
		rabbitmq.WithConsumeOptionsConcurrency(10), // consuming 10 messages instead of 1
		rabbitmq.WithConsumeOptionsQueueDurable,
		rabbitmq.WithConsumeOptionsConsumerName(consumerName),
	)
	if err != nil {
		log.Fatal(err)
	}

	// block main thread - wait for shutdown signal
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println("signal received. shutting down consumer program...")
		fmt.Println(sig)
		done <- true
	}()

	fmt.Println("awaiting signal")
	<-done // blocking channel waiting for done
	fmt.Println("stopping consumer program, bye!")
}
