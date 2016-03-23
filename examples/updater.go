package main

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	etcd "github.com/coreos/etcd/client"
	"github.com/mwitkow/go-flagz/watcher"
	flag "github.com/spf13/pflag"
)

var (
	myFlagSet = flag.NewFlagSet("custom_flagset", flag.ContinueOnError)

	myString = myFlagSet.String("somestring", "initial_value", "someusage")
	myInt    = myFlagSet.Int("someint", 1337, "someusage int")
)

func main() {
	myFlagSet.Parse(os.Args[1:])
	logger := log.New()

	client, err := etcd.New(etcd.Config{Endpoints: []string{"http://localhost:2379"}})
	if err != nil {
		log.Fatalf("Failed setting up %v", err)
	}
	w, err := watcher.New(myFlagSet, etcd.NewKeysAPI(client), "/example/flagz", logger)
	if err != nil {
		log.Fatalf("Failed setting up %v", err)
	}
	err = w.Initialize()
	if err != nil {
		log.Fatalf("Failed setting up %v", err)
	}
	w.Start()

	for true {
		logger.Infof("someint: %v somestring: %v", *myInt, *myString)
		time.Sleep(1500 * time.Millisecond)
	}
}
