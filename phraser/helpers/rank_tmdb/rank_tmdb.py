#!/usr/bin/env python

from collections import Counter
import glob
import gzip
import json

BIG_SCORE = 200
QUANT = 20_000

def normalize(s):
    s = s.strip()
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

def read_people():
    people = []
    people_fnames = glob.glob("*person_ids*.json.gz")
    for fname in people_fnames:
        for line in gzip.open(fname):
            o = json.loads(line.strip())
            popularity = o["popularity"]
            if popularity < 5: continue
            if "adult" in o and o["adult"] and popularity > 15: popularity = 15
            name = normalize(o["name"])
            if not len(name): continue
            people.append((popularity, name))
    people.sort()
    people.reverse()
    return people

def read_titles():
    titles = []
    movie_fnames = glob.glob("*movie_ids*.json.gz")
    for fname in movie_fnames:
        for line in gzip.open(fname):
            o = json.loads(line.strip())
            popularity = o["popularity"]
            if "adult" in o and o["adult"] and popularity > 15: popularity = 15
            if popularity < 5: continue
            title = normalize(o["original_title"])
            if not len(title): continue
            titles.append((popularity, title))
    tv_fnames = glob.glob("*tv_series_ids*.json.gz")
    for fname in tv_fnames:
        for line in gzip.open(fname):
            o = json.loads(line.strip())
            popularity = o["popularity"]
            if "adult" in o and o["adult"] and popularity > 15: popularity = 15
            if popularity < 5: continue
            title = normalize(o["original_name"])
            if not len(title): continue
            titles.append((popularity, title))
    titles.sort()
    titles.reverse()
    return titles

def scale_score(line_no):
    maybe = int(BIG_SCORE * (1 - (line_no / QUANT)))
    if maybe < 1: maybe = 1
    return maybe

def tally(pop_list, c):
    line_no = 0
    prev_popularity = -1
    prev_score = -1
    for popularity, name in pop_list:
        if name in c:
            c[name] += 1
        else:
            line_no += 1
            if popularity != prev_popularity:
                if line_no > QUANT: break
                prev_popularity = popularity
                prev_score = scale_score(line_no)        
            c[name] += prev_score

def write_report(c):
    outf = open("tmdb.txt", "w")        
    for name, score in c.most_common():
        outf.write(f"{score}\t{name}\n")
    outf.close()

def main():
    c = Counter()
    people = read_people()
    tally(people, c)
    titles = read_titles()
    tally(titles, c)
    write_report(c)

main()
