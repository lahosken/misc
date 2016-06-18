#!/usr/bin/env python
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
  win_file = gzip.open(win_filename, "w")
  in_file = gzip.open(in_filename)
  for line in in_file:
      try:
          ngram, year_s, count, vol_count = line.split("\t")
      except:
          continue
      if len(count) < 2: continue
      win_file.write(line)
  in_file.close()
  win_file.close()
  print win_filename
  
