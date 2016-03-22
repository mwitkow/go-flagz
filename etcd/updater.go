// Updaterof Go "flags"-compatible data base on dynamic etcd watches.
//
// Copyright 2015 Michal Witkowski. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package etcd provides an updater for go "flags"-compatible FlagSets based on dynamic changes in etcd storage.

package etcd

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	etcd "github.com/coreos/etcd/client"
)

// Controls the auto updating process of a "flags"-compatible package from Etcd.
type Updater struct {
	client    etcd.Client
	etcdKeys  etcd.KeysAPI
	flagSet   flagSet
	logger    logger
	etcdPath  string
	lastIndex uint64
	watching  bool
	context   context.Context
	cancel    context.CancelFunc
}

// Minimum interface needed to support dynamic flags.
// As implemented by "flag" and "spf13/pflag".
type flagSet interface {
	Set(name, value string) error
}

// Minimum logger interface needed.
// Default "log" and "logrus" should support these.
type logger interface {
	Printf(format string, v ...interface{})
}

func New(set flagSet, keysApi etcd.KeysAPI, etcdPath string, logger logger) (*Updater, error) {
	u := &Updater{
		flagSet:   set,
		etcdKeys:  keysApi,
		etcdPath:  etcdPath,
		logger:    logger,
		lastIndex: 0,
		watching:  false,
	}
	u.context, u.cancel = context.WithCancel(context.Background())
	return u, nil
}

// Performs the initial read of etcd for all flags and updates the specified FlagSet.
func (u *Updater) Initialize() error {
	if u.lastIndex != 0 {
		return fmt.Errorf("flagz: already initialized.")
	}
	return u.readAllFlags()
}

// Starts the auto-updating go-routine.
func (u *Updater) Start() error {
	if u.lastIndex == 0 {
		return fmt.Errorf("flagz: not initialized")
	}
	if u.watching {
		return fmt.Errorf("flagz: already watching")
	}
	u.watching = true
	go u.watchForUpdates()
	return nil
}

// Stops the auto-updating go-routine.
func (u *Updater) Stop() error {
	if !u.watching {
		return fmt.Errorf("flagz: not watching")
	}
	u.logger.Printf("flagz: stopping")
	u.cancel()
	return nil
}

func (u *Updater) readAllFlags() error {
	resp, err := u.etcdKeys.Get(u.context, u.etcdPath, &etcd.GetOptions{Recursive: true, Sort: true})
	if err != nil {
		return err
	}
	u.lastIndex = resp.Index
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
			len(errorStrings), strings.Join(errorStrings, "\n"))
	}
	return nil
}

func (u *Updater) watchForUpdates() error {
	// We need to implement our own watcher because the one in go-etcd doesn't handle errorcode 400 and 401.
	// See https://github.com/coreos/etcd/blob/master/Documentation/errorcode.md
	// And https://coreos.com/etcd/docs/2.0.8/api.html#waiting-for-a-change
	watcher := u.etcdKeys.Watcher(u.etcdPath, &etcd.WatcherOptions{AfterIndex: u.lastIndex, Recursive: true})
	u.logger.Printf("flagz: watcher started")
	for u.watching {
		resp, err := watcher.Next(u.context)
		if etcdErr, ok := err.(etcd.Error); ok && etcdErr.Code == etcd.ErrorCodeEventIndexCleared {
			// Our index is out of the Etcd Log. Reread everything and reset index.
			u.logger.Printf("flagz: handling Etcd Index error by re-reading everything: %v", err)
			time.Sleep(200 * time.Millisecond)
			u.readAllFlags()
			watcher = u.etcdKeys.Watcher(u.etcdPath, &etcd.WatcherOptions{AfterIndex: u.lastIndex, Recursive: true})
			continue
		} else if clusterErr, ok := err.(*etcd.ClusterError); ok {
			u.logger.Printf("flagz: etcd ClusterError. Will retry. %v", clusterErr.Detail())
			time.Sleep(100 * time.Millisecond)
			continue
		} else if err == context.DeadlineExceeded {
			u.logger.Printf("flagz: deadline exceeded which watching for changes, continuing watching")
			continue
		} else if err == context.Canceled {
			break
		} else if err != nil {
			u.logger.Printf("flagz: wicked etcd error. Restarting watching after some time. %v", err)
			// Etcd started dropping watchers, or is re-electing. Give it some time.
			randOffsetMs := int(500 * rand.Float32())
			time.Sleep(1*time.Second + time.Duration(randOffsetMs)*time.Millisecond)
			continue
		}
		u.lastIndex = resp.Node.ModifiedIndex
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
			u.logger.Printf("flagz: ignoring action=%v on flag=%v at etcdindex=%v", flagName, u.lastIndex)
			continue
		}
		err = u.flagSet.Set(flagName, value)
		if err != nil {
			u.logger.Printf("flagz: failed updating flag=%v, because of: %v", flagName, err)
			u.rollbackEtcdValue(flagName, resp)
		} else {
			u.logger.Printf("flagz: updated flag=%v to value=%v at etcdindex=%v", flagName, value, u.lastIndex)
		}
	}
	u.logger.Printf("flagz: watcher exited")
	return nil
}

func (u *Updater) rollbackEtcdValue(flagName string, resp *etcd.Response) {
	var err error
	if resp.PrevNode != nil {
		// It's just a new value that's wrong, roll back to prevNode value atomically.
		_, err = u.etcdKeys.Set(u.context, resp.Node.Key, resp.PrevNode.Value, &etcd.SetOptions{PrevIndex: u.lastIndex})
	} else {
		_, err = u.etcdKeys.Delete(u.context, resp.Node.Key, &etcd.DeleteOptions{PrevIndex: u.lastIndex})
	}
	if etcdErr, ok := err.(etcd.Error); ok && etcdErr.Code == etcd.ErrorCodeTestFailed {
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
	name := parts[len(parts)-1]
	return name, nil
}
