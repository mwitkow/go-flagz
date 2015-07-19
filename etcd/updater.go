//


package etcd


import (
	"github.com/coreos/go-etcd/etcd"
	"strings"
	"fmt"
	"time"
	"math/rand"
)



type Updater struct {
	client        etcd.Client
	flagSet       flagSet
	logger        logger
	etcdPath      string
	lastEtcdIndex uint64
	watching      bool
	watchStop     chan bool
}

// TODO: Figure out a struct `etcdCodes` here?
const (
	preconditionFailed = 101
	errorWatcherCleared = 400
	errorEventIndexCleared = 401
)

// Minimum interface needed to support dynamic flags.
// As implemented by "flag" and "spf13/pflag".
type flagSet interface {
	Set(name, value string) error
}

// Minimum logger interface needed.
// Default "log" and "logrus" should support these.
type logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

func New(set flagSet, client etcd.Client, etcdPath string, logger logger) (*Updater, error) {
	return &Updater{
		flagSet: set,
		client: client,
		etcdPath: etcdPath,
		logger: logger,
		lastEtcdIndex: 0,
		watching: false,
		watchStop: make(chan bool),
	}, nil
}

// Performs the initial of Etcd for all flags and updates the specified FlagSet.
func (u *Updater) Initialize() error {
	if u.lastEtcdIndex != 0 {
		return fmt.Errorf("flagz: already initialized.")
	}
	return u.readAllFlags()
}

func (u *Updater) Start() error {
	if u.lastEtcdIndex == 0 {
		return fmt.Errorf("flagz: not initialized")
	}
	if u.watching {
		return fmt.Errorf("flagz: already watching")
	}
	go u.watchForUpdates()
	return nil
}

func (u *Updater) Stop() error {
	if !u.watching {
		return fmt.Errorf("flagz: not watching")
	}
	u.watchStop <- true
	return nil
}

func (u *Updater) readAllFlags() error {
	resp, err := u.client.Get(u.etcdPath, /* sort */
		true, /* recursive */
		true)
	if err != nil {
		return err
	}
	u.lastEtcdIndex = resp.EtcdIndex
	errorStrings := []string{}
	for _, node := range resp.Node.Nodes {
		if node.Dir {
			u.logger.Printf("flagz: ignoring subdirectory %v", node.Key)
			continue
		}
		flagName, err := keyToFlag(node.Key)
		if err != nil {
			errorStrings = append(errorStrings, err.Error())
		}
		if node.Value != "" {
			err := u.flagSet.Set(flagName, node.Value)
			if err != nil {
				errorStrings = append(errorStrings, err.Error())
			}
		}
	}
	if len(errorStrings) > 0 {
		return fmt.Errorf("flagz: encountered %d errors while parsing flags from etcd: \n  %v",
			strings.Join(errorStrings, "\n"))
	}
	return nil
}

func (u * Updater) watchForUpdates() error {
	// We need to implement our own watcher because the one in go-etcd doesn't handle errorcode 400 and 401.
	// See https://github.com/coreos/etcd/blob/master/Documentation/errorcode.md
	// And https://coreos.com/etcd/docs/2.0.8/api.html#waiting-for-a-change
	for u.watching {
		resp, err := u.client.Watch(
			u.etcdPath,
			u.lastEtcdIndex + 1,
			/*recursive*/
			true,
			/* recvChan*/
			nil,
			/* stopChan */
			u.watchStop)
		if etcdErr, ok := err.(etcd.EtcdError); ok && etcdErr.ErrorCode == errorEventIndexCleared {
			// Our index is out of the Etcd Log. Reread everything and reset index.
			u.logger.Printf("flagz: handling Etcd Index error by re-reading everything: %v", err)
			time.Sleep(200 * time.Millisecond)
			u.readAllFlags()
			continue
		} else if err != nil {
			u.logger.Printf("flagz: wicked etcd error. Restarting watching after some time. %v", err)
			// Etcd started dropping watchers, or is re-electing. Give it some time.
			randOffsetMs := int(500 * rand.Float32())
			time.Sleep(1 * time.Second + time.Duration(randOffsetMs) * time.Millisecond)
			continue
		}
		u.lastEtcdIndex = resp.Node.ModifiedIndex
		if resp.Node.Dir {
			u.logger.Printf("flagz: ignoring directory %v", resp.Node.Key)
			continue
		}
		flagName, err := keyToFlag(resp.Node.Key)
		if err != nil {
			continue
		}
		value := resp.Node.Value
		if value == "" {
			u.logger.Printf("flagz: ignoring action=%v on flag=%v at etcdindex=%v", flagName, u.lastEtcdIndex)
			continue
		}
		err = u.flagSet.Set(flagName, value)
		if err != nil {
			u.logger.Printf("flagz: failed updating flag=%v, because of: %", flagName, err)
			u.rollbackEtcdValue(flagName, resp)
		} else {
			u.logger.Printf("flagz: updated flag=%v to value=%v at etcdindex=%v", flagName, value, u.lastEtcdIndex)
		}
	}
	return nil
}

func (u *Updater) rollbackEtcdValue(flagName string, resp *etcd.Response) {
	var err error
	if resp.PrevNode != nil {
		// It's just a new value that's wrong, roll back to prevNode value atomically.
		_, err = u.client.CompareAndSwap(
			resp.Node.Key,
			resp.PrevNode.Value,
			/*ttl*/
			0,
			/* prevValue */
			"",
			u.lastEtcdIndex)

	} else {
		// This was a create and the value is botched. Delete it.
		_, err = u.client.CompareAndDelete(
			resp.Node.Key,
			/* prevValue */
			"",
			u.lastEtcdIndex)
	}
	if etcdErr, ok := err.(etcd.EtcdError); ok && etcdErr.ErrorCode == preconditionFailed {
		// Someone probably rolled it back in the mean time.
		u.logger.Printf("flagz: rolled back flag=%v was changed by someone else. All good.", flagName)
	} else if err != nil {
		u.logger.Printf("flagz: rolling back flagz=%v failed: %v", flagName, err)
	} else {
		u.logger.Printf("flagz: rolled back flagz=%v to correct state. All good.", flagName)
	}
}

func keyToFlag(etcdKey string) (string, error) {
	parts := strings.Split(etcdKey, "/")
	if len(parts) <= 1 {
		return "", fmt.Errorf("flagz: can't extract flagName")
	}
	name := parts[len(parts) - 1]
	return name, nil
}



