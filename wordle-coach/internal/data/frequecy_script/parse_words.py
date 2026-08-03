from wordfreq import zipf_frequency
from pathlib import Path
DATA = Path(__file__).parent.parent


with open(DATA / "allowed.txt") as f:
    allowed_list = f.read().split()
with open(DATA / "answers.txt") as f:
    answers_list = f.read().split()

universe = answers_list + allowed_list

with open(DATA / "freq.tsv", "w", encoding="utf-8", newline="") as f:
    for word in sorted(universe):
        z = zipf_frequency(word, "en")
        f.write(f"{word}\t{z:.2f}\n")
    