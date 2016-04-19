import glob
import math
import os

TEST_PHRASES = [s.strip() for s in """
    1774 writing
    acetonitrile oxidation
    advance hails
    antibacterial action
    assisting mentally
    banned sex
    baptismal services
    better words
    cardinals seasons
    chilean spanish
    climbing capability
    covered with cheese
    digital recordings
    disavow any
    divisive political
    early electronics
    easily bind
    ellesmere manuscripts
    equipment shed
    estonian energy
    famous image
    first actions
    heard off screen
    industry bailout
    insoluable iron
    justin xxxv
    large flightless bird
    literally written
    local tape
    mostly farmers
    nationwide political
    nearest replacement
    nitroglycerin poisoning
    non planar
    notorious pirate captain
    only liquid
    personal agency
    plain moor
    poisonous darts
    political marriages
    recent difficulties
    reducing inflation
    sends telegrams
    sexual predation
    software subsystems
    some photographs
    south pembrokeshire
    special division
    speed capabilities
    starkly contrasted
    still formed
    sunny warm
    take a dive
    taylors hand
    territorial support
    the flimsiest
    theatrical counterparts
    transport process
    uneven singing
    winterer
    wise martin
""".splitlines() if s.strip()]

d = {}
for phrase in TEST_PHRASES: d[phrase] = {}

for filename in glob.glob("/home/lahosken/dumpz/wikitmp/20160413_205915/p-*.txt"):
  basename = os.path.basename(filename)
  f = open(filename)
  for line in f:
    score_s, phrase = line.strip().split("\t")
    if not phrase in d: continue
    score_i = int(score_s, 10)
    d[phrase][basename] = score_i
  f.close()

for phrase, found in d.items():
  mx = 0
  sm = 0.0
  for filename, count in found.items():
    if count > mx: mx = count
    sm = sm + math.log(1.0 + count)
  if mx and mx < 20 and sm > 5.0:
    print sm / mx, int(1000.0 * sm), mx, phrase
