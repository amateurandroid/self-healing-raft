import subprocess
import time
import requests
import pandas as pd
import os

# Configuration
TRIALS = 15
# UPDATED: Scan ports for all 15 nodes (8080 to 8094)
NODES = list(range(8080, 8095))
RESULTS_FILE = "final_experiment_results_15nodes.csv"

def get_leader():
    """Polls nodes to find the current leader ID."""
    for port in NODES:
        try:
            resp = requests.get(f"http://localhost:{port}/status", timeout=0.5)
            if resp.status_code == 200:
                for line in resp.text.split('\n'):
                    if "Leader" in line:
                        leader = int(line.split(':')[1].strip())
                        if leader != 0: return leader
        except:
            continue
    return None

def run_trial(mode, trial_id):
    print(f"\n🧪 Trial {trial_id} ({mode})...")
    
    # 1. Start Cluster with specific mode
    env = os.environ.copy()
    env["RAFT_MODE"] = mode
    # --scale nodeX is cleaner but manual definitions work better for diverse ports
    subprocess.run(["docker-compose", "up", "-d"], env=env, check=True)
    
    # 2. Wait for stable leader
    print("   Waiting for leader (may take longer with 15 nodes)...")
    leader = None
    for _ in range(60): # Increased wait time for large cluster
        leader = get_leader()
        if leader: break
        time.sleep(1)
        
    if not leader:
        print("Failed to elect leader. Skipping.")
        return None

    print(f"Leader is Node {leader}. Stabilizing...")
    time.sleep(10) # Longer stabilization for 15 nodes
    
    # 3. KILL LEADER
    print(f"Killing Node {leader}...")
    start_time = time.time()
    subprocess.run(["docker-compose", "stop", f"node{leader}"], check=True)
    
    # 4. Measure Recovery
    new_leader = None
    recovery_time = None
    
    for _ in range(400): # Poll for 20s max (50ms intervals)
        try:
            current_leader = get_leader()
            if current_leader and current_leader != leader:
                end_time = time.time()
                recovery_time = (end_time - start_time) * 1000 # ms
                print(f"Recovered! New Leader: {current_leader} in {recovery_time:.2f}ms")
                break
        except:
            pass
        time.sleep(0.05)
        
    # 5. Cleanup
    subprocess.run(["docker-compose", "down"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    
    return recovery_time

# --- MAIN EXECUTION ---
results = []

print("Starting Automated Experiments (15 Nodes)...")

# Run Baseline Trials
for i in range(1, TRIALS + 1):
    time_ms = run_trial("fixed", i)
    if time_ms: results.append({"mode": "Fixed (Baseline)", "trial": i, "recovery_ms": time_ms})

# Run Adaptive Trials
for i in range(1, TRIALS + 1):
    time_ms = run_trial("adaptive", i)
    if time_ms: results.append({"mode": "Adaptive (Ours)", "trial": i, "recovery_ms": time_ms})

# Save Results
df = pd.DataFrame(results)
df.to_csv(RESULTS_FILE, index=False)
print(f"\nExperiments Complete! Data saved to {RESULTS_FILE}")

# Quick Stat Summary
print("\nSummary Results:")
print(df.groupby("mode")["recovery_ms"].mean())