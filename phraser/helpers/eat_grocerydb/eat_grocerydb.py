#!/usr/bin/env python

import re
import csv
from collections import Counter

INGS_SPLIT = re.compile("[,()]")
ENTITY = re.compile("&#[0-9]*;")

class G:
    pass

def normalize(s):
    s = s.strip()
    s = s.replace("&#38;", "and") # html-entity-ish ampersands
    s = s.replace("&#39;", "") # html-entity-ish apostrophe
    s = ENTITY.sub(" ", s) # html-entity symbols like TM; discard i guess
    retval = ""
    for c in s:
        c = c.lower()
        if c == "'": continue
        if c == "&": retval += "and"
        if c.isspace() or c == "-":
            if not len(retval): continue
            if retval[-1] == " ": continue
            retval += " "
            continue
        if c.isalnum():
            retval += c
            continue
    return retval.strip()

def find_frags(s):
    rv = []
    all = s.split(" ")
    for start_ix in range(len(all)):
        for end_ix in range(start_ix+1, len(all)+1):
            frag = " ".join(all[start_ix:end_ix])
            if len(frag) > 50: continue
            rv.append(frag)
    return rv

def read_groc_uncurated():
    rv = []
    f = open("GroceryDB_data_uncurated.csv")
    rdr = csv.DictReader(f)
    for row in rdr:
        o = G()
        o.name = row["Name"]
        o.ings = row["Ingredients"]
        rv.append(o)
    return rv

def read_groc_foods():
    rv = []
    f = open("GroceryDB_foods.csv")
    rdr = csv.DictReader(f)
    for row in rdr:
        o = G()
        o.name = row["name"]
        o.brand = row["brand"]
        rv.append(o)
    return rv

def main():
    c = Counter()
    groc_u = read_groc_uncurated()
    for g in groc_u:
        name = normalize(g.name)
        if not len(name): continue
        c[name] += 10
        frags = find_frags(name)
        for frag in frags:
            c[frag] += 1
        ings_raw = INGS_SPLIT.split(g.ings)
        for ing_raw in ings_raw:
            ing = normalize(ing_raw)
            if not len(ing): continue
            c[ing] += 1
    groc_f = read_groc_foods()
    for g in groc_f:
        brand = normalize(g.brand)
        if not len(brand): continue
        c[brand] += 10
    for k,v in c.most_common():
        if v < 20: break
        v = int(v/2)
        if v > 50: v = int(v / 2) + 25
        if v > 100: v = 100
        print(f"{v}\t{k}")

main()
