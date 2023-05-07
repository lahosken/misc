#!/usr/bin/env python3

import csv

outf = open("cryptics-george-ho.txt", "w")

for row in csv.reader(open("clues.csv")):
    clue = row[1]
    answ = row[2]
    defn = row[3]
    outf.write("{}\n\n{}\n\n{}\n\n".format(clue, answ, defn))

outf.close()
