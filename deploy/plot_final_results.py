import pandas as pd
import matplotlib.pyplot as plt

# 1. Read the data
df = pd.read_csv("final_experiment_results.csv")

# 2. Setup the Box Plot
plt.figure(figsize=(10, 6))
# Create boxplot grouped by 'mode' column
df.boxplot(column='recovery_ms', by='mode', grid=True, patch_artist=True,
           boxprops=dict(facecolor='lightblue'))

# 3. Styling
plt.title('Recovery Time Distribution: Baseline vs. Adaptive (15 Trials)', fontsize=14)
plt.suptitle('') # Get rid of the default pandas title
plt.xlabel('Algorithm Mode', fontsize=12)
plt.ylabel('Recovery Time (ms)', fontsize=12)

# 4. Save
output_file = 'final_boxplot.png'
plt.savefig(output_file, dpi=300)
print(f"Final Graph saved as: {output_file}")