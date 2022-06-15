#!/usr/bin/env python3

from nltk.corpus import wordnet
import json
from collections import defaultdict

def load_words():
    words = defaultdict(int)
    for line in open("/home/lahosken/words_500K.txt"):
        score_s, word = line.strip().split("\t")
        if len(word) != 5: continue
        score_i = int(score_s, 10)
        if score_i < 2000: break
        words[word] = score_i
    return words

def load_fives():
    words = {}
    for line in open("/home/lahosken/fives.txt"): words[line.strip()] = True
    return words

def load_collab():
    d = defaultdict(int)
    for line in open("/home/lahosken/the_game/ref/collabWL.txt"):
        W, score_s = line.strip().split(';')
        if len(W) != 5: continue
        if len(score_s) < 2: continue
        d[W.lower()] = int(score_s, 10)
    return d

def is_root(s):
    syns = wordnet.synsets(s)
    if not syns: return False
    for syn in syns:
        for lemma_name in syn.lemma_names():
            if lemma_name == s: return True
    return False

def main():
    many_words = load_words()
    collab = load_collab()
    fives = load_fives()
    sortable = []
    for word in many_words:
        score = (1 + many_words[word]) * (1 + collab[word])
        if word in fives: score *= 2
        sortable.append((score, word))
    sortable.sort()
    sortable.reverse()
    candidates = []
    probes = {}
    for _, word in sortable:
        if len(candidates) < 1800 and collab[word] > 30 and is_root(word):
            candidates.append(word)
        probes[word] = True
        if len(candidates) >= 1800 and len(probes) >= 5000:
            break
    f = open("list.js", "w")
    f.write("// This file is automatically generated. \n")
    f.write("CANDIDATES = " + json.dumps(candidates, indent=2))
    f.write(";\n\n")
    f.write("PROBES = " + json.dumps(probes, indent=2))
    f.write(";\n\n")
    f.close()
    
main()
