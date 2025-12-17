import pandas as pd
import matplotlib.pyplot as plt

# 1. Read the data
try:
    df = pd.read_csv("final_experiment_results_15nodes.csv")
except FileNotFoundError:
    print("Error: CSV file not found.")
    exit()

# 2. Calculate Means
means = df.groupby('mode')['recovery_ms'].mean()
# Reorder to keep consistency (Fixed left, Adaptive right) if needed
# But usually pandas sorts alphabetically. Let's force our order.
modes = ["Fixed (Baseline)", "Adaptive (Ours)"]
values = [means["Fixed (Baseline)"], means["Adaptive (Ours)"]]

# 3. Setup the Bar Chart
plt.figure(figsize=(8, 6))
bars = plt.bar(modes, values, color=['#ff9999', '#66b3ff'], width=0.5)

# 4. Styling
plt.title('Average Recovery Time: 15-Node Cluster', fontsize=16, fontweight='bold')
plt.ylabel('Mean Recovery Time (ms)', fontsize=12)
plt.grid(axis='y', linestyle='--', alpha=0.7)

# Add Labels on top
for bar in bars:
    height = bar.get_height()
    plt.text(bar.get_x() + bar.get_width()/2.0, height + 50,
             f'{height:.0f} ms', ha='center', va='bottom', fontsize=14, fontweight='bold')

# 5. Save
output_file = 'final_scalability_bar_chart.png'
plt.savefig(output_file, dpi=300)
print(f"🎉 Bar Chart saved as: {output_file}")