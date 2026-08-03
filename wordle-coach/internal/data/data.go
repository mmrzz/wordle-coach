// Package data provides the wordle word lists and their corpus frequencies.
//
// The source files are embedded into the binary, so Load needs no filesystem
// access and does not care about the process working directory.
package data

import (
	"bufio"
	_ "embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// WordLen is the length of every word in the lists.
const WordLen = 5

//go:embed answers.txt
var answersFile string

//go:embed allowed.txt
var allowedFile string

//go:embed freq.tsv
var freqFile string

// Set is the corpus a solver works from.
type Set struct {
	// Answers are the words that can be the solution. Narrow the candidates
	// down within this list.
	Answers []string

	// Allowed is every word accepted as a guess, answers included, sorted.
	// Guesses are drawn from this list, which is far larger than Answers.
	Allowed []string

	freq       map[string]float64
	allowedSet map[string]struct{}
}

var (
	once      sync.Once
	loaded    *Set
	loadedErr error
)

// Load returns the embedded corpus. The files are parsed on the first call and
// the same *Set is handed out afterwards, so callers may call it whenever they
// need the data. It is safe for concurrent use, and the returned Set must be
// treated as read-only.
func Load() (*Set, error) {
	once.Do(func() {
		loaded, loadedErr = load()
	})
	return loaded, loadedErr
}

func load() (*Set, error) {
	answers, err := parseWords(answersFile, "answers.txt")
	if err != nil {
		return nil, err
	}
	extras, err := parseWords(allowedFile, "allowed.txt")
	if err != nil {
		return nil, err
	}
	freq, err := parseFreq(freqFile, "freq.tsv")
	if err != nil {
		return nil, err
	}

	// allowed.txt holds only the guesses that are not answers, so the two
	// lists are merged. The set also de-duplicates, in case that changes.
	allowedSet := make(map[string]struct{}, len(answers)+len(extras))
	for _, w := range answers {
		allowedSet[w] = struct{}{}
	}
	for _, w := range extras {
		allowedSet[w] = struct{}{}
	}
	allowed := make([]string, 0, len(allowedSet))
	for w := range allowedSet {
		allowed = append(allowed, w)
	}
	sort.Strings(allowed)

	// A word with no frequency would silently rank as the rarest one there is,
	// which is hard to spot in solver output, so catch it at load time.
	var missing []string
	for _, w := range allowed {
		if _, ok := freq[w]; !ok {
			missing = append(missing, w)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("data: freq.tsv is missing %d word(s), e.g. %q: re-run frequecy_script/parse_words.py",
			len(missing), missing[0])
	}

	return &Set{
		Answers:    answers,
		Allowed:    allowed,
		freq:       freq,
		allowedSet: allowedSet,
	}, nil
}

// IsAllowed reports whether word is a legal guess.
func (s *Set) IsAllowed(word string) bool {
	_, ok := s.allowedSet[word]
	return ok
}

// Freq returns the Zipf frequency of word, roughly 1 (very rare) to 7 (very
// common). Unknown words score 0, as do words the corpus never saw.
func (s *Set) Freq(word string) float64 {
	return s.freq[word]
}

// parseWords reads a file of one word per line. name is used in errors.
func parseWords(contents, name string) ([]string, error) {
	words := make([]string, 0, 16384)

	sc := bufio.NewScanner(strings.NewReader(contents))
	for line := 1; sc.Scan(); line++ {
		word := strings.TrimSpace(sc.Text())
		if word == "" {
			continue
		}
		if err := validate(word); err != nil {
			return nil, fmt.Errorf("data: %s line %d: %w", name, line, err)
		}
		words = append(words, word)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("data: reading %s: %w", name, err)
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("data: %s is empty", name)
	}
	return words, nil
}

// parseFreq reads a "word<TAB>zipf" file. name is used in errors.
func parseFreq(contents, name string) (map[string]float64, error) {
	freq := make(map[string]float64, 16384)

	sc := bufio.NewScanner(strings.NewReader(contents))
	for line := 1; sc.Scan(); line++ {
		text := strings.TrimSpace(sc.Text())
		if text == "" {
			continue
		}

		word, value, ok := strings.Cut(text, "\t")
		if !ok {
			return nil, fmt.Errorf("data: %s line %d: want \"word<TAB>frequency\", got %q", name, line, text)
		}
		if err := validate(word); err != nil {
			return nil, fmt.Errorf("data: %s line %d: %w", name, line, err)
		}
		f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return nil, fmt.Errorf("data: %s line %d: frequency for %q: %w", name, line, word, err)
		}
		freq[word] = f
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("data: reading %s: %w", name, err)
	}
	return freq, nil
}

// validate checks that word is the lowercase five-letter form the solver
// compares against, so a stray entry cannot quietly never match a guess.
func validate(word string) error {
	if len(word) != WordLen {
		return fmt.Errorf("%q is %d letters, want %d", word, len(word), WordLen)
	}
	for _, r := range word {
		if r < 'a' || r > 'z' {
			return fmt.Errorf("%q contains %q, want lowercase a-z", word, r)
		}
	}
	return nil
}
