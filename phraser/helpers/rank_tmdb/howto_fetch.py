#!/usr/bin/env python

import datetime

yesterday = datetime.date.today() - datetime.timedelta(days=1)
MM_DD_YYYY = yesterday.strftime("%m_%d_%Y")

MTYPES = [
        "movie_ids",
        "adult_movie_ids",
        "tv_series_ids",
        "adult_tv_series_ids",
        "person_ids",
        "adult_person_ids",
        ]

for mtype in MTYPES:
    print(f"wget http://files.tmdb.org/p/exports/{mtype}_{MM_DD_YYYY}.json.gz")
