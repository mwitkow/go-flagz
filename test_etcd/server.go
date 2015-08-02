// Go functions for starting a test-local etcd server.
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

// Package test_etcd provides helper functions to start up an etcd server for the purpose of integration testing.

package test_etcd

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"time"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type EtcdTestServer struct {
	errWriter  io.Writer
	etcdServer *exec.Cmd
	etcdDir    string
	Client     etcd.Client
	Endpoint   string
}

func New(etcdErrWriter io.Writer) (*EtcdTestServer, error) {
	s := &EtcdTestServer{errWriter: etcdErrWriter}
	endpointAddress, err := allocateLocalAddress()
	if err != nil {
		return nil, fmt.Errorf("failed allocating server addr: %v", err)
	}
	peerAddress, err := allocateLocalAddress()
	if err != nil {
		return nil, fmt.Errorf("failed allocating server addr: %v", err)
	}
	etcdBinary, err := exec.LookPath("etcd")
	if err != nil {
		return nil, err
	}
	s.etcdDir, err = ioutil.TempDir("/tmp", "etcd_testserver")
	if err != nil {
		return nil, fmt.Errorf("failed allocating new dir: %v", err)
	}
	endpoint := "http://" + endpointAddress
	peer := "http://" + peerAddress
	s.etcdServer = exec.Command(
		etcdBinary,
		"--log-package-levels=etcdmain=WARNING,etcdserver=WARNING,raft=WARNING",
		"--force-new-cluster="+"true",
		"--data-dir="+s.etcdDir,
		"--listen-peer-urls="+peer,
		"--initial-cluster="+"default="+peer+"",
		"--initial-advertise-peer-urls="+peer,
		"--advertise-client-urls="+endpoint,
		"--listen-client-urls="+endpoint)
	s.etcdServer.Stderr = s.errWriter
	s.etcdServer.Stdout = ioutil.Discard
	s.Endpoint = endpoint
	if err := s.etcdServer.Start(); err != nil {
		s.Stop()
		return nil, fmt.Errorf("cannot start etcd: %v", err)
	}
	s.Client, err = etcd.New(etcd.Config{Endpoints: []string{endpoint}})
	if err != nil {
		s.Stop()
		return s, fmt.Errorf("failed allocating client: %v", err)
	}
	time.Sleep(3 * time.Second)
	if err := s.Client.Sync(context.Background()); err != nil {
		s.Stop()
		return s, fmt.Errorf("failed connecting to test etcd server: %v", err)
	}
	return s, nil
}

func (s *EtcdTestServer) Stop() {
	var err error
	if s.etcdServer != nil {
		if err := s.etcdServer.Process.Kill(); err != nil {
			fmt.Printf("failed killing etcd process: %v")
		}
		// Just to make sure we actually finish it before continuing.
		s.etcdServer.Wait()
	}
	if s.etcdDir != "" {
		if err = os.RemoveAll(s.etcdDir); err != nil {
			fmt.Printf("failed clearing temporary dir: %v", err)
		}
	}
}

func allocateLocalAddress() (string, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer l.Close()
	return l.Addr().String(), nil
}
