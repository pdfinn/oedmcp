package dict

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type Entry struct {
	Word       string
	Definition string
	Etymology  string
	Offset     int64
	Length     int
}

type OEDDict struct {
	dataFile  *os.File
	indexFile *os.File
}

func NewOEDDict(dataPath, indexPath string) (*OEDDict, error) {
	dataFile, err := os.Open(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open OED data file: %w", err)
	}

	indexFile, err := os.Open(indexPath)
	if err != nil {
		dataFile.Close()
		return nil, fmt.Errorf("failed to open OED index file: %w", err)
	}

	return &OEDDict{
		dataFile:  dataFile,
		indexFile: indexFile,
	}, nil
}

func (d *OEDDict) Close() {
	if d.dataFile != nil {
		d.dataFile.Close()
	}
	if d.indexFile != nil {
		d.indexFile.Close()
	}
}

func (d *OEDDict) LookupWord(word string) (*Entry, error) {
	// Normalize the word for lookup
	normalizedWord := strings.ToLower(strings.TrimSpace(word))

	// Search in the index file
	offset, err := d.searchIndex(normalizedWord)
	if err != nil {
		return nil, fmt.Errorf("word not found in index: %w", err)
	}

	// Read the entry from the data file
	entry, err := d.readEntry(offset)
	if err != nil {
		return nil, fmt.Errorf("failed to read entry: %w", err)
	}

	return entry, nil
}

func (d *OEDDict) searchIndex(word string) (int64, error) {
	// Reset to beginning of index file
	d.indexFile.Seek(0, 0)
	scanner := bufio.NewScanner(d.indexFile)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			indexWord := strings.ToLower(parts[0])
			if indexWord == word {
				offset, err := strconv.ParseInt(parts[1], 10, 64)
				if err != nil {
					return 0, fmt.Errorf("invalid offset in index: %w", err)
				}
				return offset, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error reading index: %w", err)
	}

	return 0, errors.New("word not found")
}

func (d *OEDDict) readEntry(offset int64) (*Entry, error) {
	// Seek to the offset in the data file
	_, err := d.dataFile.Seek(offset, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to offset: %w", err)
	}

	// Read the entry - we'll read a reasonable chunk and parse it
	buffer := make([]byte, 32768) // 32KB should be enough for most entries
	n, err := d.dataFile.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	// Parse the entry data
	entry := &Entry{
		Offset: offset,
		Length: n,
	}

	// The OED2 data format uses special tags and formatting
	// We'll do a simplified parsing here
	entryData := buffer[:n]
	entry.Definition = d.parseEntryData(entryData)

	// Extract word and etymology if possible
	entry.Word, entry.Etymology = d.extractWordAndEtymology(entry.Definition)

	return entry, nil
}

func (d *OEDDict) parseEntryData(data []byte) string {
	// This is a simplified parser for OED2 data
	// The actual format is complex with many tags and special characters
	result := &strings.Builder{}

	inTag := false
	for i := 0; i < len(data); i++ {
		b := data[i]

		// Handle special characters and tags
		if b == 0x00 {
			// Null terminator - end of entry
			break
		} else if b == 0x01 || b == 0x02 {
			// Tag markers
			inTag = !inTag
			continue
		} else if b < 0x20 && b != 0x0A && b != 0x0D {
			// Skip other control characters except newlines
			continue
		}

		if !inTag && unicode.IsPrint(rune(b)) {
			result.WriteByte(b)
		}
	}

	return strings.TrimSpace(result.String())
}

func (d *OEDDict) extractWordAndEtymology(definition string) (string, string) {
	word := ""
	etymology := ""

	// Extract headword from <hw>...</hw> tags
	if hwStart := strings.Index(definition, "<hw>"); hwStart != -1 {
		if hwEnd := strings.Index(definition[hwStart:], "</hw>"); hwEnd != -1 {
			word = definition[hwStart+4 : hwStart+hwEnd]
			// Clean up any remaining tags or special characters
			word = strings.TrimSpace(word)
		}
	}

	// If no <hw> tag found, try first line approach
	if word == "" {
		lines := strings.Split(definition, "\n")
		if len(lines) > 0 {
			word = strings.TrimSpace(lines[0])
			word = strings.TrimRight(word, ".,;:1234567890 ")
		}
	}

	// Extract etymology from <etym>...</etym> tags
	if etymStart := strings.Index(definition, "<etym>"); etymStart != -1 {
		if etymEnd := strings.Index(definition[etymStart:], "</etym>"); etymEnd != -1 {
			etymology = definition[etymStart+6 : etymStart+etymEnd]
			// Clean up nested tags for readability
			etymology = cleanupTags(etymology)
		}
	}

	return word, etymology
}

// cleanupTags removes or simplifies common OED tags for better readability
func cleanupTags(text string) string {
	// Remove or replace common tags
	replacements := []struct {
		old, new string
	}{
		{"<cf>", ""},
		{"</cf>", ""},
		{"<xr>", ""},
		{"</xr>", ""},
		{"<x>", ""},
		{"</x>", ""},
		{"<n>", " "},
		{"</n>", ""},
		{"<xs>", ""},
		{"</xs>", ""},
		{"&oq.", "'"},
		{"&cq.", "'"},
		{"&emac.", "ē"},
		{"&amac.", "ā"},
		{"&imac.", "ī"},
		{"&omac.", "ō"},
		{"&umac.", "ū"},
		{"&eacu.", "é"},
		{"&aacu.", "á"},
		{"&iacu.", "í"},
		{"&oacu.", "ó"},
		{"&uacu.", "ú"},
	}

	result := text
	for _, r := range replacements {
		result = strings.ReplaceAll(result, r.old, r.new)
	}

	// Clean up multiple spaces
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}

	return strings.TrimSpace(result)
}

// SearchPrefix returns entries that start with the given prefix
func (d *OEDDict) SearchPrefix(prefix string) ([]string, error) {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	var results []string

	d.indexFile.Seek(0, 0)
	scanner := bufio.NewScanner(d.indexFile)

	count := 0
	maxResults := 20

	for scanner.Scan() && count < maxResults {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) >= 1 {
			word := strings.ToLower(parts[0])
			if strings.HasPrefix(word, prefix) {
				results = append(results, parts[0])
				count++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error searching index: %w", err)
	}

	return results, nil
}

// GetRandomEntry returns a random dictionary entry
func (d *OEDDict) GetRandomEntry() (*Entry, error) {
	// Get file size to determine random position
	info, err := d.indexFile.Stat()
	if err != nil {
		return nil, err
	}

	// Read a random line from the index
	d.indexFile.Seek(info.Size()/2, 0) // Start from middle

	scanner := bufio.NewScanner(d.indexFile)
	// Skip partial line
	if scanner.Scan() {
		// Now read the next complete line
		if scanner.Scan() {
			line := scanner.Text()
			parts := strings.Split(line, "\t")
			if len(parts) >= 2 {
				offset, err := strconv.ParseInt(parts[1], 10, 64)
				if err == nil {
					return d.readEntry(offset)
				}
			}
		}
	}

	return nil, errors.New("failed to get random entry")
}

// Helper function to handle the complex OED2 data format
func parseOEDTags(data []byte) map[string]string {
	tags := make(map[string]string)

	// This would need to implement the full tag parsing from the Plan 9 code
	// For now, we'll return a simplified version

	return tags
}

// ReadEntryRaw reads raw entry data for debugging
func (d *OEDDict) ReadEntryRaw(offset int64, length int) ([]byte, error) {
	_, err := d.dataFile.Seek(offset, 0)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, length)
	n, err := d.dataFile.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return buffer[:n], nil
}