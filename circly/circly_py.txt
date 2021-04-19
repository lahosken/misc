#! /usr/bin/env python3

from PIL import Image, ImageDraw
import math
import random
import sys

COLOR_THRESHHOLD = 1000

USAGE = """
Usage: ./circly.py original.jpg

Generates distortions of an image; where those distortions
use a "circly" effect, described at 
https://lahosken.san-francisco.ca.us/frivolity/prog/circly/ .

Generates out-00.png, out-01.png, out-02.png, ...out-etc.png.
Each of these overlays successively smaller circles atop the
previous.
"""

def main():
    if len(sys.argv) < 2:
        print(USAGE)
        return
    orig = Image.open(sys.argv[1])
    w, h = orig.size
    out = Image.new(mode="RGB", size=(w, h), color=(0, 0, 0))
    draw = ImageDraw.Draw(out)
    r = math.sqrt(w * h) + 1
    count = 0
    while r > 0.5 and count < 100:
        r = round(0.707*r, 2)
        cA = math.pi * r * r
        print("{:02d} r:{:.2f}".format(count, r))
        x_offset = random.random() * w
        y_offset = random.random() * h
        x_i = 0
        y_i = 0
        while x_i < w:
            x_i += 4*r
            while y_i < h:
                y_i += 4*r 
                x = int(x_offset + x_i) % w
                y = int(y_offset + y_i) % h
                r0, g0, b0 = orig.getpixel((x, y))
                rd, gd, bd = out.getpixel((x, y))
                color_dist_2 = (r0-rd)**2 + (g0-gd)**2 + (b0-bd)**2
                if color_dist_2 < COLOR_THRESHHOLD: continue
                draw.ellipse((x-r, y-r, x+r, y+r), fill=(r0, g0, b0))
            y_i -= h
        filename = "out-{:02d}.png".format(count)
        out.save(filename)
        count += 1
    return

main()
