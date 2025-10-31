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

    // Start Raft node
    rn := raftnode.NewRaftNode(id, ":"+port)
    rn.Start()

    // Simple in-memory key-value store (local only for now)
    kv := store.NewKVStore()

    // PUT handler → send command to Raft AND update local store (temporary)
    http.HandleFunc("/put", func(w http.ResponseWriter, r *http.Request) {
        k := r.URL.Query().Get("key")
        v := r.URL.Query().Get("val")

        // 1. Store locally so GET works for now
        kv.Put(k, v)

        // 2. Propose to Raft log (replication will be added soon)
        entry := []byte(fmt.Sprintf("%s=%s", k, v))
        rn.RaftPropose(entry)

        fmt.Fprintf(w, "Proposed to Raft & stored locally: %s=%s\n", k, v)
    })

    // GET handler → read from local KV store
    http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
        k := r.URL.Query().Get("key")
        if v, ok := kv.Get(k); ok {
            fmt.Fprintf(w, "VAL: %s=%s\n", k, v)
        } else {
            http.Error(w, "Not found", http.StatusNotFound)
        }
    })

    log.Printf("Node %d running HTTP on :%s", id, port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}