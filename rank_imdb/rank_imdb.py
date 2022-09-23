#!/usr/bin/env python3

from collections import Counter
from math import floor

# we approximate notability with number of reviews.
# returns a counter a la { tt007: 5000, tt008: 10010 }
# which means the show with ID tt007 had 500 reviews and tt008 had 10010 reviews
# We don't consider whether the reviews are good/bad;
# "Manos: the Hands of Fate" is famous for being bad.
def read_title_ratings():
    r = Counter()
    for line in open("title.ratings.tsv"):
        tt, rating_s, numVotes_s = line.strip().split("\t")
        if tt == "tconst": continue
        numVotes = int(numVotes_s, 10)
        if numVotes < 5000: continue # too obscure
        r[tt] += numVotes
    return r

# returns a dict of { tt: title } (tho MUCH more info lurks in this file)
def read_title_basics(ratings=None):
    d = {}
    for line in open("title.basics.tsv"):
        tt, titleType, primaryTitle, originalTitle, isAdult,startYear, endYear, runtimeMinutes, genres = line.strip().split("\t")
        if ratings and not tt in ratings: continue
        d[tt] = primaryTitle
    return d

# returns a counter of { nm007: 1234, nm008: 56 }
# which means the person with ID nm007 is famous and nm008 is kinda famous
def read_title_principals(ratings):
    n = Counter()
    for line in open("title.principals.tsv"):
        tt, ordering, nm, category, job, characters = line.strip().split("\t")
        if tt not in ratings:
            n[nm] += 1
            continue
        if category in { "actor": 1, "actress": 1, "self": 1 }:
            n[nm] += 200 * ratings[tt]
        if category in { "director": 1 }: 
            n[nm] += 100 * ratings[tt]
        else:
            n[nm] += 100
    return n

# return a dict of { nm045: Bruce Lee, nm0553269: Margo Martindale } which
# maps IDs to names
# also returns a counter of { nm045: 123, nm0553269: 456 }
# which attempts to measure famousness of accomplishments based on
# contents of this file (not so useful compared to contens of principals file
# tho, so pay more attention to that)
def read_name_basics(people):
    nm_to_name = {}
    for line in open("name.basics.tsv"):
        nm, primaryName, birth, death, profs, titles = line.strip().split("\t")
        if nm not in people: continue
        nm_to_name[nm] = primaryName
    return nm_to_name

# Given a bunch of scores a la [1, 1, 1, 1, 2, 2, 2, 3, ..., 1234567890],
# return a list of ranges to map these scores to [1...100].
def percentiles(list_of_nums):
    d = {}
    for n in list_of_nums:
        d[n] = True
    u = list(d.keys())
    u.sort()
    return [0] + [u[floor(i * len(u) / 100)-1] for i in range(1, 101)]

# Given Counter and a list of ranges such as returned by percentiles(),
# return another Counter with same keys but values mapped to 1...100 range
def percentalize(raw, percs):
    r = Counter()
    for k, v in raw.items():
        for perc in range(1, 101):
            if v < percs[perc]: break
        r[k] = perc
    return r

def normalize(s):
    s = s.strip()
    retval = ""
    for c in s:
        c = c.lower()
        if c == "'": continue
        if c.isspace() or c == "-":
            if retval[-1] == " ": continue
            retval += " "
            continue
        if c.isalnum():
            retval += c
            continue
    return retval

def write_prebaked(title_basics, ratings, names, people):
    output = Counter()
    big_title_rating = ratings.most_common(10)[-1][1]
    scale = big_title_rating / 1000
    for tt, rating in ratings.most_common(3000):
        if rating > big_title_rating: rating = big_title_rating
        if tt not in title_basics: continue
        title = normalize(title_basics[tt])
        if len(title) < 1: continue
        if title in output: continue
        output[title] = int(rating / scale)
    big_people_rating = people.most_common(10)[-1][1]
    scale = big_people_rating / 1000
    for nm, rating in people.most_common(3000):
        if rating > big_people_rating: rating = big_people_rating
        if nm not in names: continue
        name = normalize(names[nm])
        if len(name) < 1: continue
        if name in output: continue
        output[name] = int(rating / scale)
    f = open("imdb.txt", "w")
    for key, value in output.most_common():
        f.write("{}\t{}\n".format(value, key))
    f.close()

def main():
    ratings = read_title_ratings()
    title_basics = read_title_basics(ratings)
    people = read_title_principals(ratings)
    names = read_name_basics(people)
    write_prebaked(title_basics, ratings, names, people)

main()
