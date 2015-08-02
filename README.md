
# Go FlagZ - Dynamic Flag Management

** work in progress **

The purpose of this project is to provide the ability to change `flag` or `spf13/pflag` values in runtime across many 
services based on a distributed configuration store (etcd, Consul).
 
## This sounds crazy. Why?

Dynamic configuration allows for fast iteration cycles, and provides valuable flexibility in emergencies. Two examples:
 
 * You want to enable a new feature for a certain fraction of user requests in granular fashion (1%, 5%, 20%, 50%, 
 100%) without a need to restart servers.
 * Your service is getting overloaded and you want to disable certain costly features, and you can't afford 
 restarting because you'd lose important capacity.
 
All of these uniformly across a shard of your services. Of course with great power comes great responsibility :)

## How?

Declare a single `flag.FlagSet` compatible variable in some package of your codebase (e.g. `common.DynamicFlagSet`) 
that you'll use throughout to declare your flags. Then:

```go
    // First parse the flags from the command line, as normal.
    common.DynamicFlagSet.Parse(os.Args[1:])
	updater, err := flagz_etcd.New(common.DynamicFlagSet, etcdClient, "/my_service/flagz", logger)
	if err != nil {
		logger.Fatalf("failed setting up %v", err)
	}
	// Read flagz from etcd and update their values in common.DynamicFlagSet
	err = updater.Initialize()
	if err != nil {
		log.Fatalf("failed setting up %v", err)
	}
	// Start listening for dynamic updates of flags.
	updater.Start()
```

In case of errors (parsing, disallowed values) flagz will atomically roll back the bad state in etcd.

# Status

Features planned:
 
  * HTTP `Handler` for displaying state of flags.
  * Monitoring of checksum of flags.
  * Implementation of flags that implement complicated types: maps, JSON Marshalled structs, protocol buffers.

Author: michal@improbable.io
