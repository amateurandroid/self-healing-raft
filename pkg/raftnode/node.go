package raftnode

import (
    "log"

    "go.etcd.io/etcd/raft/v3"
    "go.etcd.io/etcd/raft/v3/raftpb"
)

// RaftNode structure
type RaftNode struct {
    ID      int
    Addr    string
    raft    raft.Node
    storage *raft.MemoryStorage
}

// NewRaftNode initializes a Raft node
func NewRaftNode(id int, addr string) *RaftNode {
    storage := raft.NewMemoryStorage()

    cfg := &raft.Config{
        ID:              uint64(id),
        ElectionTick:    10,
        HeartbeatTick:   1,
        Storage:         storage,
        MaxSizePerMsg:   4096,
        MaxInflightMsgs: 256,
    }

    peers := []raft.Peer{{ID: uint64(id)}}
    node := raft.StartNode(cfg, peers)

    return &RaftNode{
        ID:      id,
        Addr:    addr,
        raft:    node,
        storage: storage,
    }
}

// Start runs the Raft node loop
func (n *RaftNode) Start() {
    log.Printf("[Node %d] Raft node running at %s", n.ID, n.Addr)

    go func() {
        for {
            rd := <-n.raft.Ready()

            // TODO: Apply log entries here later
            for _, entry := range rd.CommittedEntries {
                if entry.Type == raftpb.EntryNormal && entry.Data != nil {
                    log.Printf("[Node %d] Committed entry: %s", n.ID, string(entry.Data))
                }
            }

            // Advance the node
            n.raft.Advance()
        }
    }()
}
func (n *RaftNode) RaftPropose(data []byte) {
    if err := n.raft.Propose(nil, data); err != nil {
        log.Printf("[Node %d] Failed to propose: %v", n.ID, err)
    }
}