#!/usr/bin/python

import codecs
import glob
import os.path
import shutil

DEST = "srvtest"

def Copy(frompath):
  if not frompath: return
  topath = os.path.join(DEST, frompath)
  if frompath.startswith("server/") and frompath.endswith(".yaml"):
      topath = topath.replace("server/", "")
  print topath
  todir = os.path.dirname(topath)
  if not os.path.isdir(todir):
    os.makedirs(todir)
  shutil.copyfile(frompath, topath)

def Template2Go():
  infile = codecs.open("client/index.html", "r", "utf-8")
  template = infile.read()
  infile.close()
  tdotgo = codecs.open("server/templates.go", "w", "utf-8")
  tdotgo.write('''package server

// Don't edit this file. It's automatically generated!

var tmplS = `''')
  tdotgo.write(template)
  tdotgo.write("`\n")
  tdotgo.close()

def Main():
  olds = glob.glob(DEST+"/*")
  for old in olds:
    try:
      shutil.rmtree(old)
    except OSError:
      os.remove(old)
  Template2Go()
  for p in glob.glob("server/*.go") + glob.glob("server/*.yaml"):
      if p.endswith("_test.go"): continue
      Copy(p)
  for p in glob.glob("client/*"):
      Copy(p)

Main()
