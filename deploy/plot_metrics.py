import pandas as pd
import matplotlib.pyplot as plt
import glob
import os

# 1. Setup the plot style
plt.figure(figsize=(12, 6))
plt.title('Adaptive Election Timeout Over Time', fontsize=16)
plt.xlabel('Time (Seconds from start)', fontsize=12)
plt.ylabel('Election Timeout (ms)', fontsize=12)
plt.grid(True, linestyle='--', alpha=0.6)

# 2. Find all node_*.csv files in the current directory
csv_files = sorted(glob.glob("node_*_metrics.csv"))

if not csv_files:
    print("No CSV files found! Make sure node_1_metrics.csv etc. are in this folder.")
    exit()

print(f"found {len(csv_files)} CSV files: {csv_files}")

# 3. Loop through each file and plot data
for file in csv_files:
    try:
        # Read the CSV
        df = pd.read_csv(file)
        
        # Strip whitespace from column names just in case
        df.columns = df.columns.str.strip()
        
        # Check if data exists
        if df.empty:
            print(f"⚠️ Skipped {file} (empty)")
            continue

        # Normalize Time: Start at 0 seconds
        start_time = df['timestamp'].min()
        df['seconds'] = (df['timestamp'] - start_time) / 1000.0
        
        # Extract Node ID for the legend
        node_id = df['node_id'].iloc[0]
        
        # Plot Time vs Timeout
        plt.plot(df['seconds'], df['timeout_ms'], marker='o', markersize=4, linestyle='-', label=f'Node {node_id}')
        
        print(f"Plotted data from {file} ({len(df)} rows)")

    except Exception as e:
        print(f"Error reading {file}: {e}")

# 4. Finalize and Save
plt.legend()
plt.tight_layout()

output_file = 'adaptive_timeout_graph.png'
plt.savefig(output_file, dpi=300)
print(f"\n🎉 Graph saved as: {output_file}")
plt.show() # Attempt to show the plot window