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

func main() {
	// creating RabbitMQ publisher
	publisher, err := rabbitmq.NewPublisher(
		"amqp://guest:guest@localhost:5672",
		rabbitmq.Config{},
		rabbitmq.WithPublisherOptionsLogging,
	)
	if err != nil {
		log.Fatal(err)
	}
	// close publisher after program finishes
	defer func() {
		err := publisher.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// if we send a message to an exchange and it cannot route to any queue
	// rabbitmq returns us that message
	returns := publisher.NotifyReturn()

	// we run return messages in a separate goroutine so that everytime an
	// exchange returns a message, we don't pause main thread
	go func() {
		for r := range returns {
			log.Printf("message returned from server: %s", string(r.Body))
		}
	}()

	// we want to make sure the message was delivered to the exchange
	confirmations := publisher.NotifyPublish()
	go func() {
		// here we could do something else with the message once it wasn't delivered to rabbitmq
		for c := range confirmations {
			log.Printf("message confirmed from server. tag: %v, ack: %v", c.DeliveryTag, c.Ack)
		}
	}()

	// block main thread - wait for shutdown signal
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println("signal received. shutting down producer program...")
		fmt.Println(sig)
		done <- true
	}()

	fmt.Println("awaiting signal")

	// ticker to send a message to a channel every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		// publish the message to rabbitmq exchange
		case <-ticker.C:
			err = publisher.Publish(
				[]byte("hello, this is a purchase done at "+time.Now().Format("02/01/2006 15:04:05")), // message itself
				[]string{"purchase-key"},                                   // routing keys
				rabbitmq.WithPublishOptionsContentType("application/json"), // format of the message
				rabbitmq.WithPublishOptionsMandatory,
				rabbitmq.WithPublishOptionsPersistentDelivery,
				rabbitmq.WithPublishOptionsExchange("market-exchange"), // to which exchange should we publish
			)
			if err != nil {
				log.Println(err)
			}
		// exit the program
		case <-done:
			fmt.Println("stopping publisher program, bye!")
			return
		}
	}

}
