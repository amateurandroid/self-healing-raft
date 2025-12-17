package raftnode

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/amateurandroid/self-healing-raft/pkg/store"
	"github.com/gogo/protobuf/proto"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type RaftNode struct {
	ID      uint64
	Addr    string
	Peers   map[uint64]string
	raft    raft.Node
	storage *raft.MemoryStorage
	kv      *store.KVStore
	client  *http.Client
	cfg     *raft.Config // Store config for adaptive timeout

	// RTT smoothing
	rttMu      sync.Mutex
	rttEWMA    float64 // ms
	jitterEWMA float64 // ms
	alpha      float64 // EWMA smoothing factor
	randSrc    *rand.Rand
}

func NewRaftNode(id int, addr string, peers map[uint64]string, kv *store.KVStore) *RaftNode {
	storage := raft.NewMemoryStorage()
	cfg := &raft.Config{
		ID:              uint64(id),
		// CHANGED: Increased from 10 to 50 to handle injected latency safely
		ElectionTick:    50, 
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
		CheckQuorum:     true,
	}

	var initialPeers []raft.Peer
	for pid := range peers {
		initialPeers = append(initialPeers, raft.Peer{ID: pid})
	}

	// CHANGED: Add Node ID to the seed to ensure unique randomness per node
	seed := time.Now().UnixNano() + int64(id)*100000
	r := rand.New(rand.NewSource(seed))

	rn := &RaftNode{
		ID:      uint64(id),
		Addr:    addr,
		Peers:   peers,
		raft:    raft.StartNode(cfg, initialPeers),
		storage: storage,
		kv:      kv,
		client:  &http.Client{Timeout: 2 * time.Second},
		cfg:     cfg,
		alpha:   0.25,
		randSrc: r,
	}

	rn.initCSV()
	return rn
}

func (n *RaftNode) Start() {
	log.Printf("[Node %d]Raft node started at %s", n.ID, n.Addr)

	// 1) Ticking - TUNED: Faster tick (20ms) to make system sensitive to latency
	go func() {
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			n.raft.Tick()
		}
	}()

	// 2) Process Ready State
	go func() {
		for rd := range n.raft.Ready() {
			if err := n.storage.Append(rd.Entries); err != nil {
				log.Printf("[Node %d] storage append error: %v", n.ID, err)
			}
			for _, msg := range rd.Messages {
				n.send(msg)
			}
			for _, entry := range rd.CommittedEntries {
				if entry.Type == raftpb.EntryNormal && len(entry.Data) != 0 {
					n.apply(entry.Data)
				}
			}
			n.raft.Advance()
		}
	}()

	// 3) Start RTT monitor
	go n.monitorRTT(2 * time.Second) // Check network every 2s
}

func (n *RaftNode) RaftPropose(data []byte) {
	if err := n.raft.Propose(context.TODO(), data); err != nil {
		log.Printf("[Node %d] propose error: %v", n.ID, err)
	}
}

func (n *RaftNode) HandleRaft(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	var msg raftpb.Message
	if err := proto.Unmarshal(body, &msg); err != nil {
		http.Error(w, "unmarshal error", http.StatusBadRequest)
		return
	}

	if err := n.raft.Step(context.TODO(), msg); err != nil {
		http.Error(w, "step error: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (n *RaftNode) send(m raftpb.Message) {
	if m.To == n.ID {
		return
	}

	// INJECTED LATENCY (Simulate Hosted Server Lag)
	// Delays outbound messages by 50-100ms
	delay := time.Duration(50+rand.Intn(50)) * time.Millisecond
	time.Sleep(delay)

	addr := n.Peers[m.To]
	data, err := proto.Marshal(&m)
	if err != nil {
		log.Printf("[Node %d] marshal error: %v", n.ID, err)
		return
	}

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/raft", addr), bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/octet-stream")

	_, err = n.client.Do(req)
	if err != nil && !strings.Contains(err.Error(), "connection refused") {
		log.Printf("[Node %d] Send failed to %d (%s): %v", n.ID, m.To, addr, err)
	}
}

func (n *RaftNode) apply(data []byte) {
	parts := strings.SplitN(string(data), "=", 2)
	if len(parts) == 2 {
		n.kv.Put(parts[0], parts[1])
		log.Printf("[Node %d] Applied: %s=%s", n.ID, parts[0], parts[1])
	}
}

// PingPeer performs a lightweight HTTP GET /probe to the peer and returns RTT.
func (n *RaftNode) PingPeer(peerID uint64) (time.Duration, error) {
	addr, ok := n.Peers[peerID]
	if !ok || addr == "" {
		return 0, fmt.Errorf("peer %d unknown", peerID)
	}

	url := fmt.Sprintf("http://%s/probe", addr)
	start := time.Now()

	// INJECTED LATENCY (So Adaptive Logic "Sees" the Lag)
	// This ensures the RTT calculation includes the simulated network delay
	delay := time.Duration(50+rand.Intn(50)) * time.Millisecond
	time.Sleep(delay)

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := n.client.Do(req)
	if err != nil {
		return 0, err
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return time.Since(start), nil
}

func (n *RaftNode) monitorRTT(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		st := n.raft.Status()
		leader := st.Lead

		var target uint64
		if leader != 0 && leader != n.ID {
			target = leader
		} else {
			for pid := range n.Peers {
				if pid != n.ID {
					target = pid
					break
				}
			}
		}

		if target == 0 {
			continue
		}

		rtt, err := n.PingPeer(target)
		if err != nil {
			continue
		}

		// AUTOMATION CHECK: Only adapt if mode is not "fixed"
		if os.Getenv("RAFT_MODE") != "fixed" {
			n.updateRTTStatsAndTimeout(rtt)
		}
	}
}

func (n *RaftNode) updateRTTStatsAndTimeout(rtt time.Duration) {
	ms := float64(rtt.Milliseconds())

	n.rttMu.Lock()
	if n.rttEWMA == 0 {
		n.rttEWMA = ms
		n.jitterEWMA = 0
	} else {
		alpha := n.alpha
		n.rttEWMA = alpha*ms + (1-alpha)*n.rttEWMA
		dev := math.Abs(ms - n.rttEWMA)
		n.jitterEWMA = alpha*dev + (1-alpha)*n.jitterEWMA
	}
	ewma := n.rttEWMA
	jitter := n.jitterEWMA
	n.rttMu.Unlock()

	base := 50.0
	adaptiveMs := ewma + 3.0*jitter + base

	minMs := 150.0
	maxMs := 800.0 // Allow higher max for laggy networks

	if adaptiveMs < minMs {
		adaptiveMs = minMs
	}
	if adaptiveMs > maxMs {
		adaptiveMs = maxMs
	}

	randFactor := 1.0 + n.randSrc.Float64()*0.5
	adaptiveMs = adaptiveMs * randFactor

	// Note: Ticker is 20ms. So 200ms timeout = 10 ticks.
	newTick := int(math.Max(1, math.Floor(adaptiveMs/20.0))) 
	n.cfg.ElectionTick = newTick

	log.Printf("[Node %d] RTT EWMA=%.2fms jitter=%.2fms -> ElectionTimeout=%.0fms (%d ticks)",
		n.ID, ewma, jitter, adaptiveMs, newTick)

	n.logMetricsToCSV(ms, ewma, jitter, adaptiveMs, newTick)
}

func (n *RaftNode) UpdateElectionTimeout(rtt time.Duration) {
    n.updateRTTStatsAndTimeout(rtt)
}

func (n *RaftNode) Status() raft.Status {
	return n.raft.Status()
}

func (n *RaftNode) initCSV() {
	filename := fmt.Sprintf("node_%d_metrics.csv", n.ID)
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("[Node %d] Failed to create CSV: %v", n.ID, err)
		return
	}
	defer f.Close()
	f.WriteString("timestamp,node_id,rtt_raw,rtt_ewma,jitter,timeout_ms,ticks\n")
}

func (n *RaftNode) logMetricsToCSV(rawRTT, ewma, jitter, timeout float64, ticks int) {
	filename := fmt.Sprintf("node_%d_metrics.csv", n.ID)
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("[Node %d] Failed to write CSV: %v", n.ID, err)
		return
	}
	defer f.Close()

	line := fmt.Sprintf("%d,%d,%.2f,%.2f,%.2f,%.2f,%d\n",
		time.Now().UnixMilli(), n.ID, rawRTT, ewma, jitter, timeout, ticks)

	if _, err := f.WriteString(line); err != nil {
		log.Printf("[Node %d] CSV Write error: %v", n.ID, err)
	}
}