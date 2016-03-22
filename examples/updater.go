package main

import (
	"flag"
	"log"
	"os"
	"time"

	etcd "github.com/coreos/etcd/client"
	flagz_etcd "github.com/mwitkow-io/go-flagz/etcd"
)

var (
	myFlagSet = flag.NewFlagSet("custom_flagset", flag.ContinueOnError)

	myString = myFlagSet.String("somestring", "initial_value", "someusage")
	myInt    = myFlagSet.Int("someint", 1337, "someusage int")
)

func main() {
	myFlagSet.Parse(os.Args[1:])
	logger := log.New(os.Stderr, "updater", 0)

	client, err := etcd.New(etcd.Config{Endpoints: []string{"http://localhost:2379"}})
	if err != nil {
		logger.Fatalf("Failed setting up %v", err)
	}
	kapi := etcd.NewKeysAPI(client)

	updater, err := flagz_etcd.New(myFlagSet, kapi, "/example/flagz", logger)
	if err != nil {
		logger.Fatalf("Failed setting up %v", err)
	}
	err = updater.Initialize()
	if err != nil {
		logger.Fatalf("Failed setting up %v", err)
	}
	updater.Start()

	for true {
		logger.Printf("someint: %v somestring: %v", *myInt, *myString)
		time.Sleep(1500 * time.Millisecond)
	}
}
