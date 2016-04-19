package main

import (
	"bufio"
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
    fmt.Fprintf(os.Stderr, "  $ %s --txtpath dumpz/textfiles --wikipath dumpz/mediawiki\n", os.Args[0])
    fmt.Fprintf(os.Stderr, "  Go get some lunch. Later, ./Phrases_*.txt has tab-separated freq,phrase info.\n")
    fmt.Fprintf(os.Stderr, "\n")
    flag.PrintDefaults()
    fmt.Fprintf(os.Stderr, "\n")
    fmt.Fprintf(os.Stderr, "  More info at http://github.com/lahosken/misc/phraser\n")
    fmt.Fprintf(os.Stderr, "\n")
    os.Exit(0)
  }
}

var cpuprofile = flag.String("cpuprofile", "", "Write cpu profile to file")
var wikiFodderPath = flag.String("wikipath", "", "Dir full of wiki-export .xml files")
var txtFodderPath = flag.String("txtpath", "", "Dir full of .txt files")
var tmpPath = flag.String("tmppath", "", "Don't parse wikis/textfiles. Instead, read from previously-generated tmp dir")
var outPath = flag.String("outpath", "", "Instead of writing to ~/Phrases_20160419_131415.txt, write to this path")

const (
  TMP_FILENAME_FORMAT = "p-%s-%09d.txt"
  DICT_TAMP_THRESHHOLD_ENTRIES = 6000000
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
  // bolded. These spans ?often? indicate especially interesting text
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
)

type counter struct {
  d map[string]uint64
}
func (c *counter) boost(s string, n uint64) {
  if c.d == nil {
    c.d = map[string]uint64{}
  }
  c.d[s] += n
}
func (c *counter) inc(s string) {
  c.boost(s, 1)
}
func (c *counter) tamp() {
  for s, score := range c.d {
    if score <= uint64(strings.Count(s, " ")) {
      delete(c.d, s)
    }
  }
}

// Write out contents to a file.
func (c counter) persist(outPath string) {
  log.Printf(" PERSIST %v", outPath)
  if len(c.d) > 4000000 {
    log.Printf("   SORT...")
  }
  sortMe := map[uint64]([]string){}
  maxScore := uint64(0)
  for phrase, score := range c.d {
    if score > maxScore { maxScore = score }
    if sortMe[score] == nil {
      sortMe[score] = []string{}
    }
    sortMe[score] = append(sortMe[score], phrase)
  }
  if len(c.d) > 4000000 {
    log.Printf("   BIG SORT DONE")
  }
  os.MkdirAll(filepath.Dir(outPath), 0776)
  outF, err := os.Create(outPath)
  if err != nil {
    log.Fatalf("Couldn't open outfile %v %v", outPath, err)
  }
  writtenCount := 0
  lastSync := 0
  for score := maxScore; score > 0; score -= 1 {
    if sortMe[score] == nil { continue }
    sort.Sort(sort.StringSlice(sortMe[score]))
    for _, phrase := range sortMe[score] {
      outF.WriteString(fmt.Sprintf("%d\t%s\n", score, phrase))
      writtenCount += 1
    }
    delete(sortMe, score)
    if writtenCount > lastSync + 5000 {
      outF.Sync()
      lastSync = writtenCount
    }
    if writtenCount > DICT_TAMP_THRESHHOLD_ENTRIES { break }
  }
  outF.Close()
}

// Takes a string like "Is Omotic Afro-Asiatic?",
// returns ["is", "omotic", "afro", "asiatic"]
func tokenize(snippet string) (tokens []string) {
  // "Don't fear the reaper!" -> "dont fear the reaper"
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

// Given mediawiki blob, boost count of found phrases
func ingestWikiPage(page string, co *counter) {
  page = strings.ToLower(page)
  textModeP := false // are we in the page's <text> element?
  found := new(counter)
  for _, line := range strings.Split(page, "\n") {
    if !textModeP {
      if redirREMatch := redirRE.FindStringSubmatch(line); redirREMatch != nil {
        line2snippets(redirREMatch[1], found)
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
        line2snippetsPastBrackets(titleREMatch[1], found)
        line2snippetsPastBrackets(titleREMatch[1], found)
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
  line = strings.Replace(line, "&#039;", "'", -1)
  snippets.inc(line)
  return
}

// given a string like "one two three",
// tally up "one" "two" "one two" "three"
// "two three" "one two three"
func tallySnippet(co *counter, snippet string) {
  tokens := tokenize(snippet)
  for startIx, _ := range tokens {
    for endIx := startIx+1; endIx <= len(tokens); endIx += 1 {
      key := strings.Join(tokens[startIx:endIx], " ")
      co.inc(key)
      if len(key) > 35 {
        break
      }
    }
  }
}

func tallySnippets(tally *counter, found counter) {
  for snippet, count := range found.d {
    tokens := tokenize(snippet)
    for startIx, _ := range tokens {
      for endIx := startIx+1; endIx <= len(tokens); endIx += 1 {
        key := strings.Join(tokens[startIx:endIx], " ")
        tally.boost(key, count)
        if len(key) > 35 {
          break
        }      
      }
    }
  }
}

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

// read data from text files, write it to tmp files
func ReadTextfiles(fodderPath, tmpPath string) {
  log.Printf("WRITING TO %s", tmpPath)
  paraCount := 0
  inFilePaths, _ := filepath.Glob(filepath.Join(fodderPath, "*.txt"))
  for _, inFilePath := range inFilePaths {
    shortName := strings.Split(filepath.Base(inFilePath), ".")[0]
    log.Printf("READING %v", inFilePath)
    fodderF, err := os.Open(inFilePath)
    if err != nil {
      log.Fatalf("Error opening txt file %s %v", inFilePath, err)
    }
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
    found.inc(para)
    tally := counter{}
    tallySnippets(&tally, found)
    tallyPath := filepath.Join(tmpPath, fmt.Sprintf(TMP_FILENAME_FORMAT, shortName, paraCount))
    tally.persist(tallyPath)
  }
}
// read data from wikis, write it to tmp files
func ReadWikis(fodderPath, tmpPath string) {
  log.Printf("WRITING TO %s", tmpPath)
  pageCount := 0
  inFilePaths, _ := filepath.Glob(filepath.Join(fodderPath, "*.xml"))
  for _, inFilePath := range inFilePaths {
    wikiName := strings.Split(filepath.Base(inFilePath), "_")[0]
    wikiName = strings.Split(wikiName, "-")[0]
    log.Printf("READING %v", inFilePath)
    fodderF, err := os.Open(inFilePath)
    if err != nil {
      log.Fatalf("Error opening wiki file %s %v", inFilePath, err)
    }
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
      if (strings.HasPrefix(line, "  <page>")) { 
        pageCount += 1
        page = ""
        continue
      }
      if (strings.HasPrefix(line, "  </page>")) {
        ingestWikiPage(page, co)
        if len(co.d) > 4 * DICT_TAMP_THRESHHOLD_ENTRIES {
          tampCount += 1
          log.Printf(" TAMP %d", tampCount)
          co.tamp()
          if len(co.d) > DICT_TAMP_THRESHHOLD_ENTRIES {
            // if STILL full, rotate out...
            tallyPath := filepath.Join(tmpPath, fmt.Sprintf(TMP_FILENAME_FORMAT, wikiName, pageCount))
            co.persist(tallyPath)
            co = new(counter)
            tampCount = 0
          }
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
    tallyPath := filepath.Join(tmpPath, fmt.Sprintf(TMP_FILENAME_FORMAT, wikiName, pageCount))
    co.persist(tallyPath)
  }
}

func Reduce(tmpPath string, outPath string) {
  tmpFileGlobPattern := regexp.MustCompile(`%.*?d`).ReplaceAllString(TMP_FILENAME_FORMAT, "*")
  tmpFilenames, err := filepath.Glob(filepath.Join(tmpPath, tmpFileGlobPattern))
  if err != nil {
    log.Fatal("Couldnt glob temp files %v", err)
  }
  bigCounter := counter{}
  lineRE := regexp.MustCompile(`(\d+)\t(.*)`)

  // We make two passes.
  //
  // If we try to pick up ALL the entries, we fill up our memory with
  // a bunch of phrases that aren't actually common. So instead, two passes
  //   first: count everything w/count >1
  //   second: for items w/ count <=1, "boost" things we saw in 1st pass
  //
  // No, this isn't super-rigorous; it leaves out some things. But those
  // things mmmostly wouldn't appear in the first 1M phrases, so good enough
  // for our purposes.
  // E.g., "disavow any" is _darned_ unusual in that it shows up as a "1" in
  // a dozen places but never higher than 1. Giving it a theoretical score of
  // 9704, about 3.4millionth in the list, i.e., pretty darned far down.
  //
  // Instead of magic number 2, could use 5 and still get pretty good results.
  // But 2 means we aren't abusing this un-rigorous trick so egregiously.
  // Some slapdash experimenting (see maxes.py script) suggests that 3 would
  // be a pretty good number if want to catch things that'll end up with score
  // > 10000

  // First pass
  for _, tmpFilename := range tmpFilenames {
    log.Printf("READING %v", tmpFilename)
    tmpF, err := os.Open(tmpFilename)
    if err != nil {
      log.Fatalf("Error opening tmp file %s %v", tmpFilename, err)
    }
    scan := bufio.NewScanner(tmpF)
    for {
      fileNotDone := scan.Scan();
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
      if score <= 1 { break }
      phrase := match[2]
      bigCounter.boost(phrase, uint64(8.0 * math.Log1p(float64(score))))
    }
    tmpF.Close()
  }
  // Second pass
  for _, tmpFilename := range tmpFilenames {
    log.Printf("READING %v", tmpFilename)
    tmpF, err := os.Open(tmpFilename)
    if err != nil {
      log.Fatalf("Error opening tmp file %s %v", tmpFilename, err)
    }
    scan := bufio.NewScanner(tmpF)
    for {
      if fileNotDone := scan.Scan() ; !fileNotDone {
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
      if score > 1 { continue }
      phrase := match[2]
      // only "boost" existing, no new stuff
      if _, contains := bigCounter.d[phrase]; contains {
        bigCounter.boost(phrase, uint64(8.0 * math.Log1p(float64(score))))
      }
    }
    tmpF.Close()
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
  if *tmpPath == "" && *wikiFodderPath == "" && *txtFodderPath == "" {
    log.Fatal("Neither --wikipath nor --txtpath set. Nothing to do!")
  }
  if *tmpPath == "" {
    *tmpPath = filepath.Join(os.TempDir(), "phraser", nowPath)
    if *txtFodderPath != "" {
      ReadTextfiles(*txtFodderPath, *tmpPath) // read txts, write tmp files
    }
    if *wikiFodderPath != "" {
      ReadWikis(*wikiFodderPath, *tmpPath) // read wikis, write tmp files
    }
  }
  Reduce(*tmpPath, *outPath) // read tmp files, add up final number
}
