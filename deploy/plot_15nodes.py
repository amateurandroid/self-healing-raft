import pandas as pd
import matplotlib.pyplot as plt

# 1. Read the data
try:
    df = pd.read_csv("final_experiment_results_15nodes.csv")
except FileNotFoundError:
    print("Error: CSV file not found. Make sure you are in the 'deploy' folder.")
    exit()

# 2. Setup the Box Plot
plt.figure(figsize=(10, 6))
# Create boxplot grouped by 'mode' column
# We manually specify order to keep it consistent
df.boxplot(column='recovery_ms', by='mode', grid=True, patch_artist=True,
           boxprops=dict(facecolor='#a8dadc')) # Light blue color

# 3. Styling
plt.title('Scalability Test: 15-Node Cluster Recovery Time', fontsize=16, fontweight='bold')
plt.suptitle('') # Remove default pandas title
plt.xlabel('Algorithm Mode', fontsize=12)
plt.ylabel('Recovery Time (ms)', fontsize=12)

# Calculate medians for labels
medians = df.groupby('mode')['recovery_ms'].median()
for i, median in enumerate(medians):
    plt.text(i + 1, median + 100, f'{median:.0f} ms', 
             horizontalalignment='center', fontweight='bold', color='darkblue')

# 4. Save
output_file = 'final_scalability_15nodes.png'
plt.savefig(output_file, dpi=300)
print(f"Scalability Graph saved as: {output_file}")