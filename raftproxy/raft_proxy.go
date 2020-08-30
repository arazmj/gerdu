package raftproxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/arazmj/gerdu/cache"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	raftTimeout         = 10 * time.Second
	retainSnapshotCount = 2
	tcpTimeout          = 10 * time.Second
)

type RaftCache interface {
	raft.FSM
	cache.UnImplementedCache
	OpenRaft(storage string) error
}

type RaftProxy struct {
	raft     *raft.Raft
	Imp      cache.UnImplementedCache
	raftAddr string
	joinAddr string
	localId  string
	RaftCache
}

type command struct {
	Op    string `json:"op,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func NewRaftProxy(imp cache.UnImplementedCache, raftAddr, joinAddr, localId string) *RaftProxy {
	return &RaftProxy{
		Imp:      imp,
		raftAddr: raftAddr,
		joinAddr: joinAddr,
		localId:  localId,
	}
}

// Put updates or insert a new entry, evicts the old entry
// if cache size is larger than capacity
func (c *RaftProxy) Put(key string, value string) (created bool) {
	cmd := &command{
		Op:    "put",
		Key:   key,
		Value: value,
	}

	future, err := c.applyCommand(cmd)

	if err != nil {
		log.Errorf("Error applyCommand %v", err)
		return false
	}

	if future.Error() != nil {
		log.Errorf("Error in raft apply future %v", future.Error())
		return false
	}

	return future.Response().(bool)
}

func (c *RaftProxy) Delete(key string) (ok bool) {
	cmd := &command{
		Op:  "delete",
		Key: key,
	}

	future, err := c.applyCommand(cmd)

	if err != nil {
		log.Errorf("Error applyCommand %v", err)
		return false
	}

	if future.Error() != nil {
		log.Fatalf("Error in raft apply future %v", future.Error())
		return false
	}

	return future.Response().(bool)
}

func (c *RaftProxy) Get(key string) (value string, ok bool) {
	cmd := &command{
		Op:  "get",
		Key: key,
	}

	future, err := c.applyCommand(cmd)

	if err != nil {
		log.Errorf("Error applyCommand %v", err)
		return "", false
	}

	if future.Error() != nil {
		log.Fatalf("Error in raft apply future %v", future.Error())
		return "", false
	}

	response := future.Response().(getResponse)
	return response.value, response.ok
}

func (c *RaftProxy) applyCommand(cmd *command) (raft.ApplyFuture, error) {
	if c.raft.State() != raft.Leader {
		return nil, errors.New(fmt.Sprintf("not a leader but a %v %p", c.raft.State(), c.raft))
	}

	b, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}

	return c.raft.Apply(b, raftTimeout), nil
}

type getResponse struct {
	value string
	ok    bool
}

type fsm RaftProxy

func (f *fsm) Apply(l *raft.Log) interface{} {
	var cmd command
	if err := json.Unmarshal(l.Data, &cmd); err != nil {
		log.Fatalf("failed to unmarshal command: %s", err.Error())
	}

	log.Infof("Apply command: %v", cmd)
	switch cmd.Op {
	case "get":
		value, ok := f.Imp.Get(cmd.Key)
		response := getResponse{
			value: value,
			ok:    ok,
		}
		return response
	case "put":
		return f.Imp.Put(cmd.Key, cmd.Value)
	case "delete":
		return f.Imp.Delete(cmd.Key)
	default:
		log.Fatalf("unrecognized command op: %s", cmd.Op)
	}
	return nil
}

func (c *RaftProxy) OpenRaft(storage string) error {
	// Setup Raft configuration.
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(c.localId)

	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", c.raftAddr)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(c.raftAddr, addr, 3, tcpTimeout, os.Stderr)
	if err != nil {
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(storage, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Create the log store and stable store.
	var logStore raft.LogStore
	var stableStore raft.StableStore
	if storage == "" {
		logStore = raft.NewInmemStore()
		stableStore = raft.NewInmemStore()
	} else {
		boltDB, err := raftboltdb.NewBoltStore(filepath.Join(storage, "raft.db"))
		if err != nil {
			return fmt.Errorf("new bolt store: %s", err)
		}
		logStore = boltDB
		stableStore = boltDB
	}

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(config, (*fsm)(c), logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}

	c.raft = ra

	if c.joinAddr == "" {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	} else {
		b, err := json.Marshal(map[string]string{"addr": c.raftAddr, "id": c.localId})
		if err != nil {
			return err
		}
		resp, err := http.Post(fmt.Sprintf("http://%s/join", c.joinAddr), "", bytes.NewReader(b))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		return nil

	}

	return nil
}

// Join joins a node, identified by nodeID and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
func (c *RaftProxy) Join(nodeID, addr string) error {
	log.Infof("received join request for remote node %s at %s", nodeID, addr)

	configFuture := c.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		log.Errorf("failed to get raft configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				log.Warnf("node %s at %s already member of cluster, ignoring join request", nodeID, addr)
				return nil
			}

			future := c.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := c.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	log.Infof("node %s at %s joined successfully", nodeID, addr)
	return nil
}
