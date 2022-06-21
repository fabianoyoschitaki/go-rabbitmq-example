# RabbitMQ with Go example

## Go application
go mod init github.com/fabianoyoschitaki/go-rabbitmq-fullcycle
go get github.com/wagslane/go-rabbitmq

### publisher: go run producer/main.go

- create market-exchange exchange
- create at least one queue bound to market-exchange (created purchases-completed queue) using purchase-key binding key
- publisher will publish a message every 2 seconds to market-exchange exchange using purchase-key
- it doesn't need to know to which queues it's going to be delivered
- exchange will deliver the message to purchases-completed queue since it's bound to market-exchange exchange through purchase-key binding key. there's no consumer at this point so the queue will keep growing

### consumer: go run consumer/main.go

- consumer simulates an application called notification-service
- consumer reads from purchases-completed using purchase-key binding key
- it doesn't need to know about who produced to the queue
