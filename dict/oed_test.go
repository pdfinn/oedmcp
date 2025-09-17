package dict

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// MockOEDData creates a simple mock OED data file for testing
func createMockOEDFiles(t *testing.T) (string, string) {
	tempDir := t.TempDir()

	// Create mock data file with simplified OED-like format
	dataPath := filepath.Join(tempDir, "oed2")
	dataContent := `<e><hg><hw>test</hw> <pr><ph>tEst</ph></pr></hg>. <etym>f. Latin testum earthen pot</etym> <s4>A procedure for critical evaluation; a means of determining the presence, quality, or truth of something.</s4></e>` + "\x00" +
		`<e><hg><hw>example</hw> <pr><ph>Ig"zA:mp@l</ph></pr></hg>. <etym>f. Latin exemplum</etym> <s4>A thing characteristic of its kind or illustrating a general rule.</s4></e>` + "\x00"

	if err := os.WriteFile(dataPath, []byte(dataContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create mock index file
	indexPath := filepath.Join(tempDir, "oed2index")
	indexContent := "test\t0\nexample\t184\n"
	if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
		t.Fatal(err)
	}

	return dataPath, indexPath
}

func TestNewOEDDict(t *testing.T) {
	t.Run("ValidFiles", func(t *testing.T) {
		dataPath, indexPath := createMockOEDFiles(t)

		dict, err := NewOEDDict(dataPath, indexPath)
		if err != nil {
			t.Fatalf("NewOEDDict failed: %v", err)
		}
		defer dict.Close()

		if dict.dataFile == nil {
			t.Error("dataFile is nil")
		}
		if dict.indexFile == nil {
			t.Error("indexFile is nil")
		}
	})

	t.Run("InvalidDataPath", func(t *testing.T) {
		_, err := NewOEDDict("/nonexistent/oed2", "/nonexistent/oed2index")
		if err == nil {
			t.Error("NewOEDDict should fail with invalid data path")
		}
	})

	t.Run("InvalidIndexPath", func(t *testing.T) {
		tempDir := t.TempDir()
		dataPath := filepath.Join(tempDir, "oed2")
		os.WriteFile(dataPath, []byte("data"), 0644)

		_, err := NewOEDDict(dataPath, "/nonexistent/oed2index")
		if err == nil {
			t.Error("NewOEDDict should fail with invalid index path")
		}
	})
}

func TestLookupWord(t *testing.T) {
	dataPath, indexPath := createMockOEDFiles(t)
	dict, err := NewOEDDict(dataPath, indexPath)
	if err != nil {
		t.Fatal(err)
	}
	defer dict.Close()

	t.Run("ExistingWord", func(t *testing.T) {
		entry, err := dict.LookupWord("test")
		if err != nil {
			t.Fatalf("LookupWord failed: %v", err)
		}

		if entry == nil {
			t.Fatal("Entry is nil")
		}

		if !strings.Contains(entry.Definition, "test") {
			t.Errorf("Definition doesn't contain word 'test': %q", entry.Definition)
		}
	})

	t.Run("CaseInsensitive", func(t *testing.T) {
		entry, err := dict.LookupWord("TEST")
		if err != nil {
			t.Fatalf("LookupWord failed: %v", err)
		}

		if entry == nil {
			t.Fatal("Entry is nil")
		}
	})

	t.Run("NonexistentWord", func(t *testing.T) {
		_, err := dict.LookupWord("nonexistent")
		if err == nil {
			t.Error("LookupWord should fail for nonexistent word")
		}
	})

	t.Run("WordWithSpaces", func(t *testing.T) {
		entry, err := dict.LookupWord("  test  ")
		if err != nil {
			t.Fatalf("LookupWord failed: %v", err)
		}

		if entry == nil {
			t.Fatal("Entry is nil")
		}
	})
}

func TestExtractWordAndEtymology(t *testing.T) {
	dataPath, indexPath := createMockOEDFiles(t)
	dict, err := NewOEDDict(dataPath, indexPath)
	if err != nil {
		t.Fatal(err)
	}
	defer dict.Close()

	t.Run("ExtractFromEntry", func(t *testing.T) {
		entry, err := dict.LookupWord("test")
		if err != nil {
			t.Fatal(err)
		}

		// The word should be extracted
		if entry.Word == "" {
			t.Error("Word not extracted")
		}

		// Etymology should be extracted and cleaned
		if entry.Etymology == "" {
			t.Error("Etymology not extracted")
		}

		// Check that tags are cleaned up
		if strings.Contains(entry.Etymology, "<etym>") {
			t.Error("Etymology still contains tags")
		}
	})
}

func TestSearchPrefix(t *testing.T) {
	dataPath, indexPath := createMockOEDFiles(t)
	dict, err := NewOEDDict(dataPath, indexPath)
	if err != nil {
		t.Fatal(err)
	}
	defer dict.Close()

	t.Run("ValidPrefix", func(t *testing.T) {
		results, err := dict.SearchPrefix("te")
		if err != nil {
			t.Fatalf("SearchPrefix failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("No results returned")
		}

		found := false
		for _, word := range results {
			if word == "test" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected 'test' in results")
		}
	})

	t.Run("NoMatches", func(t *testing.T) {
		results, err := dict.SearchPrefix("xyz")
		if err != nil {
			t.Fatalf("SearchPrefix failed: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Expected no results, got %d", len(results))
		}
	})

	t.Run("CaseInsensitive", func(t *testing.T) {
		results, err := dict.SearchPrefix("TE")
		if err != nil {
			t.Fatalf("SearchPrefix failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("No results returned for uppercase prefix")
		}
	})
}

func TestCleanupTags(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "text with <cf>tag</cf> content",
			expected: "text with tag content",
		},
		{
			input:    "multiple  <n>  spaces  </n>  here",
			expected: "multiple spaces here",
		},
		{
			input:    "&oq.quote&cq. and &emac.macron&emac.",
			expected: "'quote' and ēmacronē",
		},
		{
			input:    "<xr><x>cross-ref</x></xr>",
			expected: "cross-ref",
		},
	}

	for _, test := range tests {
		result := cleanupTags(test.input)
		if result != test.expected {
			t.Errorf("cleanupTags(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestGetRandomEntry(t *testing.T) {
	// Create mock files with multiple entries for testing randomness
	tempDir := t.TempDir()
	dataPath := filepath.Join(tempDir, "oed2")

	// Create data with several distinct entries
	entries := []string{
		`<e><hg><hw>apple</hw></hg> <etym>Old English æppel</etym> <s4>A round fruit.</s4></e>`,
		`<e><hg><hw>banana</hw></hg> <etym>Portuguese banana</etym> <s4>A yellow elongated fruit.</s4></e>`,
		`<e><hg><hw>cherry</hw></hg> <etym>French cerise</etym> <s4>A small red fruit.</s4></e>`,
		`<e><hg><hw>date</hw></hg> <etym>Latin dactylus</etym> <s4>A sweet brown fruit.</s4></e>`,
		`<e><hg><hw>elderberry</hw></hg> <etym>Old English ellærn</etym> <s4>A dark purple berry.</s4></e>`,
	}

	// Build data file with null separators
	var dataContent []byte
	offsets := make(map[string]int64)
	currentOffset := int64(0)

	for _, entry := range entries {
		word := extractWordFromEntry(entry)
		offsets[word] = currentOffset
		entryBytes := []byte(entry)
		dataContent = append(dataContent, entryBytes...)
		dataContent = append(dataContent, 0x00) // null separator
		currentOffset += int64(len(entryBytes) + 1)
	}

	if err := os.WriteFile(dataPath, dataContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Create index file
	indexPath := filepath.Join(tempDir, "oed2index")
	var indexContent []string
	for word, offset := range offsets {
		indexContent = append(indexContent, fmt.Sprintf("%s\t%d", word, offset))
	}

	if err := os.WriteFile(indexPath, []byte(strings.Join(indexContent, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Open dictionary
	dict, err := NewOEDDict(dataPath, indexPath)
	if err != nil {
		t.Fatal(err)
	}
	defer dict.Close()

	t.Run("ReturnsValidEntry", func(t *testing.T) {
		entry, err := dict.GetRandomEntry()
		if err != nil {
			t.Fatalf("GetRandomEntry failed: %v", err)
		}

		if entry == nil {
			t.Fatal("Entry is nil")
		}

		if entry.Word == "" {
			t.Error("Word is empty")
		}

		if entry.Definition == "" {
			t.Error("Definition is empty")
		}
	})

	t.Run("ReturnsRandomEntries", func(t *testing.T) {
		// Get multiple random entries and check for variation
		wordCounts := make(map[string]int)
		numSamples := 50

		for i := 0; i < numSamples; i++ {
			entry, err := dict.GetRandomEntry()
			if err != nil {
				t.Fatalf("GetRandomEntry failed on iteration %d: %v", i, err)
			}

			// Extract the main word from the entry
			word := entry.Word
			if word == "" && entry.Definition != "" {
				// Try to extract from definition if Word is empty
				if idx := strings.Index(entry.Definition, "</hw>"); idx > 0 {
					start := strings.LastIndex(entry.Definition[:idx], "<hw>")
					if start >= 0 {
						word = entry.Definition[start+4 : idx]
					}
				}
			}

			if word != "" {
				wordCounts[word]++
			}
		}

		// We should have gotten at least 2 different words
		if len(wordCounts) < 2 {
			t.Errorf("Expected random entries, but got only %d unique word(s) in %d samples: %v",
				len(wordCounts), numSamples, wordCounts)
		}

		// Check that no single word dominates too heavily (shouldn't be more than 80% of samples)
		for word, count := range wordCounts {
			percentage := float64(count) / float64(numSamples) * 100
			if percentage > 80 {
				t.Errorf("Word '%s' appeared in %.1f%% of samples, suggesting non-random behavior",
					word, percentage)
			}
		}
	})
}

// Helper function to extract word from entry for test setup
func extractWordFromEntry(entry string) string {
	start := strings.Index(entry, "<hw>")
	end := strings.Index(entry, "</hw>")
	if start >= 0 && end > start {
		return entry[start+4 : end]
	}
	return ""
}

func TestClose(t *testing.T) {
	dataPath, indexPath := createMockOEDFiles(t)
	dict, err := NewOEDDict(dataPath, indexPath)
	if err != nil {
		t.Fatal(err)
	}

	// Close should not panic
	dict.Close()

	// Double close should also not panic
	dict.Close()
}