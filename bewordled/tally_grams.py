#!/usr/bin/env python3

from collections import Counter

words = []
for line in open("/home/lahosken/words_500K.txt"):
    score_s, word = line.strip().split("\t")
    if len(word) > 9: continue
    words.append(word)

c = Counter()    

for n1 in [1, 2, 3]:
 for n2 in [1, 2, 3]:
  for n3 in [1, 2, 3]:
      n = n1 + n2 + n3
      start1 = 0
      end1 = n1
      start2 = n1
      end2 = n1 + n2
      start3 = n1 + n2
      end3 = n1 + n2 + n3
      count = 0
      for word in words:
          count += 1
          if len(word) != n: continue
          g1 = word[start1:end1]
          g2 = word[start2:end2]
          g3 = word[start3:end3]
          c[g1] += 1
          c[g2] += 1
          c[g3] += 1
          if count > 1500: break

f = open("1_grams.txt", "w")
for k, v in c.most_common():
    if len(k) != 1: continue
    f.write("{}\t{}\n".format(k, v))
f = open("2_grams.txt", "w")
for k, v in c.most_common():
    if len(k) != 2: continue
    f.write("{}\t{}\n".format(k, v))
f = open("3_grams.txt", "w")
for k, v in c.most_common():
    if len(k) != 3: continue
    f.write("{}\t{}\n".format(k, v))
