# rank_imdb.py

This script crunches IMDB data to generate lists of well-known movies, TV shows,
and people. It generates a text file listing several titles and names along
with ratings; this is suitable for use as a Phraser "prebaked" file.

To use it, grab some data exported from IMDB. This data is documented
at https://www.imdb.com/interfaces/ . You won't need all of the files listed
there, just:

+ <tt>name.basics.tsv.gz</tt> List of people with basic info.
+ <tt>title.basics.tsv.gz</tt> List of movies, shows, _etc_ with basic info.
+ <tt>title.principals.tsv.gz</tt> List of associations: titles ⬌ people most-associated with those titles
+ <tt>title.ratings.tsv.gz</tt> Ratings for titles: quality score and number of reviews.

Gunzip these files to uncompress them.

Run the <tt>rank_imdb.py</tt> script. It expects to find the <tt>*.tsv</tt>
files in the same directory. It generates two files in the same directory:

+ <tt>imdb_names.txt</tt> A crossword-dictionary text file of people-names.
    This file strives to rank crossword clue-able-ness of names. Along with
    entries for "Brad Pitt" it also has "Brad" and "Pitt". The theory is that
    Brad Pitt's fame makes it pretty easy to clue BRAD in a crossword, something
    like "Famous actor: ____ Pitt"
+ <tt>imdb_titles.txt</tt> A crossword-dictionary text file of titles.

To approximate well-known-ness, the script uses the number of reviews on IMDB.
As of a couple of days ago when I grabbed my data, the movie
Spider-Man: Into the Spider-Verse had
~500 thousand ratings; the TV show "Doom Patrol" had ~50 thousand;
thus this measure ranks Spider-Verse as more famous than Doom Patrol.
This works pretty well for finding movies beloved by IMDB's movie fans;
but it overlooks some other things. _E.g._, by this measure, the movie
"Wheel of Fortune and Fantasy" (2021) (which I never heard of),
ranks much _much_ higher than the long-running TV show "Wheel of Fortune."
Apparently, IMDB users don't tend to write reviews for TV game shows.

I [blogged some notes about the IMDB data](https://lahosken.san-francisco.ca.us/new/2022/08/25/crunching-imdb-data-imdb-internet-movie-database/).