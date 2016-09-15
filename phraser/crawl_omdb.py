#!/usr/bin/env python

import codecs
import json
import sys
import time
import urllib
import urllib2

if len(sys.argv) < 2:
    print "Usage: python crawl_omdb.py path/to/previously/generated/phrase_file.txt"
    exit

phrases = open(sys.argv[1])
line_count = 0
error_count = 0
outf = codecs.open("omdb_0.txt", "w", encoding="utf-8")
for line in phrases:
  line_count += 1
  time.sleep(0.1)
  _, phrase = line.strip().split("\t")
  print line_count, phrase
  query = urllib.urlencode({'s': phrase})
  try:
    response = urllib2.urlopen("http://www.omdbapi.com/?" + query)
    error_count = 0
  except:
    error_count += 1
    if error_count > 3:
        break
    time.sleep(1.0)
    continue
  try:
    j = json.load(response)
    if j[u'Response'] == u'True':
      score = int(j[u'totalResults'], 10) + 1
      for i in j[u'Search']:
          outf.write(u'{}\t{}\n'.format(score, i[u'Title']))
          score -= 1
  except:
    error_count += 1
    if error_count > 3:
        break
    time.sleep(1.0)
    # continue

  if not (line_count % 10000):
      outf.close()
      outf = codecs.open("omdb_{}.txt".format(line_count), "w", encoding="utf-8")
outf.close()
