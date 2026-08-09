package names

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strings"

	"tas/internal/util"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

const (
	GenerateCountFlagName = "count"

	minLength = 4
	maxLength = 9
)

// Note: This is partially AI-generated code!
var GenerateCmdConfig = &cobra.Command{

	Use:   "generate",
	Short: "generate new system names based on a 2nd-order Markov Chain trained on existing names",
	Run:   generateCmd,
}

func generateCmd(cmd *cobra.Command, args []string) {
	//set up logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logCfg := zerolog.ConsoleWriter{Out: os.Stdout}
	log := zerolog.New(logCfg)

	//open the file
	log.Info().Msg("Opening world names file...")
	fname := fmt.Sprintf("%s%s", defaultWorldNamesPath, defaultWorldNamesFile)
	rawLines, err := util.ReadWorldNamesFromFile(fname)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to read from world names file")
	}

	//get the number of names to generate
	count, err := cmd.Flags().GetUint(GenerateCountFlagName)
	if err != nil {
		panic(1)
	}

	// Instantly train on execution (< 1ms overhead)
	engine := newGenerator(rawLines)

	// generate and output names
	fmt.Printf("=== GENERATED PLANET NAMES (N=%d) ===\n", count)
	cfg := filterConfig{
		AllowDuplicates:          false,
		MaxConsecutiveVowels:     3,
		MaxConsecutiveConsonants: 3,
	}
	filter := newFilterPipeline(rawLines, cfg)
	for i := 0; i < int(count); i++ {
		name, err := engine.generateFiltered(minLength, maxLength, filter)
		if err != nil {
			// Safely drop/retry or track error count
			continue
		}
		fmt.Println(name)
	}
}

// -------------------------------------
// Markov 2nd order training and generating
// with filter
// -------------------------------------

type generator struct {
	// Key: 2-character state prefix (e.g., "^^", "^a", "th")
	// Value: Slice of runes representing all valid next steps (weighted by frequency)
	transitions map[string][]rune
}

func newGenerator(names []string) *generator {
	g := &generator{
		transitions: make(map[string][]rune),
	}
	g.train(names)
	return g
}

// train populates the transition map with start and end anchors.
func (g *generator) train(names []string) {
	for _, name := range names {
		cleaned := strings.TrimSpace(strings.ToLower(name))
		if cleaned == "" {
			continue
		}

		// Pad with anchor boundaries for a 2nd-order chain
		// e.g., "terra" becomes "^^terra$"
		padded := "^^" + cleaned + "$"
		runes := []rune(padded)

		// Slide windows of 2 characters to map to the 3rd character
		for i := 0; i < len(runes)-2; i++ {
			prefix := string(runes[i : i+2])
			nextChar := runes[i+2]

			// Appending naturally builds our probability weight
			g.transitions[prefix] = append(g.transitions[prefix], nextChar)
		}
	}
}

// generate creates a single new name based on corpus weights.
func (g *generator) generate(minLength, maxLength int) (string, error) {
	var sb strings.Builder

	// Start with the initial state
	prefix := "^^"

	for sb.Len() <= maxLength {
		nextChoices, exists := g.transitions[prefix]
		if !exists || len(nextChoices) == 0 {
			// Hit an unexpected dead-end state
			return "", errors.New("generation trapped in a terminal dead-end state")
		}

		// Cryptographically secure pseudo-random index selection
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(nextChoices))))
		if err != nil {
			return "", err
		}
		nextChar := nextChoices[n.Int64()]

		// If we hit the natural end token, evaluate if it fits constraints
		if nextChar == '$' {
			if sb.Len() >= minLength {
				break
			}
			// If it's too short, reset and try again to keep execution fast
			sb.Reset()
			prefix = "^^"
			continue
		}

		sb.WriteRune(nextChar)

		// Shift our prefix window over by 1 character
		// e.g., if prefix was "^^" and nextChar was 't', new prefix is "^t"
		// If prefix was "^t" and nextChar was 'h', new prefix is "th"
		runes := []rune(prefix)
		prefix = string(runes[1:]) + string(nextChar)
	}

	// Final verification of boundaries
	if sb.Len() < minLength || sb.Len() > maxLength {
		// Fallback iteration rule to ensure we return clean data
		return g.generate(minLength, maxLength)
	}

	// Capitalize the first letter for your Traveller system output
	raw := sb.String()
	return strings.ToUpper(raw[:1]) + raw[1:], nil
}

// GenerateFiltered builds a name and forces it through the FilterPipeline rules
func (g *generator) generateFiltered(minLength, maxLength int, filter *filterPipeline) (string, error) {
	maxAttempts := 100 // Prevent infinite loops if criteria are overly restrictive

	for attempt := 0; attempt < maxAttempts; attempt++ {
		name, err := g.generate(minLength, maxLength)
		if err != nil {
			continue
		}

		if filter.isValid(name) {
			return name, nil
		}
	}

	return "", errors.New("failed to generate a valid name within filtering constraints")
}

// -------------------------------------
// Names filter
// -------------------------------------

type filterConfig struct {
	AllowDuplicates          bool
	MaxConsecutiveVowels     int
	MaxConsecutiveConsonants int
}

type filterPipeline struct {
	config     filterConfig
	corpusMap  map[string]bool
	vowelRegex *regexp.Regexp
	consRegex  *regexp.Regexp
}

func newFilterPipeline(trainingNames []string, cfg filterConfig) *filterPipeline {
	// Build lookup map for fast uniqueness validation
	cMap := make(map[string]bool)
	for _, name := range trainingNames {
		cMap[strings.ToLower(strings.TrimSpace(name))] = true
	}

	// Compile high-performance regex strings based on config
	// e.g., if MaxConsecutiveVowels is 3, regex checks for [aeiou]{3,}
	vowPattern := `[aeiou]{` + string(rune('0'+cfg.MaxConsecutiveVowels)) + `,}`
	consPattern := `[^aeiou]{` + string(rune('0'+cfg.MaxConsecutiveConsonants)) + `,}`

	return &filterPipeline{
		config:     cfg,
		corpusMap:  cMap,
		vowelRegex: regexp.MustCompile(vowPattern),
		consRegex:  regexp.MustCompile(consPattern),
	}
}

// isValid checks if a generated name clears all aesthetic and structural bars.
func (f *filterPipeline) isValid(name string) bool {
	lowerName := strings.ToLower(name)

	// 1. Minimum sanity check for ultra-short generation loops
	if len(lowerName) < 3 {
		return false
	}

	// 2. Uniqueness Filter (Block direct plagiarism of training corpus)
	if !f.config.AllowDuplicates && f.corpusMap[lowerName] {
		return false
	}

	// 3. Phonotactic Constraints via Regex
	// Catch unpronounceable vowel stacks (e.g., "Thiooa")
	if f.vowelRegex.MatchString(lowerName) {
		return false
	}
	// Catch unpronounceable consonant stacks (e.g., "Grdst")
	if f.consRegex.MatchString(lowerName) {
		return false
	}

	// 4. Strict Letter Constraints (e.g., The 'Q' Rule)
	// Since there are no spaces or special characters, "q" must always be followed by a vowel
	if strings.Contains(lowerName, "q") {
		// Ensure it's not a trailing Q, and is followed by 'u' or another vowel
		qIdx := strings.Index(lowerName, "q")
		if qIdx == len(lowerName)-1 {
			return false // Ends in Q
		}
		nextChar := lowerName[qIdx+1]
		if !strings.ContainsRune("aeiou", rune(nextChar)) {
			return false // Q followed by a consonant (like 'qk' or 'qt')
		}
	}

	// 5. Stuttering / Syllable Repetition Detector
	// Catches Markov loops like "Balalala" or "Xenonon"
	if f.hasExcessiveRepetition(lowerName) {
		return false
	}

	return true
}

// detect repeating sub-string patterns of length 2 or 3
func (f *filterPipeline) hasExcessiveRepetition(s string) bool {
	// Check for repeating bigrams (e.g., "ananan")
	for i := 0; i < len(s)-3; i++ {
		bigram := s[i : i+2]
		if strings.Count(s, bigram) >= 3 {
			return true
		}
	}
	// Check for repeating trigrams (e.g., "ororor")
	for i := 0; i < len(s)-5; i++ {
		trigram := s[i : i+3]
		if strings.Count(s, trigram) >= 2 && strings.Contains(s, trigram+trigram) {
			return true
		}
	}
	return false
}
