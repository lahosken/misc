#!/usr/bin/env python3

import sys
import itertools

comb = sys.argv[1:]

WORDS = {}
for line in open("word_list.txt"):
    score_s, word = line.strip().split("\t")
    score_i = int(score_s, 10)
    WORDS[word] = score_i

count = {}

for prod in itertools.product(comb, repeat=3):
    word = "".join(prod)
    if not word in WORDS: continue
    if not prod[0] in count: count[prod[0]] = 0
    if not prod[1] in count: count[prod[1]] = 0
    if not prod[2] in count: count[prod[2]] = 0
    count[prod[0]] += 1
    count[prod[1]] += 1
    count[prod[2]] += 1

low = len(WORDS)
for c in count:
    if count[c] < low: low = count[c]

acc = []
for c in count:
    acc += int(2 * count[c] / low) * [c]

print(acc)
