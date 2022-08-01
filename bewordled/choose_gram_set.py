#!/usr/bin/env python3

import itertools
import math

UN = "e t a l n o p g m w".split(" ")

BI = "ng st le re in th al se ar en ch on an ea ce te nd co nt ma ne ri".split(" ")

TR = "sta ard pla ght her ent com art ice cal lea and ure rea the tur cha ain all one ear tin pri per ste".split(" ")

WORDS = {}
for line in open("word_list.txt"):
    score_s, word = line.strip().split("\t")
    score_i = int(score_s, 10)
    WORDS[word] = score_i

best_i = 0
best_s = ""

for nt in range(1, 4):
 for nb in range(nt+1, nt+3):
  for nl in range(nb+1, nb+3):
   for comb3 in itertools.combinations(TR, nt):
    bad2s = {}
    for c in comb3:
      bad2s[c[:2]] = True
      bad2s[c[-2:]] = True
    bi = [b for b in BI if b not in bad2s]
    for comb2 in itertools.combinations(bi, nb):
     overlaps = [(b, t) for b in comb2 for t in comb3 if t.startswith(b)]
     overlaps += [(b, t) for b in comb2 for t in comb3 if t.endswith(b)]
     if len(overlaps): continue
     bad1s = {}
     for c in comb3:
         bad1s[c[0]] = True
         bad1s[c[-1]] = True
     for c in comb2:
         bad1s[c[0]] = True
         bad1s[c[-1]] = True
     un = [l for l in UN if l not in bad1s]
     for comb1 in itertools.combinations(un, nl):
      overlaps += [(u, b) for u in comb1 for b in comb2 if b.endswith(u)]
      overlaps += [(u, b) for u in comb1 for b in comb2 if b.startswith(u)]
      overlaps += [(u, t) for u in comb1 for t in comb3 if t.endswith(u)]
      overlaps += [(u, t) for u in comb1 for t in comb3 if t.startswith(u)]
      if len(overlaps):
          continue
      comb = comb1 + comb2 + comb3
      i = 0
      acc_l = []
      bummerWord = False
      for prod in itertools.product(comb, repeat=3):
          word = "".join(prod)
          if word in ['anal', 'arse', 'doo', 'heroin', 'pee', 'shit', 'stalin', 'tard', 'teat', 'tit']:
              bummerWord = True
              break
          if word in WORDS:
              acc_l.append(word)
              i += 10 + int(math.log(WORDS[word])) + len(word)
      if bummerWord: continue
      acc_l.sort()
      acc = " ".join(acc_l)
      c3ok = True
      for c3 in comb3:
          contains = [a for a in acc_l if c3 in a]
          if len(contains) < 2: c3ok = False
      if not c3ok: continue
      i = int((i * 1000) / math.pow(nb + nl + nt, 3))
      i *= int((i * 1000) / math.sqrt(nb + nl + nt))
      if i >= 9 * best_i / 10:
          print("{} : {} / {}".format(i, acc, " ".join(comb)))
      if i > best_i:
          best_i = i
