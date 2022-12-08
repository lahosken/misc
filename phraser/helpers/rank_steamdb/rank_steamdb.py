#!/usr/bin/env python3

# Uses big blob of json steamdb.min.json.gz from
# https://github.com/leinstay/steamdb
# Outputs steamdb_prebaked.txt for your prebaked dir and
#         steamdb_text.txt for your txtpath dir.
# Prebaked has titles, scored by popularity
# Text has descriptions

from collections import Counter
import json
import math
import re

EL_RE = re.compile(r"\<[^>]*\>")

def normalize(s):
    tr = ""
    for c in s:
        if c == "&": tr += " and "
        if c == "'": continue
        if c.isalnum():
            tr += c.lower()
            continue
        tr += " "
    tr = tr.strip()
    while "  " in tr:
        tr = tr.replace("  ", " ")
    return tr

s = open("steamdb.min.json").read()
j = json.loads(s)

desc_f = open("steamdb_text.txt", "w")
best = Counter()
    
for r in j:
    stsp_owners = r["stsp_owners"] or 0
    igdb_popularity = r["igdb_popularity"] or 0.0
    score = int(math.log(stsp_owners + igdb_popularity + 1))
    name = normalize(str(r["name"]))
    desc = EL_RE.sub("", r["description"] or "")
    desc = desc.replace("&reg;", "")
    desc = desc.replace("&trade;", "")
    desc = desc.replace("&quot;", '"')
    devpub = "{}/{}".format(r["developers"], r["publishers"])
    if r["developers"] == r["publishers"]:
        devpub = r["developers"]
    if score > 5 and len(name) > 0:
        if score > best[name]: best[name] = score
        desc_f.write("\n{}\n\n{}\n\n{}\n".format(r["name"], devpub, desc))

prebaked_f = open("steamdb_prebaked.txt", "w")
for k, v in best.most_common():
    prebaked_f.write("{}\t{}\n".format(v, k))
