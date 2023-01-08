#!/usr/bin/env python3

# regenerate the list of "built in" Meme image backgrounds

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

print("Content-Type: text/html")
print()

print("<tt> did something happen? </tt>")

