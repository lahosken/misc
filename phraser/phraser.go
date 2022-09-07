// App that looks for good puzzle-answer words and phrases.
package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  $ %s --txtpath dumpz/textfiles --wikipath dumpz/mediawiki --ngrampath dumpz/ngram \n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Then go get some lunch. Later, ./Phrases_*.txt has tab-separated freq,phrase info.\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n  More info at http://github.com/lahosken/misc/phraser\n\n")
		os.Exit(0)
	}
}

var wikiFodderPath = flag.String("wikipath", "", "Dir full of wiki-export .xml files")
var txtFodderPath = flag.String("txtpath", "", "Dir full of .txt files")
var ngramFodderPath = flag.String("ngrampath", "", "Dir full of Google Ngram files")
var prebakedFodderPath = flag.String("prebakedpath", "", "Dir full of previously-generated Phraser lists")
var xwdFodderPath = flag.String("xwdpath", "", "Dir full of crossword scored word-list files")
var tmpPath = flag.String("tmppath", "", "Don't parse wikis/textfiles. Instead, read from previously-generated tmp dir")
var outPath = flag.String("outpath", "", "Instead of writing to ~/Phrases_20160419_131415.txt, write to this path")
var cpuprofile = flag.String("cpuprofile", "", "Write cpu profile to file")

const (
	tmpFilenameFormat         = "p-%s-%012d.txt"
	dictTampThreshholdEntries = 5000000
	dictOutputThreshhold      = 8000000
	lowScore                  = 5 // TODO instead of magic constant, could compute histogram in persist()
)

var (
	// find start of the "body" of wiki entry in XML
	textRE = regexp.MustCompile(`<text[^>]*>(.*)$`)
	// Title of wiki entry in XML.
	titleRE = regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	// redirRE: some wiki entries are just pointers to others e.g.
	// <page>
	//   <title>Octothorpe</title>
	//   <redirect title="Number sign" />
	//   ...
	// </page>
	redirRE = regexp.MustCompile(`<redirect title="([^"]+)" />`)
	// find {{things}} in wiki markup
	cur2RE = regexp.MustCompile(`\{\{([^\{\}]+)\}\}`)
	// find text "spans" that have been set off somehow (italicized,
	// parenthecized...). E.g., in
	//    The '''S. S. ''Minnow''''' is a fictional charter boat on the
	//    hit 1960s television sitcom ''Gilligan's Island''.
	// "Minnow" and "Gilligan's Island" are italicized; "S. S. Minnow" is
	// bolded. These spans often(?) indicate especially interesting text
	// sections.  E.g., in addition to tallying the text "The S S Minnow is a..."
	// we also should tally "S S Minnow", "Minnow".
	spanREs = []*regexp.Regexp{
		regexp.MustCompile(`''''(.*?)''''`),
		regexp.MustCompile(`'''(.*?)'''`),
		regexp.MustCompile(`''(.*?)''`),
		regexp.MustCompile(`&quot;(.*?)&quot;`),
		regexp.MustCompile(`&blockquote;(.*?)&blockquote;`),
		regexp.MustCompile(`‘(.*?)’`), // single "smart" quotes
	}
	// wiki entries are full of stuff to delete
	deleteMeREs = []*regexp.Regexp{
		regexp.MustCompile(`&lt;!--.*?--&gt;`),
		regexp.MustCompile(`<!--.*?-->`),
		regexp.MustCompile(`#redirect`),
		regexp.MustCompile(`&lt;math.*?&lt;math&gt;`), // not great, these can nest?
		regexp.MustCompile(`\[.*?\]`),
		regexp.MustCompile(`&lt;ref.*`),
		regexp.MustCompile(`.*&lt;/ref&gt;`),
		regexp.MustCompile(`&lt;math.*`),
		regexp.MustCompile(`.*&lt;/math&gt;`),
		regexp.MustCompile(`&quot;`),
		regexp.MustCompile(`&lt;sup.*?&lt;/sup&gt;`),
		regexp.MustCompile(`&lt;big.*?&lt;/big&gt;`),
		regexp.MustCompile(`&lt;.*?&gt;`),
	}
	// delete these, replacing with whitespace
	spaceREs = []*regexp.Regexp{
		// could these all be strings instead of regexps?
		regexp.MustCompile(`&amp;nbsp;`),
		regexp.MustCompile(`&amp;ndash;`),
		regexp.MustCompile(`&lt;br&gt;`),
		regexp.MustCompile(`&ndash;`),
	}
	// find [[things]] in wiki markup
	bra2RE = regexp.MustCompile(`\[\[([^\[\]]+)\]\]`)
	// find partial [[things]] in wiki markup that were broken
	// by line breaks.
	bra2FragREs = []*regexp.Regexp{
		regexp.MustCompile(`.*\]\]`),
		regexp.MustCompile(`\[\[.*`),
	}
	// We delete these. Doesn't go with the deleteMeREs since we do
	// this earlier on.
	refRE = regexp.MustCompile(`&lt;ref.*?&lt;/ref&gt;`)
	// If an Ngram starts/ends with one of these stopwords, don't count it. "of the west", e.g., is
	// not a super-great phrase.
	ngramBadStarts = map[string]bool{"of": true, "and": true, "or": true, ",": true}
	ngramBadEnds   = map[string]bool{
		"of": true, "and": true, "but": true, "or": true, "a": true, "an": true, "the": true, "if": true,
		"when": true, "than": true, "because": true, "while": true, "where": true, "unless": true,
		"except": true, "so": true, "as": true, "to": true, "very": true, ",": true,
	}
)

// A counter is a string -> int key value store.
//
// Useful for keeping track of freqencies with which snippets appear
// in a sample; or a "score" for normalized strings; or...
type counter struct {
	d map[string]uint64
}

// boost boosts score for a string by N.
func (c *counter) boost(s string, n uint64) {
	if c.d == nil {
		c.d = map[string]uint64{}
	}
	c.d[s] += n
}

// inc boosts score for a string by 1.
func (c *counter) inc(s string) {
	c.boost(s, 1)
}

// tamp deletes some low-scored entries.
//
// If your counter has grown to consume most memory,
// you might tamp it down with this.
func (c *counter) tamp() {
	for s, score := range c.d {
		if score <= uint64(strings.Count(s, " ")) {
			delete(c.d, s)
		}
	}
}

// persist writes out contents to a file.
func (c counter) persist(outPath string) {
	log.Printf(" PERSIST %v", outPath)
	if len(c.d) > 5000000 {
		log.Printf("   SORT...")
	}
	sortMe := map[uint64]([]string){}
  bigScores := []string{}
	maxScore := uint64(0)
	for phrase, score := range c.d {
		if score > maxScore {
			maxScore = score
		}
		if sortMe[score] == nil {
			sortMe[score] = []string{}
      if score > 1000000 {
        bigScores = append(bigScores, fmt.Sprintf("% 20d", score))
      }
		}
		sortMe[score] = append(sortMe[score], phrase)
	}
	if len(c.d) > 5000000 {
		log.Printf("   BIG SORT DONE SIZE=%v", len(c.d))
	}
	os.MkdirAll(filepath.Dir(outPath), 0776)
	outF, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("couldn't open outfile %v %v", outPath, err)
	}
	writtenCount := 0
	lastSync := 0
  sort.Sort(sort.Reverse(sort.StringSlice(bigScores)))
  for _, scoreS := range bigScores {
    score, err := strconv.ParseUint(strings.TrimSpace(scoreS), 10, 64)
    if err != nil {
      log.Fatalf("couldn't parse a score %s %v", scoreS, err)
    }
		sort.Sort(sort.StringSlice(sortMe[score]))
		for _, phrase := range sortMe[score] {
			outF.WriteString(fmt.Sprintf("%d\t%s\n", score, phrase))
			writtenCount += 1
		}
		delete(sortMe, score)
		if writtenCount > dictOutputThreshhold {
			break
		}
		if writtenCount > lastSync+50000 {
			outF.Sync()
			lastSync = writtenCount
		}
  }
  if maxScore > 1000000 {
    maxScore = 1000000
  }
	for score := maxScore; score > 0; score -= 1 {
		if sortMe[score] == nil {
			continue
		}
		sort.Sort(sort.StringSlice(sortMe[score]))
		for _, phrase := range sortMe[score] {
			outF.WriteString(fmt.Sprintf("%d\t%s\n", score, phrase))
			writtenCount += 1
		}
		delete(sortMe, score)
		if writtenCount > dictOutputThreshhold {
			break
		}
		if writtenCount > lastSync+50000 {
			outF.Sync()
			lastSync = writtenCount
		}
	}
	outF.Close()
}

// tokenize takes "Is Omotic Afro-Asiatic?", returns ["is", "omotic", "afro", "asiatic"]
//
// Does some normalization and splits string into list of word-like things.
func tokenize(snippet string) (tokens []string) {
	alphanum := func(r rune) rune {
		switch {
		case unicode.IsDigit(r), unicode.IsLetter(r):
			return unicode.ToLower(r)
		case r == '\'':
			return -1
		default:
			return ' '
		}
	}
	snippet = strings.Map(alphanum, snippet)
	for _, token := range strings.Fields(snippet) {
		if len(token) > 0 {
			tokens = append(tokens, token)
		}
	}
	return
}

// Given mediawiki blob, ingestWikiPage boosts score of found phrases.
func ingestWikiPage(page string, co *counter) {
	page = strings.ToLower(page)
	textModeP := false // are we in the page's <text> element?
	found := new(counter)
	title := ""
	for _, line := range strings.Split(page, "\n") {
		if !textModeP {
			if redirREMatch := redirRE.FindStringSubmatch(line); redirREMatch != nil {
				line2snippetsPastBrackets(title, found)
				line2snippetsPastBrackets(title, found)
				line2snippetsPastBrackets(redirREMatch[1], found)
				continue
			}
			if titleREMatch := titleRE.FindStringSubmatch(line); titleREMatch != nil {
				// colon signals discardiness: User:, File:, Talk:, Etc:...
				if strings.Contains(titleREMatch[1], ":") {
					return
				}
				if strings.HasSuffix(titleREMatch[1], "/Gallery") {
					return
				}
				title = titleREMatch[1]
				line2snippetsPastBrackets(title, found)
				line2snippetsPastBrackets(title, found)
				continue
			}
			if textREMatch := textRE.FindStringSubmatch(line); textREMatch != nil {
				line = textREMatch[1]
				textModeP = true
				// Fall through.
			}
		}
		if textModeP {
			if strings.Contains(line, "</text>") {
				line = strings.Replace(line, "</text>", "", 11)
				textModeP = false
			}
			if strings.Contains(line, "&lt;gallery") {
				// this "article" is prrrobably mostly a list of .jpgs. abort!
				break
			}
			line2snippets(line, found)
		}
	}
	tallySnippets(co, *found)
}

// helper function for line2snippets.
func line2snippetsPastBrackets(line string, snippets *counter) {
	for _, deleteMeRE := range deleteMeREs {
		line = deleteMeRE.ReplaceAllString(line, "")
	}
	if cur2REMatches := cur2RE.FindAllStringSubmatch(line, -1); cur2REMatches != nil {
		for _, cur2REMatch := range cur2REMatches {
			pipeFields := strings.Split(cur2REMatch[1], "|")
			if len(pipeFields) == 2 {
				switch pipeFields[0] {
				case "main", "further", "further2", "see also", "commons category", "portal":
					line2snippetsPastBrackets(pipeFields[1], snippets)

				case "nowrap", "small", "smaller", "quote":
					line2snippetsPastBrackets(pipeFields[1], snippets)
					lless := strings.Replace(line, cur2REMatch[0], pipeFields[1], 1)
					line2snippetsPastBrackets(lless, snippets)
					return

				default:
					line = strings.Replace(line, cur2REMatch[0], "", 1)
				}
			} else {
				line = strings.Replace(line, cur2REMatch[0], "", 1)
			}
		}
	}
	line = strings.Split(line, "{{")[0]
	if strings.Contains(line, "}}") {
		line = strings.Split(line, "}}")[0]
	}
	// We're past [[...]] and {{...}} . Any of this crap remaining means we're
	// prrrobably in something icky, e.g. line-broken fragment of a [[...]]
	for _, symptom := range []string{"|", "file:", "image:"} {
		if strings.Contains(line, symptom) {
			return
		}
	}
	line = strings.Replace(line, "#redirect ", "", -1)
	for _, spaceRE := range spaceREs {
		line = spaceRE.ReplaceAllString(line, " ")
	}
	line = strings.Replace(line, "&amp;", "&", -1)
	for _, spanRE := range spanREs {
		if matches := spanRE.FindAllStringSubmatch(line, -1); matches != nil {
			for _, match := range matches {
				line2snippetsPastBrackets(match[1], snippets)
				line = strings.Replace(line, match[0], match[1], 1)
			}
		}
	}
	// apostrophe variants that shouldn't litter "shouldn t" in our results
	line = strings.Replace(line, "&#039;", "", -1)
	line = strings.Replace(line, "’", "", -1)
	snippets.inc(line)
	return
}

// Given "Electric Boogaloo", boost counts for "electric", "boogaloo", and "electric boogaloo"
func tallySnippets(tally *counter, found counter) {
	for snippet, score := range found.d {
		if score < 1 {
			continue
		}
		tokens := tokenize(snippet)
		for startIx, _ := range tokens {
			for endIx := startIx + 1; endIx <= len(tokens); endIx += 1 {
				key := strings.Join(tokens[startIx:endIx], " ")
				tally.boost(key, score)
				if len(key) > 35 {
					break
				}
			}
		}
	}
}

// line2snippets handles one line of wikitext, fills counter with snippets.
func line2snippets(line string, snippets *counter) {
	// bra2RE: double square-brackets around non-square brackets [[Foo]]
	line = strings.TrimLeft(line, " *:")
	if strings.HasPrefix(line, "|") || strings.HasPrefix(line, "!") || strings.HasPrefix(line, "{|") {
		return
	}
	line = refRE.ReplaceAllString(line, "")
	if bra2REMatches := bra2RE.FindAllStringSubmatch(line, -1); bra2REMatches != nil {
		for _, bra2REMatch := range bra2REMatches {
			pipeFields := strings.Split(bra2REMatch[1], "|")
			if strings.HasPrefix(pipeFields[0], "category:") || strings.HasPrefix(pipeFields[0], ":category:") {
				line2snippetsPastBrackets(pipeFields[0][9:], snippets)
				return // [[Category:...]] tends to be on its own line
			}
			// file:, image:, talk: are hard to get right. Meh, discard 'em.
			if strings.Contains(pipeFields[0], ":") {
				line = strings.Replace(line, bra2REMatch[0], "", 1)
				continue
			}
			if len(pipeFields) == 2 {
				// for input "in 1901, [[Mahatma Gandhi|Gandhi]] stopped in Mauritius"
				// we want to count strings
				//   Mahatma Gandhi
				//   Gandhi
				//   in 1901, Gandhi stopped in Maritius
				line2snippetsPastBrackets(pipeFields[0], snippets)
				// count "Gandhi" double; it's how folks really refer to it
				line2snippetsPastBrackets(pipeFields[1], snippets)
				line2snippetsPastBrackets(pipeFields[1], snippets)
				// continue, getting the anchor text in context
				line = strings.Replace(line, bra2REMatch[0], pipeFields[1], 1)
			} else if len(pipeFields) == 1 {
				// for input "prime minister [[Indira Gandhi]] of India" we want to
				// count strings
				//   Indira Gandhi
				//   prime minister Indira Gandhi of India
				// count the link triple
				line2snippetsPastBrackets(pipeFields[0], snippets)
				line2snippetsPastBrackets(pipeFields[0], snippets)
				line2snippetsPastBrackets(pipeFields[0], snippets)
				// continuing, getting anchor text in context
				line = strings.Replace(line, bra2REMatch[0], pipeFields[0], 1)
			} else {
				// discard other [[obscure wiki markup]]
				line = strings.Replace(line, bra2REMatch[0], "", 1)
			}
		}
	}
	for _, bra2FragRE := range bra2FragREs {
		line = bra2FragRE.ReplaceAllString(line, "")
	}
	line2snippetsPastBrackets(line, snippets)
}

// readNgrams reads data from Google N-gram files, write it to tmp files.
//
// Ngrams: https://storage.googleapis.com/books/ngrams/books/datasetsv3.html
//
// I've only ever used this on Ngram files that I'd previously winnowed
// down (removing low-frequency entries). Dunno how/if it would work on
// raw files; I wasn't patient enough to try. Also, I wasn't willing to
// devote most of my hard drive to keeping those files around for the attempt.
func readNgrams(fodderPath, tmpPath string) {
	lineCount := uint64(0)
	inFilePaths, _ := filepath.Glob(filepath.Join(fodderPath, "*-winnowed.gz"))
	found := counter{}
	for _, inFilePath := range inFilePaths {
		shortName := strings.Split(filepath.Base(inFilePath), ".")[0]
		if strings.Contains(shortName, "-_") {
			// we ignore _ADJ_, _NOUN_, etc. ngrams, which these "_" files are full of, so...
			continue
		}
		is1gramP := strings.HasPrefix(shortName, "1-")
		log.Printf("READING %v", inFilePath)
		fodderF, err := os.Open(inFilePath)
		if err != nil {
			log.Printf(" OPEN ERR %v", err)
			continue
		}
    defer fodderF.Close()
		gUnzipper, err := gzip.NewReader(fodderF)
		if err != nil {
			log.Fatalf("GZip reader has a sad: %v", err)
		}
		fodderScan := bufio.NewScanner(gUnzipper)
		for {
			fileNotDone := fodderScan.Scan()
			if !fileNotDone {
				break
			}
			line := fodderScan.Text()
			lineCount += 1
			line = strings.TrimSpace(line)
			ngram_and_counts := strings.Split(line, "\t")
			if len(ngram_and_counts) < 2 {
				continue
			}
			ngram := ngram_and_counts[0]
			if strings.Contains(ngram, "_") { // avoid grammar-labeled _NOUN_ etc
				continue
			}
			if strings.Contains(ngram, "'") {
				// I can't figure out apostrophes are encoded. I see "one ' s" and
				// "lion 's" and... I should remove spaces, but I'm not sure _which_.
				// I give up, skip these.
				continue
			}
			for i := 1; i < len(ngram_and_counts); i++ {
				fields := strings.Split(ngram_and_counts[i], ",")
				year, err := strconv.Atoi(fields[0])
				if err != nil {
					continue
				}
				years_ago := time.Now().Year() - year
				match_count, err := strconv.Atoi(fields[1])
				if err != nil {
					continue
				}
				volume_count, err := strconv.Atoi(fields[2])
				if err != nil {
					continue
				}
				if volume_count > years_ago {
					// capital letters are a (weak) indicator of quality.
					// "The United States of America" is better puzzle-fodder than
					// "and it can be expressed".
					frags := strings.Split(ngram, ".")
					for _, frag := range frags {
						if (!is1gramP) && (len(tokenize(frag)) < 2) {
							continue
						}
						capitalCount := 0.0
						for _, r := range frag {
							if unicode.IsUpper(r) {
								capitalCount += 1.0
							}
						}
						score := (3.0 + capitalCount) * math.Sqrt(math.Sqrt(float64(match_count))+float64(volume_count)) / float64(years_ago+1)
						if score >= 1.0 {
							found.boost(frag, uint64(score))
						}
					}
				}
			}
		}
		gUnzipper.Close()
		fodderF.Close()
		if len(found.d) > dictTampThreshholdEntries {
			for snippet, score := range found.d {
				words := strings.Split(snippet, " ")
				if ngramBadStarts[words[0]] || ngramBadEnds[words[len(words)-1]] {
					delete(found.d, snippet)
					if score > 0 {
						for len(words) > 1 && ngramBadStarts[words[0]] {
							words = words[1:]
						}
						for len(words) > 1 && ngramBadEnds[words[len(words)-1]] {
							words = words[:len(words)-1]
						}
						if len(words) > 1 {
							subSnippet := strings.Join(words, " ")
							found.boost(subSnippet, score)
						}
					}
				}
			}
			tally := counter{}
			tallySnippets(&tally, found)
			tallyPath := filepath.Join(tmpPath, fmt.Sprintf(tmpFilenameFormat, shortName, lineCount))
			tally.persist(tallyPath)
			found = counter{}
		}
	}
	tally := counter{}
	tallySnippets(&tally, found)
	tallyPath := filepath.Join(tmpPath, fmt.Sprintf(tmpFilenameFormat, "googlebooks-eng-all-DONE-winnowed", lineCount))
	tally.persist(tallyPath)
}

// readTextFiles reads data from text files, write it to tmp files.
func readTextFiles(fodderPath, tmpPath string) {
	paraCount := 0
	inFilePaths, _ := filepath.Glob(filepath.Join(fodderPath, "*.txt"))
	for _, inFilePath := range inFilePaths {
		shortName := strings.Split(filepath.Base(inFilePath), ".")[0]
		log.Printf("READING %v", inFilePath)
		fodderF, err := os.Open(inFilePath)
		if err != nil {
			log.Fatalf("couldn't open txt file %s %v", inFilePath, err)
		}
    defer fodderF.Close()
		found := counter{}
		fodderScan := bufio.NewScanner(fodderF)
		para := ""
		for {
			fileNotDone := fodderScan.Scan()
			if !fileNotDone {
				break
			}
			line := fodderScan.Text()
			if strings.Contains(line, "START OF THIS PROJECT GUTENBERG") {
				// Apparently, we've been carefully tallying up the Project Gutenberg
				// file header. That's not so useful; jettison it.
				found.d = map[string]uint64{}
				continue
			}
			if strings.Contains(line, "END OF THIS PROJECT GUTENBERG") {
				// Let's not tally the Project Gutenberg file footer.
				break
			}
			// some fortune files' sections are separated by % lines
			line = strings.Replace(line, "%", "", -1)
			// some text files' sections are separated by * lines
			line = strings.Replace(line, "*", " ", -1)
			if strings.Contains(line, "[PAGE") {
				line = ""
			}
			line = strings.TrimSpace(line)
			if line == "" || len(para) > 5000 {
				para = para + " " + line
				found.inc(para)
				if strings.Contains(para, ".") {
					for _, sentence := range strings.Split(para, ".") {
						sentence = strings.TrimSpace(sentence)
						if strings.Contains(sentence, " ") {
							found.inc(sentence)
						}
					}
				}
				para = ""
				paraCount += 1
			}
			line = strings.ToLower(line)
			para = para + " " + line
		}
		fodderF.Close()
		found.inc(para)
		boost := uint64((10000 + len(found.d)) / (1000 + len(found.d)))
		if boost > 1 {
			for key, _ := range found.d {
				found.boost(key, boost)
			}
		}
		tally := counter{}
		tallySnippets(&tally, found)
		tallyPath := filepath.Join(tmpPath, fmt.Sprintf(tmpFilenameFormat, shortName, paraCount))
		tally.persist(tallyPath)
	}
}

func readXWdLists(fodderPath, tmpPath string) {
	inFilePaths, _ := filepath.Glob(filepath.Join(fodderPath, "*t"))
	for _, inFilePath := range inFilePaths {
		tally := counter{}
		shortName := strings.Split(filepath.Base(inFilePath), ".")[0]
		log.Printf("READING %v", inFilePath)
		fodderF, err := os.Open(inFilePath)
		if err != nil {
			log.Fatalf("couldn't open txt file %s %v", inFilePath, err)
		}
    defer fodderF.Close()
		found := counter{}
		fodderScan := bufio.NewScanner(fodderF)
		for {
			fileNotDone := fodderScan.Scan()
			if !fileNotDone {
				break
			}
			line := fodderScan.Text()
			line = strings.TrimSpace(line)
			if !strings.Contains(line, ";") {
				continue
			}
			phraseAndScore := strings.Split(line, ";")
			phrase := phraseAndScore[0]
			phrase = strings.ToLower(phrase)
			phrase = strings.TrimSpace(phrase)
			score, err := strconv.Atoi(phraseAndScore[1])
			if err != nil {
				continue
			}
			if len(phrase) <= 0 {
				continue
			}
			if score > 100 {
				score = 100
			}
			// Sometimes these lists leave out spaces. E.g.,
			// you'll see "THEPINKPANTHER" instead of "the pink panther".
			// Alas, we don't have a great way to detect this and
			// could thus think there's a word "thepinkpanther".
			// So fake it: if we see a long "word" with no spaces,
			// don't boost its score so much:
			if len(phrase) > 5 && !strings.Contains(phrase, " ") {
				score /= len(phrase)
			}
			if score < 10 {
				continue
			}
			found.boost(phrase, uint64(score*score*score*score*2000))
		}
		tallySnippets(&tally, found)
		tallyPath := filepath.Join(tmpPath, fmt.Sprintf(tmpFilenameFormat, "XwdLs-"+shortName, 0))
		tally.persist(tallyPath)
	}
}

func readPrebaked(fodderPath, tmpPath string) {
	inFilePaths, _ := filepath.Glob(filepath.Join(fodderPath, "*.txt"))
  fileCount := 0
	for _, inPath := range inFilePaths {
    fileCount += 1
		nickName := strings.Split(filepath.Base(inPath), ".")[0]
		outPath := filepath.Join(tmpPath, fmt.Sprintf(tmpFilenameFormat, nickName, fileCount))
    inF, err := os.Open(inPath)
		if err != nil {
			log.Fatalf("couldn't open prebaked file %s %v", inPath, err)
		}
    defer inF.Close()
    os.MkdirAll(filepath.Dir(outPath), 0776)
    outF, err := os.Create(outPath)
    if err != nil {
      log.Fatalf("couldn't open tmp file %v %v", outPath, err)
    }
    defer outF.Close()
    scan := bufio.NewScanner(inF)
    lineRE := regexp.MustCompile(`(\d+)\t(.*)`)
    for {
			if fileNotDone := scan.Scan(); !fileNotDone {
				break
			}
			line := scan.Text()
      match := lineRE.FindStringSubmatch(line)
			if match == nil {
				log.Fatal("weird prebaked file line %s %s", inPath, line)
			}
      sqrtScore, err := strconv.Atoi(match[1])
			if err != nil {
				log.Fatal("weird prebaked file line (non-int score?) %s %s", inPath, line)
			}
      score := uint64(math.Pow(float64(sqrtScore), 2))
      phrase := match[2]
      outF.WriteString(fmt.Sprintf("%d\t%s\n", score, phrase))
    }
  }
}

// readWikis reads data from wikis, write it to tmp files.
func readWikis(fodderPath, tmpPath string) {
	pageCount := 0
	inFilePaths, _ := filepath.Glob(filepath.Join(fodderPath, "*.xml"))
	for _, inFilePath := range inFilePaths {
		wikiName := strings.Split(filepath.Base(inFilePath), "_")[0]
		wikiName = strings.Split(wikiName, "-")[0]
		log.Printf("READING %v", inFilePath)
		fodderF, err := os.Open(inFilePath)
		if err != nil {
			log.Fatalf("couldn't open wiki file %s %v", inFilePath, err)
		}
    defer fodderF.Close()
		var co *counter = new(counter)
		tampCount := 0
		fodderScan := bufio.NewScanner(fodderF)
		page := ""
		for {
			fileNotDone := fodderScan.Scan()
			if !fileNotDone {
				break
			}
			line := fodderScan.Text()
			if strings.HasPrefix(line, "  <page>") {
				pageCount += 1
				page = ""
				continue
			}
			if strings.HasPrefix(line, "  </page>") {
				ingestWikiPage(page, co)
				if tampCount < 8 && len(co.d) > dictTampThreshholdEntries {
					tampCount += 1
					log.Printf(" TAMP %d", tampCount)
					co.tamp()
				}
				if len(co.d) > dictTampThreshholdEntries {
					// if STILL full, rotate out...
					tallyPath := filepath.Join(tmpPath, fmt.Sprintf(tmpFilenameFormat, wikiName, pageCount))
					co.persist(tallyPath)
					co = new(counter)
					tampCount = 0
				}
				continue
			}
			if page == "" {
				page = line
			} else {
				page = page + "\n" + line
			}
		}
		fodderF.Close()
		tallyPath := filepath.Join(tmpPath, fmt.Sprintf(tmpFilenameFormat, wikiName, pageCount))
		co.persist(tallyPath)
	}
}

// reduce combines the scores we wrote out to temp files
func Reduce(tmpPath string, outPath string) {
	tmpFileGlobPattern := regexp.MustCompile(`%.*?d`).ReplaceAllString(tmpFilenameFormat, "*")
	tmpFilenames, err := filepath.Glob(filepath.Join(tmpPath, tmpFileGlobPattern))
	if err != nil {
		log.Fatal("couldn't glob temp files %v", err)
	}
	bigCounter := counter{}
	lineRE := regexp.MustCompile(`(\d+)\t(.*)`)

	// We make two passes.
	//
	// If we try to pick up ALL the entries, we fill up our memory with
	// uncommon phrases. Instead, two passes
	//   first: count everything w/count >15
	//   second: for items w/ count <=15, "boost" things we saw in 1st pass
	//
	// No, this isn't super-rigorous; it leaves out some things. But those
	// things mmmostly wouldn't appear in the first 1M phrases, so good enough
	// for our purposes.

	magicNumber := 15

	// First pass
	for _, tmpFilename := range tmpFilenames {
		log.Printf("READING %v", tmpFilename)
		tmpF, err := os.Open(tmpFilename)
		if err != nil {
			log.Fatalf("couldn't open tmp file %s %v", tmpFilename, err)
		}
    defer tmpF.Close()
		scan := bufio.NewScanner(tmpF)
		for {
			fileNotDone := scan.Scan()
			if !fileNotDone {
				break
			}
			line := scan.Text()
			match := lineRE.FindStringSubmatch(line)
			if match == nil {
				log.Fatal("Weird tmp file line %s %s", tmpFilename, line)
			}
			score, err := strconv.Atoi(match[1])
			if err != nil {
				log.Fatal("Weird tmp file line (non-int score?) %s %s", tmpFilename, line)
			}
			if score <= magicNumber {
				break
			}
			phrase := match[2]
			bigCounter.boost(phrase, uint64(math.Sqrt(float64(score))))
		}
		tmpF.Close()
	}
	// Second pass
	for _, tmpFilename := range tmpFilenames {
		log.Printf("READING %v", tmpFilename)
		tmpF, err := os.Open(tmpFilename)
		if err != nil {
			log.Fatalf("couldn't open tmp file %s %v", tmpFilename, err)
		}
    defer tmpF.Close()
		scan := bufio.NewScanner(tmpF)
		for {
			if fileNotDone := scan.Scan(); !fileNotDone {
				break
			}
			line := scan.Text()
			match := lineRE.FindStringSubmatch(line)
			if match == nil {
				log.Fatal("weird tmp file line %s %s", tmpFilename, line)
			}
			score, err := strconv.Atoi(match[1])
			if err != nil {
				log.Fatal("weird tmp file line (non-int score?) %s %s", tmpFilename, line)
			}
			if score > magicNumber {
				continue
			}
			phrase := match[2]
			// only "boost" existing, no new stuff
			if _, contains := bigCounter.d[phrase]; contains {
				bigCounter.boost(phrase, uint64(math.Sqrt(float64(score))))
			}
		}
		tmpF.Close()
	}
	if len(bigCounter.d) > dictOutputThreshhold {
		// Probably our bigCounter is chock-full of stuff and the machine is straining
		// under the load. Delete low-score entries so that we don't waste time
		// sorting them and such.
		log.Printf(" LEN(bigCounter.d)=%v BEFORE lowScore PURGE", len(bigCounter.d))
		for phrase, score := range bigCounter.d {
			if score <= lowScore {
				delete(bigCounter.d, phrase)
			}
		}
		log.Printf(" LEN(bigCounter.d)=%v AFTER lowScore PURGE", len(bigCounter.d))
	}
	bigCounter.persist(outPath)
}

func main() {
	flag.Parse()
	if len(flag.Args()) > 0 && flag.Args()[0] == "help" {
		flag.Usage()
	}
	nowPath := time.Now().Format("20060102_150405")
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if *outPath == "" {
		u, _ := user.Current()
		*outPath = filepath.Join(
			u.HomeDir,
			fmt.Sprintf("Phrases_%s.txt", nowPath))
	}
	if *tmpPath == "" && *wikiFodderPath == "" && *ngramFodderPath == "" && *txtFodderPath == "" && *prebakedFodderPath == "" && *xwdFodderPath == "" {
		log.Fatal("none of --wikipath, --ngrampath, --prebakedpath --txtpath, --xwdpath set. No input, nothing to do!")
	}
	if *tmpPath == "" {
		*tmpPath = filepath.Join(os.TempDir(), "phraser", nowPath)
		if *ngramFodderPath != "" {
			readNgrams(*ngramFodderPath, *tmpPath) // read ngrams, write tmp files
		}
		if *prebakedFodderPath != "" {
			readPrebaked(*prebakedFodderPath, *tmpPath) // read previously-generated files, write tmp files
		}
		if *txtFodderPath != "" {
			readTextFiles(*txtFodderPath, *tmpPath) // read txts, write tmp files
		}
		if *wikiFodderPath != "" {
			readWikis(*wikiFodderPath, *tmpPath) // read wikis, write tmp files
		}
		if *xwdFodderPath != "" {
			readXWdLists(*xwdFodderPath, *tmpPath) // read Xwd word lists, write tmp files
		}
	}
	Reduce(*tmpPath, *outPath) // read tmp files, add up final number
}
