package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/go-etcd/etcd"
	flagz_etcd "github.com/mwitkow-io/go-flagz/etcd"
	"os"
	"time"
)

var (
	myFlagSet = flag.NewFlagSet("custom_flagset", flag.ContinueOnError)

	myString = myFlagSet.String("somestring", "initial_value", "someusage")
	myInt    = myFlagSet.Int("someint", 1337, "someusage int")
)

func main() {
	myFlagSet.Parse(os.Args[1:])
	logger := log.New()

	client := etcd.NewClient([]string{})
	updater, err := flagz_etcd.New(myFlagSet, client, "/example/flagz", logger)
	if err != nil {
		log.Fatalf("Failed setting up %v", err)
	}
	err = updater.Initialize()
	if err != nil {
		log.Fatalf("Failed setting up %v", err)
	}
	updater.Start()

	for true {
		logger.Infof("someint: %v somestring: %v", *myInt, *myString)
		time.Sleep(1500 * time.Millisecond)
	}
}
