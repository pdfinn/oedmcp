package dict

import (
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