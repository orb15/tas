package names

import (
	"fmt"
	"math"
	"os"
	"strings"

	"tas/internal/util"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// Note: This is partially AI-generated code!
var AnalyzeCmdConfig = &cobra.Command{

	Use:   "analyze",
	Short: "perform an analysis on the default world names file to determine suitability for use in Markov Chain generation",
	Run:   analyzeCmd,
}

func analyzeCmd(cmd *cobra.Command, args []string) {
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

	analyzer := newAnalyser(rawLines)
	res := analyzer.analyze()

	fmt.Println("=== CORPUS DIAGNOSTIC REPORT ===")
	fmt.Printf("Total Names Processed: %d (%d Unique)\n", res.TotalNames, res.UniqueNames)
	fmt.Printf("Average Name Length:   %.2f characters\n", res.AvgLength)
	fmt.Println("--------------------------------")
	fmt.Printf("Conditional Entropy:   %.4f bits\n", res.ConditionalEntropy)
	fmt.Printf("Dead-End Sparsity:     %.2f%%\n", res.SparsityPercentage)
	fmt.Printf("Phonetic CV Diversity: %d unique patterns\n", res.UniqueCVPatterns)
	fmt.Println("--------------------------------")
	fmt.Printf("Top Starting Bigrams:  %v\n", res.TopStarts)
	fmt.Printf("Top Ending Characters: %v\n", res.TopEnds)
	fmt.Println("--------------------------------")
	fmt.Printf("Alphabetical Gaps (Starts):\n")
	if len(res.MissingStarts) > 0 {
		fmt.Printf("❌ Completely Missing:     %v\n", res.MissingStarts)
	} else {
		fmt.Println("✅ Completely Missing:     None (All 26 letters represented)")
	}
	if len(res.CriticallyLowStarts) > 0 {
		fmt.Printf("⚠️  Critically Low (<1%%):   %v\n", res.CriticallyLowStarts)
	}

	// Actionable Hobbyist Interpretation
	fmt.Println("\n=== HEALTH ASSESSMENT ===")
	if res.ConditionalEntropy < 1.5 {
		fmt.Println("⚠️  WARNING: Low entropy! The Markov chain will heavily copy your training data verbatim.")
	} else if res.ConditionalEntropy > 3.2 {
		fmt.Println("⚠️  WARNING: High entropy! Names will be highly unpredictable and likely unpronounceable.")
	} else {
		fmt.Println("✅ Clear structural balance found for predictable yet unique generation.")
	}

	if res.SparsityPercentage > 25.0 {
		fmt.Println("⚠️  WARNING: High Dead-End Sparsity. The generator will frequently get stuck mid-generation.")
	}

}

// -------------------------------------
// Analysis engine
// -------------------------------------

type analysisResult struct {
	TotalNames          int
	UniqueNames         int
	AvgLength           float64
	ConditionalEntropy  float64
	SparsityPercentage  float64
	UniqueCVPatterns    int
	TopStarts           []string
	TopEnds             []string
	MissingStarts       []string
	CriticallyLowStarts []string
}

// analyser processes the raw training corpus.
type analyser struct {
	names []string
}

func newAnalyser(names []string) *analyser {
	cleaned := make([]string, 0, len(names))
	for _, name := range names {
		trimmed := strings.TrimSpace(strings.ToLower(name))
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return &analyser{names: cleaned}
}

// analyze runs all practical structural assessments for a 2nd-order Markov chain.
func (a *analyser) analyze() analysisResult {
	res := analysisResult{
		TotalNames: len(a.names),
	}
	if res.TotalNames == 0 {
		return res
	}

	uniqueMap := make(map[string]bool)
	cvPatterns := make(map[string]bool)
	var totalLen int

	// Track single starting character counts
	startLetterCounts := make(map[rune]int)
	// Initialize all 26 letters to ensure we catch absolute zeros
	for r := 'a'; r <= 'z'; r++ {
		startLetterCounts[r] = 0
	}

	for _, name := range a.names {
		uniqueMap[name] = true
		totalLen += len(name)
		cvPatterns[a.toCV(name)] = true
	}
	res.UniqueNames = len(uniqueMap)
	res.AvgLength = float64(totalLen) / float64(res.TotalNames)
	res.UniqueCVPatterns = len(cvPatterns)

	transitions := make(map[string]map[rune]int)
	prefixCounts := make(map[string]int)
	startCounts := make(map[string]int)
	endCounts := make(map[string]int)

	for _, name := range a.names {
		runes := []rune(name)
		if len(runes) < 2 {
			continue
		}

		// Track single starting letter
		startLetterCounts[runes[0]]++

		startCounts[string(runes[:2])]++
		endCounts[string(runes[len(runes)-1:])]++

		for i := 0; i < len(runes)-2; i++ {
			prefix := string(runes[i : i+2])
			next := runes[i+2]

			if _, exists := transitions[prefix]; !exists {
				transitions[prefix] = make(map[rune]int)
			}
			transitions[prefix][next]++
			prefixCounts[prefix]++
		}
	}

	// Calculate Alphabet Coverage and Biases
	criticalThreshold := float64(res.TotalNames) * 0.01 // 1% Threshold
	for r := 'a'; r <= 'z'; r++ {
		count := startLetterCounts[r]
		charStr := string(r)
		if count == 0 {
			res.MissingStarts = append(res.MissingStarts, charStr)
		} else if float64(count) < criticalThreshold {
			res.CriticallyLowStarts = append(res.CriticallyLowStarts, fmt.Sprintf("%s (%d)", charStr, count))
		}
	}

	// ... Keep Entropy and Sparsity calculations identical to previous version ...
	var totalTransitions int
	for _, count := range prefixCounts {
		totalTransitions += count
	}
	var conditionalEntropy float64
	for prefix, nextMap := range transitions {
		pPrefix := float64(prefixCounts[prefix]) / float64(totalTransitions)
		var internalSum float64
		for _, count := range nextMap {
			pNextGivenPrefix := float64(count) / float64(prefixCounts[prefix])
			internalSum += pNextGivenPrefix * math.Log2(pNextGivenPrefix)
		}
		conditionalEntropy += pPrefix * (-internalSum)
	}
	res.ConditionalEntropy = conditionalEntropy

	var deadEnds int
	for _, name := range a.names {
		runes := []rune(name)
		if len(runes) >= 2 {
			lastBigram := string(runes[len(runes)-2:])
			if _, hasTransitions := transitions[lastBigram]; !hasTransitions {
				deadEnds++
			}
		}
	}
	res.SparsityPercentage = (float64(deadEnds) / float64(len(prefixCounts))) * 100
	res.TopStarts = a.getTopKeys(startCounts, 3)
	res.TopEnds = a.getTopKeys(endCounts, 3)

	return res
}

// Convert a name to Consonant-Vowel structural token
func (a *analyser) toCV(name string) string {
	var sb strings.Builder
	vowels := "aeiou"
	for _, r := range name {
		if strings.ContainsRune(vowels, r) {
			sb.WriteRune('V')
		} else {
			sb.WriteRune('C')
		}
	}
	return sb.String()
}

// Helper to get top N heavily biased entries
func (a *analyser) getTopKeys(m map[string]int, limit int) []string {
	var sorted []string
	// Simple bubble/insertion sort for small map extraction to keep dependencies zero
	for len(m) > 0 && len(sorted) < limit {
		maxVal := -1
		maxKey := ""
		for k, v := range m {
			if v > maxVal {
				maxVal = v
				maxKey = k
			}
		}
		sorted = append(sorted, fmt.Sprintf("%s (%d)", maxKey, maxVal))
		delete(m, maxKey)
	}
	return sorted
}
