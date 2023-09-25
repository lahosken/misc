#!/usr/bin/env python3

import sys
from difflib import ndiff
from collections import Counter

HEAD = """
<html>
  <head>
    <title>TODO</title>
    <meta charset="utf-8">
    <script type="text/javascript" src="puzzle.js"></script>
    <link rel="stylesheet" href="puzzle.css">
  </head>
<body>

<p>This is a "word ladder" to drill you learning the names of 
   TODO. By "word ladder", I mean that if a letter appears more
   than once in a line and in adjacent lines, typing that letter
   in one blank will make your appear in appropriate blanks in
   the line and adjacent lines.

<p> [<a href="{outpath}">quiz</a> | <a href="{solnpath}">solution</a> ]

<p><button onclick="resetAllPuzzleStateOnPage()">Reset everything</button></p>

<p>&nbsp;

"""

TAIL = """
</body>
</html>
"""

def load_list(inpath):
  rv = []
  for line in open(inpath):
      rv.append(line.strip().upper().replace(".", ""))
  return rv

def distance(s1, s2):
    def an(s):
        return "".join(sorted([c for c in s]))
    def dist(s1, s2):
        diffs = 0
        for l in ndiff(s1, s2):
            if l[0] != " ": diffs += 1000
        return int(diffs / (len(s1) + len(s2) + 1))
    return dist(s1, s2) + dist(an(s1), an(s2))

def all_distances(l):
    rv = []
    for s1 in l:
        for s2 in l:
            if s2 <= s1: continue
            rv.append((distance(s1, s2), s1, s2))
    rv.sort()
    return rv

def chainify(distances):
    already = Counter()
    color = {}
    links = []
    for _, s1, s2 in distances:
        color[s1] = s1
        color[s2] = s2
    for _, s1, s2 in distances:
        if already[s1] >= 2: continue
        if already[s2] >= 2: continue
        if color[s1] == color[s2]: continue
        already[s1] += 1
        already[s2] += 1
        old_color = color[s1]
        for s, v in color.items():
            if v == old_color: color[s] = color[s2]
        links.append((s1, s2))
    chain = []
    for s, v in already.items():
        if v == 1: chain = [s]
    while len(chain) < len(already):
        chain_end = chain[-1]
        for s1, s2 in links:
            if s1 == chain_end and s2 not in chain:
                chain.append(s2)
                break
            if s2 == chain_end and s1 not in chain:
                chain.append(s1)
                break
    return chain

def render(chain, outpath):
    chain = [""] + chain + [""]
    solnpath = outpath.replace(".html", "_soln.html")
    f = open(outpath, "w")
    f_soln = open(solnpath, "w")
    f.write(HEAD.format(outpath=outpath, solnpath=solnpath))
    f_soln.write(HEAD.format(outpath=outpath, solnpath=solnpath))

    count = 0
    prev_key = {}
    gimmes = 3
    for i in range(1, len(chain)-1):
        link = chain[i]
        data_text_l = []
        data_text_soln = ""
        data_extracts_l = []
        key = {}
        for c in link:
            if c == " ":                
                data_text_l.append("@")
                data_text_soln += "@"
                continue
            if c not in "ABCDEFGHIJKLMNOPQRSTUVWXYZ":
                data_text_l.append(c)
                data_text_soln += c
                continue
            if chain[i-1].count(c) + link.count(c) + chain[i+1].count(c) < 2:
                data_text_l.append(".")
                data_text_soln += c
                continue
            data_text_l.append("#")
            data_text_soln += c
            if c in prev_key:
                key[c] = prev_key[c]
            if c not in key:
                count += 1
                key[c] = count
            data_extracts_l.append(key[c])
        while data_text_l.count(".") >= data_text_l.count("#"):
            if not "." in data_text_l: break
            for di in range(len(data_text_l)):
                if data_text_l[di] == ".":
                    data_text_l[di] = link[di]
                    break
        while i in [1, len(chain)-2] and data_text_l.count(".") > 0:
            for di in range(len(data_text_l)):
                if data_text_l[di] == ".":
                    data_text_l[di] = link[di]
                    break
        while gimmes > 0 and data_text_l.count(".") > 0:
            gimmes -= 1
            for di in range(len(data_text_l)):
                if data_text_l[di] == ".":
                    data_text_l[di] = link[di]
                    break
        data_text = "".join(data_text_l)
        data_extracts = " ".join([str(x) for x in data_extracts_l])
        prev_key = key
        f.write(f"""
         <div class="puzzle-entry" data-mode="linear"
               data-text="{data_text}"
               data-extracts="{data_extracts}">
         </div>
        """)
        f_soln.write(f"""
         <div class="puzzle-entry" data-mode="linear solution"
               data-text="{data_text}"
               data-text-solution="{data_text_soln}">
         </div>
        """)
    f.close()

def main():
  if len(sys.argv) < 2:
    print("Usage: ./make-ladder.py text-file-with-one-time-per-line.txt")
    return
  inpath = sys.argv[-1]
  outpath = inpath.replace(".txt", "").replace(".", "_") + ".html"
  open(outpath, "a").close()
  in_list = load_list(inpath)
  distances = all_distances(in_list)
  chain = chainify(distances)
  render(chain, outpath)

main()
