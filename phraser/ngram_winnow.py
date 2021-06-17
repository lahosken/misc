#!/usr/bin/env python3
# In googlebooks-eng-all-5gram-* , >90% of each file carefully records
# rare ngrams with just 1-2 appearances.  We're interested in common
# ngrams, so discard rare ones.

import glob
import gzip
import os

in_filenames = glob.glob("*.gz")
in_filenames.sort()
for in_filename in in_filenames:
  if "winnowed" in in_filename: continue
  win_filename = in_filename.replace(".gz", "-winnowed.gz")
  if os.path.exists(win_filename): continue
  win_file = gzip.open(win_filename, "wt")
  in_file = gzip.open(in_filename, "rt")
  print("READING " + in_filename)
  for line in in_file:
    if "_" in line: continue # Don't want "Bakersfield_NOUN", so skip it. Just use "Bakersfield"
    phrase, in_data_s = line.strip().split("\t", 1)
    out_data = []
    for in_data in in_data_s.split("\t"):
      year_s, count_s, vols_s = in_data.split(",")
      if int(count_s, 10) < 20: continue
      if int(vols_s, 10) < 3: continue
      out_data.append(in_data)
    if len(out_data) < 3: continue
    win_file.write(phrase + "\t" + "\t".join(out_data) + "\n")
  in_file.close()
  win_file.close()
  print("FINISHED " + win_filename)
  
