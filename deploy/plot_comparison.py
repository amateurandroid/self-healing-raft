import matplotlib.pyplot as plt

# ----------------- DATA INPUT -----------------
# Replace these numbers with your exact observations if they differ.
# Based on your Docker logs:
# Baseline (Fixed): Took slightly longer to detect failure (approx 300-400ms)
# Adaptive (Smart): Tuned itself down to network speed (approx 200ms)
scenarios = ['Baseline (Fixed)', 'Adaptive (Ours)']
recovery_times = [350, 200] # values in milliseconds
colors = ['#ff9999', '#66b3ff']

# ----------------- PLOTTING -----------------
plt.figure(figsize=(8, 6))
bars = plt.bar(scenarios, recovery_times, color=colors, width=0.5)

# Add title and labels
plt.title('Mean Time to Recovery (MTR) Comparison', fontsize=16)
plt.ylabel('Recovery Time (ms)', fontsize=12)
plt.grid(axis='y', linestyle='--', alpha=0.7)

# Add value labels on top of bars
for bar in bars:
    height = bar.get_height()
    plt.text(bar.get_x() + bar.get_width()/2.0, height,
             f'{height} ms', ha='center', va='bottom', fontsize=12, fontweight='bold')

# Save
output_file = 'recovery_comparison.png'
plt.savefig(output_file, dpi=300)
print(f"🎉 Comparison graph saved as: {output_file}")
plt.show()