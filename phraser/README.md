# phraser

Counts up commonly-occuring phrases/words in sample text.

1. Setting up.

   **Get some sample text.**

   Get some text files. Put them in a dir.

    ```bash
    $ mkdir ~/dumpz
    $ mkdir ~/dumpz/textfiles
    $ cd ~/dumpz/textfiles
    $ wget http://www.gutenberg.org/cache/epub/100/pg100.txt
    ```

   Get some mediawiki export files. Put them in a dir.

    ```bash
    $ mkdir ~/dumpz
    $ mkdir ~/dumpz/wikimedia
    $ # (Read https://en.wikipedia.org/wiki/Wikipedia:Database_download to get recent wikipedia download url)
    $ cd ~dumpz/wikimedia
    $ wget http://something/something/enwiki-something-pages-articles-xml.bz2
    $ bunzip enwiki-something-pages-articles-xml.bz2
    ```

   Build phraser using [go](https://golang.org)

    ```bash
    go build github.com/lahosken/misc/phraser
    ```

2. Run. Noisy logs are noisy, sorry.

    ```bash
    $ ./phraser --txtpath dumpz/minitext/ --wikipath dumpz/miniwiki/
    2016/04/19 11:46:16 WRITING TO /tmp/phraser/20160419_114616
    2016/04/19 11:46:16 READING dumpz/minitext/pg100.txt
    2016/04/19 11:46:24  PERSIST /tmp/phraser/20160419_114616/p-pg100-000009793.txt
    2016/04/19 11:46:24    SORT...
    2016/04/19 11:46:28    BIG SORT DONE
    2016/04/19 11:46:43 WRITING TO /tmp/phraser/20160419_114616
    2016/04/19 11:46:43 READING dumpz/miniwiki/enwiki-20160204-pages-articles.xml
    2016/04/19 11:48:40  TAMP 1
    2016/04/19 11:50:28  TAMP 2
    2016/04/19 11:52:18  TAMP 3
    ...
    2016/04/19 12:12:47  TAMP 15
    2016/04/19 12:12:53  PERSIST /tmp/phraser/20160419_114616/p-enwiki-000043274.txt
    2016/04/19 12:12:53    SORT...
    2016/04/19 12:12:55    BIG SORT DONE
    2016/04/19 12:15:10  TAMP 1
    2016/04/19 12:17:03  TAMP 2
    2016/04/19 12:18:46  TAMP 3
      ...hours pass...
    2016/04/19 13:05:33 READING /tmp/phraser/20160419_114616/p-enwiki-000043274.txt
    2016/04/19 13:05:41 READING /tmp/phraser/20160419_114616/p-enwiki-000122214.txt
    2016/04/19 13:05:54 READING /tmp/phraser/20160419_114616/p-enwiki-000202308.txt
    2016/04/19 13:06:04 READING /tmp/phraser/20160419_114616/p-enwiki-000247955.txt
    2016/04/19 13:06:17 READING /tmp/phraser/20160419_114616/p-pg100-000009793.txt
    2016/04/19 13:06:23 READING /tmp/phraser/20160419_114616/p-enwiki-000043274.txt
    2016/04/19 13:06:34 READING /tmp/phraser/20160419_114616/p-enwiki-000122214.txt
    2016/04/19 13:06:39 READING /tmp/phraser/20160419_114616/p-enwiki-000202308.txt
    2016/04/19 13:06:51 READING /tmp/phraser/20160419_114616/p-enwiki-000247955.txt
    2016/04/19 13:07:16 READING /tmp/phraser/20160419_114616/p-pg100-000009793.txt
    2016/04/19 13:07:27  PERSIST /home/you/Phrases_20160419_114616.txt
    2016/04/19 13:07:27    SORT...
    2016/04/19 13:07:34    BIG SORT DONE
    $ 
    $ # WRITING TO points out tmp dir.
    $ # READING dumpz/minitext/pg100.txt reading an input file
    $ # PERSIST /tmp/phraser/.../p-....txt saving an intermediate count
    $ #   Usually it's one "PERSIST" per document, but wikipedia is so big,  it PERSISTs a few times
    $ # SORT... BIG SORT DONE sorting preparatory to saving
    $ # TAMP # a data structure got big, we're discarding some phrases
    $ # READING /tmp/phraser/.../p-....txt re-loading intermediate counts. Does it twice each!
    $ 
    ```

3. Gaze upon output. Each line of the file has format `score[TAB]phrase`.

    ```bash
    $ cd
    $ head -20 Phrases_20160419_114616.txt # What are most commonly-occuring phrases?
    575	the
    556	of
    544	and
    534	in
    529	a
    529	to
    498	is
    495	for
    494	of the
    491	as
    489	was
    489	with
    485	that
    483	by
    480	on
    476	in the
    472	it
    471	from
    470	his
    466	he
    $
    $ grep foo Phrases_20160419_114616.txt | head
    340	food
    339	football
    313	foot
    276	football league
    264	footballer
    262	food and
    261	american football
    259	fool
    254	national football
    250	fools
    $
    ```




