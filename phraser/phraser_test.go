package main

import (
//  "strings"
  "testing"
)

func TestTokenizeTypical(t *testing.T) {
  tokens := tokenize("Is Omotic Afro-Asiatic?")
  expected := []string{"is", "omotic", "afro", "asiatic"}
  if len(tokens) != len(expected) {
    t.Errorf("tokenize() returned strange value %v ", tokens)
    return
  }
  for ix, _ := range expected {
    if tokens[ix] != expected[ix] {
      t.Errorf("tokenize() returned strange value %v", tokens)
      return
    }
  }
}

func TestTokenizeApostrophe(t *testing.T) {
  tokens := tokenize("It's a small world, but I wouldn't want to paint it.")
  expected := []string{"its", "a", "small", "world", "but", "i", "wouldnt", "want", "to", "paint", "it"}
  if len(tokens) != len(expected) {
    t.Errorf("tokenize() returned strange value %v ", tokens)
    return
  }
  for ix, _ := range expected {
    if tokens[ix] != expected[ix] {
      t.Errorf("tokenize() returned strange value %v", tokens)
      return
    }
  }
}

func TestIngestWikiPageRedirect(t *testing.T) {
  co := counter{}
  // sample from Twilight Saga wikia. "Vampire Mythology" redirects to Vampire
  page := `
    <title>Vampire Mythology</title>
    <ns>0</ns>
    <id>1980</id>
    <redirect title="Vampire" />
      <sha1>6uw73ynta01am58li46gk6bbd5oistd</sha1>
    <revision>
      <id>165645</id>
      <timestamp>2010-08-23T04:53:55Z</timestamp>
      <contributor>
        <username>JoKalliauer</username>
        <id>2081935</id>
      </contributor>
      <comment>Vampire</comment>
      <text xml:space="preserve" bytes="21">#REDIRECT [[Vampire]]</text>
    </revision>`
  ingestWikiPage(page, &co)
  if co.d["vampire"] < 1 {
    t.Errorf("ingestWikiPage for redir page returned strange value %v", co)
    return
  }
  if co.d["vampire mythology"] < 1 {
    t.Errorf("ingestWikiPage for redir page returned strange value %v", co)
    return
  }
}

func TestIngestWikiPageFile(t *testing.T) {
  co := counter{}
  // sample from Gossip Girl wikia. This non-text-y .jpg shouldn't boost
  // phrase counts
  page := `
    <title>File:Dair-Wallpaper-dan-and-blair-1535983-1280-800.jpg</title>
    <ns>6</ns>
    <id>3866</id>
      <sha1>phoiac9h4m842xq45sp7s6u21eteeq1</sha1>
    <revision>
      <id>8414</id>
      <timestamp>2011-02-24T18:23:00Z</timestamp>
      <contributor>
        <username>SuperTash</username>
        <id>3252730</id>
      </contributor>
      <text xml:space="preserve" bytes="0" />
    </revision>`
  ingestWikiPage(page, &co)
  if len(co.d) > 0 {
    t.Errorf("ingestWikiPage for File: page returned strange value %v", co)
  }
}

func TestIngestWikiPagePisa(t *testing.T) {
  co := counter{}
  // thank you Futurama wikia for this sample
  page := `
    <title>Pisa</title>
    <ns>0</ns>
    <id>47</id>
      <sha1>1qlmxl1z6ssqmezx0zu2d5tk3tvhbl3</sha1>
    <revision>
      <id>66515</id>
      <timestamp>2014-10-28T14:04:57Z</timestamp>
      <contributor>
        <username>RRabbit42</username>
        <id>961279</id>
      </contributor>
      <comment>removing a category only used by this page</comment>
      <text xml:space="preserve" bytes="653">{{Location
|title = Pisa
|image = [[File:LeaningTowerOfPisa.png|250px]]
|planet = [[Earth]]
|town = 
|appearance = &quot;[[The Cryonic Woman]]&quot;
}}

'''Pisa''' is a city on [[Earth]], in the country of Italy. It was the former location of the [[Leaning Tower of Pisa]] until the 2600s when it was moved to a beach in [[New New York City]]. &lt;ref&gt;&quot;[[When Aliens Attack]]&quot;&lt;/ref&gt;

In [[3001]], Pisa was one of several cities [[Fry]] and [[Bender]] flew over during their joyride in the [[Planet Express ship]]. &lt;ref&gt;&quot;[[The Cryonic Woman]]&quot;&lt;/ref&gt;

== Appearances ==
* &quot;[[The Cryonic Woman]]&quot;

==Footnotes==
&lt;references/&gt;

[[Category:Cities]]
[[Category:Locations]]</text>
    </revision>`
  ingestWikiPage(page, &co)
  for _, expectedPresent := range []string{
    "pisa", "city on", "earth", "city on earth", "locations",
  } {
    if co.d[expectedPresent] < 1 {
      t.Errorf("ingestWikiPage for Pisa example expected %s but didn't see it", expectedPresent)
      return
    }
  }
  for _, expectedAbsent := range []string { // don't let these sneak in
    "quot", "category locations", "ref", "references",
  } {
    if co.d[expectedAbsent] != 0 {
      t.Errorf("ingestWikiPage for Pisa example should have ignored %s but alas counted it", expectedAbsent)
      return
    }
  }
  if co.d["pisa"] <= co.d["city on"] {
      t.Errorf(`ingestWikiPage for Pisa example should have scored "pisa" higher than "city on" but didn't.`)
      return    
  }
}
