package main

import (
	"log"
	subscribeemails "subscribeemail"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	c, err := client.Dial(client.Options{
		HostPort:  client.DefaultHostPort,
		Namespace: client.DefaultNamespace,
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal Client. ", err)
	}
	defer c.Close()

	w := worker.New(c, subscribeemails.TaskQueueName, worker.Options{})

	w.RegisterWorkflow(subscribeemails.SubscriptionWorkflow)
	w.RegisterActivity(subscribeemails.SendEmail)

	log.Println("Worker is starting")

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start Worker. ", err)
	}
}
