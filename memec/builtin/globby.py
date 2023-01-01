#!/usr/bin/env python3

import glob
import json

gifs = glob.glob("*.gif")
pngs = glob.glob("*.png")
jpgs = glob.glob("*.j*g")

l = gifs + pngs + jpgs
l.sort()

outf = open("builtin.js", "w")
outf.write("const BUILTIN = " + json.dumps(l, indent=2))
outf.close()

