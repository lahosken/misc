#!/usr/bin/env python3

import re
import time
import urllib.request
import xml.dom.minidom

SITES = {
    "haj-math-meme": "https://lemmy.blahaj.zone/feeds/c/mathmemes.xml?sort=Active",
    "leml-prog-hum": "https://lemmy.ml/feeds/c/programmerhumor.xml?sort=Active",
    "leml-memes": "https://lemmy.ml/feeds/c/memes.xml?sort=Active",
    }

LINK_RE = re.compile(r'<a href="([^"]*)"')
IMG_SUFFIXES = [".jpg", ".png", ".webp", ".jpeg"]

def fetch_one(site):
    req = urllib.request.urlopen(SITES[site])
    tree = xml.dom.minidom.parseString(req.read())

    for item in tree.getElementsByTagName("description"):
        imgs = []
        link_matches = LINK_RE.findall(item.firstChild.wholeText)
        for lm in link_matches:
            if not len([s for s in IMG_SUFFIXES if s in lm]): continue
            imgs.append(lm)
        if not len(imgs):
            continue
        prepend = ""
        for img in imgs:
            prepend += f'<img src="{img}">\n'
            prepend += "\n"
        item.firstChild.replaceWholeText(prepend + item.firstChild.wholeText)
    f = open(site + ".xml", "w")
    f.write(tree.toprettyxml())
    f.close()

def main():
    for site in SITES:
        try:
            fetch_one(site)
        except:
            pass
        time.sleep(1.0)

main()
