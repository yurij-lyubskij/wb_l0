package main

import (
	nats "github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
	"log"
	"sync"
	"time"
	_ "yuriy_test/models"
)

func main() {

	clusterID := "test-cluster"
	clientID := "stan-pub"
	URL := stan.DefaultNatsURL
	// Connect Options.
	opts := []nats.Option{nats.Name("NATS Streaming Example Publisher")}
	// Connect to NATS
	nc, err := nats.Connect(URL, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	sc, err := stan.Connect(clusterID, clientID, stan.NatsConn(nc))
	if err != nil {
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, URL)
	}
	defer sc.Close()

	subj := "msg"

	ch := make(chan bool)
	var glock sync.Mutex
	var guid string
	acb := func(lguid string, err error) {
		glock.Lock()
		log.Printf("Received ACK for guid %s\n", lguid)
		defer glock.Unlock()
		if err != nil {
			log.Fatalf("Error in server ack for guid %s: %v\n", lguid, err)
		}
		if lguid != guid {
			log.Fatalf("Expected a matching guid in ack callback, got %s vs %s\n", lguid, guid)
		}
		ch <- true
	}
	//по таймеру
	ticker := time.NewTicker(5 * time.Second)
	i := 0
	for _ = range ticker.C {
		i++
		glock.Lock()
		msg, err := randomJson()
		if err != nil {
			log.Fatalf("Message can't be sent: %v\n", err)
		}
		guid, err = sc.PublishAsync(subj, msg, acb)
		if err != nil {
			log.Fatalf("Error during async publish: %v\n", err)
		}
		glock.Unlock()
		if guid == "" {
			log.Fatal("Expected non-empty guid to be returned.")
		}
		log.Printf("Published [%s] : '%s' [guid: %s]\n", subj, msg, guid)

		select {
		case <-ch:
			break
		case <-time.After(5 * time.Second):
			log.Fatal("timeout")
		}
		if i >= 10 {
			// надо останавливать
			ticker.Stop()
			break
		}
	}
}
