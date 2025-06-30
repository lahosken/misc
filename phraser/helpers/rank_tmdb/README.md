# rank_tmdb.py

This script crunches TMDB data to generate lists of well-known movies, TV shows,
and people. It generates a text file listing several titles and names along
with ratings; this is suitable for use as a Phraser "prebaked" file.

To use it, grab some data exported from TMDB. This data is documented
at https://developer.themoviedb.org/docs/daily-id-exports .
A helper script `howto_fetch.py` outputs command lines to download
yesterday's data (assuming you have the `wget` command). You might
run `howto_fetch.py` and then copy-paste those commands onto your command line.

Run the <tt>rank_tmdb.py</tt> script. It expects to find the <tt>*.json.gz</tt>
files in the same directory. It generates one file in the same directory:

+ <tt>tmdb.txt</tt> A text file of popular titles and names, with a popularity
  score for each.

TMDB supplies a popularity score, so we use that. Some of the scores seem sus
to me; perhaps unhinged fans have figured out how to "game" the system to drive
up this popularity score?  Using the score is easy, so I use it; but I'm
kinda glad this is just one signal among many.

TMDB keeps track of adult media. Adult media and stars _occasionally_ but rarely
show up in puzzles. Thus, I'd like to have these names show up in my
wordlists, but I don't want them ranking super-high. To make this happen, I put
an arbitrary cap of 15 on their scores when reading from TMDB's files.
