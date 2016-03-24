// Copyright 2015 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package main

import (
	"log"

	"os"
	"time"

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
	logger := log.New(os.Stderr, "updater", 0)

	client, err := etcd.New(etcd.Config{Endpoints: []string{"http://localhost:2379"}})
	if err != nil {
		logger.Fatalf("Failed setting up %v", err)
	}
	w, err := watcher.New(myFlagSet, etcd.NewKeysAPI(client), "/example/flagz", logger)
	if err != nil {
		logger.Fatalf("Failed setting up %v", err)
	}
	err = w.Initialize()
	if err != nil {
		logger.Fatalf("Failed setting up %v", err)
	}
	w.Start()

	for true {
		logger.Printf("someint: %v somestring: %v", *myInt, *myString)
		time.Sleep(1500 * time.Millisecond)
	}
}
