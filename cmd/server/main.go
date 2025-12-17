package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/amateurandroid/self-healing-raft/pkg/raftnode"
	"github.com/amateurandroid/self-healing-raft/pkg/store"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: go run cmd/server/main.go <id> <port>")
	}

	id, _ := strconv.Atoi(os.Args[1])
	port := os.Args[2]

	selfAddr := fmt.Sprintf("localhost:%s", port)

	// UPDATED: 15-Node Configuration
	peers := map[uint64]string{
		1: "node1:8080", 2: "node2:8081", 3: "node3:8082",
		4: "node4:8083", 5: "node5:8084", 6: "node6:8085",
		7: "node7:8086", 8: "node8:8087", 9: "node9:8088",
		10: "node10:8089", 11: "node11:8090", 12: "node12:8091",
		13: "node13:8092", 14: "node14:8093", 15: "node15:8094",
	}

	kv := store.NewKVStore()
	rn := raftnode.NewRaftNode(id, selfAddr, peers, kv)
	rn.Start()

	http.HandleFunc("/raft", rn.HandleRaft)

	http.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
		k := r.URL.Query().Get("key")
		v := r.URL.Query().Get("val")

		if k == "" || v == "" {
			http.Error(w, "Usage: /put?key=foo&val=bar", http.StatusBadRequest)
			return
		}

		entry := []byte(fmt.Sprintf("%s=%s", k, v))
		rn.RaftPropose(entry)

		fmt.Fprintf(w, "Proposed to Raft: %s=%s\n", k, v)
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		k := r.URL.Query().Get("key")
		if v, ok := kv.Get(k); ok {
			fmt.Fprintf(w, "%s=%s\n", k, v)
		} else {
			http.Error(w, "Key not found", http.StatusNotFound)
		}
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status := rn.Status()
		fmt.Fprintf(w,
			"NodeID: %d\nLeader: %d\nTerm: %d\nCommitIndex: %d\nAppliedIndex: %d\n",
			status.ID, status.Lead, status.Term, status.Commit, status.Applied)
	})

	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	log.Printf("Node %d started on :%s", id, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}