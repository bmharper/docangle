import matplotlib.pyplot as plt

example_numbers = [
    0.016,
    0.015,
    0.018,
    0.017,
    0.016,
    0.016,
    0.016,
    0.015,
    0.015,
    0.015,
]

# Plot histogram
plt.figure(figsize=(12, 6))
plt.hist(example_numbers, bins=50, edgecolor='black')
plt.title("Histogram of Provided Numbers")
plt.xlabel("Value")
plt.ylabel("Frequency")
plt.grid(True)
plt.show()
