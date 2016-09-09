// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

// Package kubernetes provides an a K8S ConfigMap watcher for the jobs systems.

package configmap

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/fsnotify/fsnotify"
	flag "github.com/spf13/pflag"
	"github.com/mwitkow/go-flagz"
)

const (
	k8sInternalsPrefix = ".."
	k8sDataSymlink = "..data"
)


var (
	errFlagNotDynamic = fmt.Errorf("flag is not dynamic")
	errFlagNotFound = fmt.Errorf("flag not found")
)

// Minimum logger interface needed.
// Default "log" and "logrus" should support these.
type loggerCompatible interface {
	Printf(format string, v ...interface{})
}

type Updater struct {
	started bool
	dirPath string
	watcher *fsnotify.Watcher
	flagSet *flag.FlagSet
	logger  loggerCompatible
	done    chan bool

}

func New(flagSet *flag.FlagSet, dirPath string, logger loggerCompatible) (*Updater, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("flagz: error initializing fsnotify watcher.")
	}
	return &Updater{
		flagSet: flagSet,
		logger:  logger,
		dirPath: dirPath,
		watcher: watcher,
	}, nil
}

func (u *Updater) Initialize() error {
	if u.started {
		return fmt.Errorf("flagz: already initialized updater.")
	}
	return u.readAll(/* allowNonDynamic */ false)
}

// Start kicks off the go routine that watches the directory for updates of values.
func (u *Updater) Start() error {
	if u.started {
		return fmt.Errorf("flagz: updater already started.")
	}
	u.watcher.Add(path.Join(u.dirPath, "..")) // add parent in case the dirPath is a symlink itself
	u.watcher.Add(u.dirPath) // add the dir itself.

	u.done = make(chan bool)
	go u.watchForUpdates()
	return nil
}

// Stops the auto-updating go-routine.
func (u *Updater) Stop() error {
	if !u.started {
		return fmt.Errorf("flagz: not updating")
	}
	u.done <- true
	u.watcher.Remove(u.dirPath)
	return nil
}

func (u *Updater) readAll(dynamicOnly bool) error {
	files, err := ioutil.ReadDir(u.dirPath)
	if err != nil {
		return fmt.Errorf("flagz: updater initialization: %v", err)
	}
	errorStrings := []string{}
	for _, f := range files {
		if strings.HasPrefix(path.Base(f.Name()), "..") {
			// skip random ConfigMap internals
			continue
		}
		fullPath := path.Join(u.dirPath, f.Name())
		if err := u.readFlagFile(fullPath, dynamicOnly); err != nil {
			if err == errFlagNotDynamic && dynamicOnly {
				// ignore
			} else {
				errorStrings = append(errorStrings, fmt.Sprintf("flag %v: %v", f.Name(), err.Error()))
			}
		}
	}
	if len(errorStrings) > 0 {
		return fmt.Errorf("encountered %d errors while parsing flags from directory  \n  %v",
			len(errorStrings), strings.Join(errorStrings, "\n"))
	}
	return nil
}


func (u *Updater) readFlagFile(fullPath string, dynamicOnly bool) error {
	flagName := path.Base(fullPath)
	flag := u.flagSet.Lookup(flagName)
	if flag == nil {
		return errFlagNotFound
	}
	if dynamicOnly && !flagz.IsFlagDynamic(flag) {
		return errFlagNotDynamic
	}
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}
	// do not call flag.Value.Set, instead go through flagSet.Set to change "changed" state.
	return u.flagSet.Set(flagName, string(content))
}

func (u *Updater) watchForUpdates() {
	u.logger.Printf("starting watching")
	for {
		select {
		case event := <-u.watcher.Events:
			if event.Name == u.dirPath || event.Name == path.Join(u.dirPath, k8sDataSymlink) {
				// case of the whole directory being re-symlinked
				switch event.Op {
				case fsnotify.Create:
					u.watcher.Add(u.dirPath)
					u.logger.Printf("flagz: Re-reading flags after ConfigMap update.")
					if err := u.readAll(/* dynamicOnly */ true); err != nil {
						u.logger.Printf("flagz: directory reload yielded errors: %v", err.Error())
					}
				case fsnotify.Remove:
				}

			} else if strings.HasPrefix(event.Name, u.dirPath) && !isK8sInternalDirectory(event.Name) {
				switch event.Op {
				case fsnotify.Create, fsnotify.Write, fsnotify.Rename:
					if err := u.readFlagFile(event.Name, true); err != nil {
						flagName := path.Base(event.Name)
						u.logger.Printf("flagz: failed setting flag %s: %v", flagName, err.Error())
					}
				}
			}

		case <-u.done:
			return
		}
	}
}

func isK8sInternalDirectory(filePath string) bool {
	basePath := path.Base(filePath)
	return strings.HasPrefix(basePath, k8sInternalsPrefix)
}