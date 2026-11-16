# Self-Healing Distributed System (Adaptive Raft)

A replicated key-value store in Go built on etcd's Raft library (`go.etcd.io/etcd/raft/v3`), with an experiment comparing a fixed election timeout against an adaptive one.

## What's mine and what's the library

- **etcd's Raft library** handles consensus: leader election and log replication.
- **My work:**
  - the node wrapper and key-value store around the library
  - the HTTP API
  - the adaptive election timeout
  - the Docker cluster setup
  - the fault-injection harness that measures recovery time

## Adaptive election timeout

A fixed election timeout has to be tuned for the worst network a cluster will ever see. Set too low, you get unnecessary elections. Set too high, a dead leader takes longer to replace.

The adaptive mode tracks observed network latency with an exponentially weighted moving average (α = 0.25) and sets the election timeout from it, clamped to 150–800 ms. The goal is to detect a failed leader faster when the network is healthy without triggering spurious elections when it slows down.

## API

Each node serves:

| Endpoint | Purpose |
|---|---|
| `GET /put?key=K&val=V` | Propose a write through Raft |
| `GET /get?key=K` | Read from the node's local store |
| `GET /status` | Node ID, current leader, term, commit index, applied index |
| `GET /ping`, `GET /probe` | Liveness checks |

## Running the cluster

```bash
cd deploy
docker-compose up -d                     # 15-node cluster (ports 8080–8094), adaptive timeout
RAFT_MODE=fixed docker-compose up -d     # same cluster with the fixed timeout
```

Each node reads `RAFT_MODE`. Any value other than `fixed` turns on the adaptive timeout.

## The experiment

`deploy/run_trials.py` runs the same test for each mode (`fixed`, then `adaptive`):

1. Start the cluster and wait for a leader.
2. Let it stabilize for 10 seconds.
3. Stop the leader's container (`docker-compose stop`).
4. Poll every node's `/status` every 50 ms until a new leader appears, and record the elapsed time.

It runs 15 trials per mode and writes the results to CSV. The `plot_*.py` scripts chart them.

## Limitations

- **Timing resolution:** the harness polls every 50 ms and checks nodes one after another, so measured recovery times include polling delay. That matters more with 15 nodes than with 3.
- **Network conditions:** the cluster runs on one machine's Docker network. The RTT probe adds a simulated 50–100 ms delay before each measurement so the adaptive logic sees WAN-like latency, but Raft's own messages travel without added delay.
- **Sample size:** 15 trials per mode is enough to see a trend but not to make strong statistical claims.

## What I'd do next

- Inject real network latency with `tc netem`, so Raft traffic is delayed too, not just the RTT probe
- Measure recovery from container logs or event timestamps, not polling
- Add a linearizability check on reads and writes during leader failover
