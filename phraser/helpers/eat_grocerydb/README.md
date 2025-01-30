eat_grocerydb.py
================

https://github.com/Barabasi-Lab/GroceryDB studied groceries
at Walmart, Target, and Whole Foods.  They published their
data set. They were mostly interested in nutrition, but they
there was some stuff useful for phraser.  E.g., the ingredients
list mentioned lots of processed foods.  We see their names
all the time, but they don't get a lot of Wikipedia cross-references
so `phraser` doesn't know they're important.  Why isn't
`modified corn starch` in the phrase list?  It's not an
_amazing_ puzzle answer, but it's better than `art and of`, right?

To use:

Download `GroceryDB_data_uncurated.csv` and `GroceryDB_foods.csv`
from  https://github.com/Barabasi-Lab/GroceryDB/tree/main/data

In the same dir:

$ `./eat_grocerydb.py > grocerydb.txt`

Copy `grocerydb.txt` to your "prebaked" dir