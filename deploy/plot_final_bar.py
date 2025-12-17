import pandas as pd
import matplotlib.pyplot as plt

# 1. Read the data from your automated trials
try:
    df = pd.read_csv("final_experiment_results.csv")
except FileNotFoundError:
    print("Error: final_experiment_results.csv not found. Run run_trials.py first.")
    exit()

# 2. Calculate Mean (MTR) and Standard Deviation (SD) for both modes
stats = df.groupby('mode')['recovery_ms'].agg(['mean', 'std'])
stats = stats.reset_index()

# Extract data for plotting
modes = stats['mode']
means = stats['mean']
stds = stats['std']

# 3. Setup the Bar Chart
plt.figure(figsize=(8, 6))
# Create bars with error bars (yerr)
bars = plt.bar(modes, means, yerr=stds, capsize=10, color=["#7403ff", "#00fff2"], edgecolor='black', alpha=0.9)

# 4. Styling and Labels
plt.title('Mean Time to Recovery (MTR): Fixed vs. Adaptive', fontsize=16, fontweight='bold')
plt.ylabel('Recovery Time (ms)', fontsize=14)
plt.xlabel('Raft Algorithm', fontsize=14)
plt.grid(axis='y', linestyle='--', alpha=0.6)

# 5. Add Mean Value Labels on top of error bars
for bar, mean_val, std_val in zip(bars, means, stds):
    # Calculate Y position: Top of the bar + Length of error bar + Buffer
    y_pos = mean_val + std_val + 50 
    
    label = f'{mean_val:.0f} ms'

    plt.text(bar.get_x() + bar.get_width()/2.0, y_pos,
             label, ha='center', va='bottom', fontsize=12, fontweight='bold')

# 6. Save and Display
output_file = 'final_result.png'
plt.savefig(output_file, dpi=300)
print(f"Graph saved as: {output_file}")