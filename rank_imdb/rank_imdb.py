#!/usr/bin/env python3

from collections import Counter
from math import floor

# we approximate notability with number of reviews.
# returns a counter a la { tt007: 500, tt008: 1001 }
# which means the show with ID tt007 had 500 reviews and tt008 had 1001 reviews
# We don't consider whether the reviews are good/bad;
# "Manos: the Hands of Fate" is famous for being bad.
def read_title_ratings():
    r = Counter()
    for line in open("title.ratings.tsv"):
        tt, rating_s, numVotes_s = line.strip().split("\t")
        if len(numVotes_s) < 3: continue # less than 100 votes? too obscure
        if tt == "tconst": continue
        numVotes = int(numVotes_s, 10)
        r[tt] += numVotes
    return r

# returns a dict of { tt: title } (tho MUCH more info lurks in this file)
def read_title_basics():
    d = {}
    for line in open("title.basics.tsv"):
        tt, titleType, primaryTitle, originalTitle, isAdult,startYear, endYear, runtimeMinutes, genres = line.strip().split("\t")
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
def read_name_basics(ratings):
    n = Counter()
    nm_to_name = {}
    for line in open("name.basics.tsv"):
        nm, primaryName, birth, death, profs, titles = line.strip().split("\t")
        if nm == "nconst": continue
        nm_to_name[nm] = primaryName
        if len(birth) > 3: n[nm] += 5 # someone cared enough to note birth year
        for tt in titles.split(","):
            if tt in ratings and "act" in profs:
                n[nm] += ratings[tt]
            else:
                n[nm] += 1
    return (nm_to_name, n)

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

# Write the imdb_titles.txt file
def write_titles(title_basics, ratings):
    already = {}
    f = open("imdb_titles.txt", "w")
    for tt, rating in ratings.most_common():
        if tt not in title_basics: continue
        fancy_title = title_basics[tt]
        title = "".join([c for c in fancy_title if c.isalnum() or c == " "])
        title_nospc = title.replace(" ", "").lower()
        if title_nospc in already: continue
        already[title_nospc] = True
        f.write("{};{}\n".format(title, rating))
    f.close()

# Write the imdb_names.txt file
def write_names(names, people):
    already = {}
    f = open("imdb_names.txt", "w")
    for nm, rating in people.most_common():
        if nm not in names: continue
        fancy_name = names[nm]
        name = "".join([c for c in fancy_name if c.isalnum() or c == " "])
        name_nospc = name.replace(" ", "").lower()
        if len(name_nospc) < 3: continue
        if name_nospc in already: continue
        already[name_nospc] = True
        f.write("{};{}\n".format(name, rating))
        if name.count(" ") != 1: continue
        # Julia Roberts is famous. Thus, "Julia Roberts" and
        # "Julia" and "Roberts" are all nicely clue-able.
        # So let's make entries for first and last name:
        first, last = name.split(" ")
        first_nospc = first.replace(" ", "").lower()
        last_nospc = last.replace(" ", "").lower()
        if len(first_nospc) >= 3 and first_nospc not in already:
            already[first_nospc] = True
            f.write("{};{}\n".format(first, rating))
        if len(last_nospc) >= 3 and last_nospc not in already:
            already[last_nospc] = True
            f.write("{};{}\n".format(last, rating))
    f.close()
    
def main():
    ratings_raw = read_title_ratings()
    ratings_50k = Counter()
    for k, v in ratings_raw.most_common(50000):
        ratings_50k[k] = v
    rating_percentiles = percentiles(ratings_50k.values())
    ratings = percentalize(ratings_50k, rating_percentiles)
    title_basics = read_title_basics()
    write_titles(title_basics, ratings)
    people_raw = Counter()
    people_raw += read_title_principals(ratings_raw)
    names, raw = read_name_basics(ratings_raw)
    people_raw += raw
    people_50k = Counter()
    for k, v in people_raw.most_common(50000):
        people_50k[k] = v
    people_raw = 0
    people_percentiles = percentiles(people_50k.values())
    people = percentalize(people_50k, people_percentiles)
    write_names(names, people)

main()
