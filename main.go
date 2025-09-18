package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pdfinn/oedmcp/config"
	"github.com/pdfinn/oedmcp/dict"
)

// Format types for OED entries
const (
	FormatFull  = "full"  // Complete entry, cleaned and formatted
	FormatClean = "clean" // Standard format, no XML, readable
	FormatBrief = "brief" // Just definition and etymology
	FormatRaw   = "raw"   // Raw XML for debugging
)

// formatEntry formats a dictionary entry according to the specified format
func formatEntry(entry *dict.Entry, format string, includeEtym bool) string {
	switch format {
	case FormatRaw:
		// Return the original definition with all XML tags
		result := fmt.Sprintf("OED Entry for '%s':\n\nDefinition:\n%s\n", entry.Word, entry.Definition)
		if includeEtym && entry.Etymology != "" {
			result += fmt.Sprintf("\nEtymology:\n%s\n", entry.Etymology)
		}
		return result

	case FormatBrief:
		// Minimal format - just the essentials
		cleaned := cleanDefinitionForBrief(entry.Definition)
		result := fmt.Sprintf("%s: %s", entry.Word, cleaned)
		if includeEtym && entry.Etymology != "" {
			etymClean := stripAllTags(entry.Etymology)
			result += fmt.Sprintf("\nEtymology: %s", etymClean)
		}
		return result

	case FormatFull:
		// Full format - comprehensive cleaned output with all details
		return formatFullEntry(entry, includeEtym)

	case FormatClean:
		fallthrough
	default:
		// Clean format - standard cleaned output
		return formatCleanEntry(entry, includeEtym)
	}
}

// formatFullEntry creates a comprehensive, cleaned format with all details
func formatFullEntry(entry *dict.Entry, includeEtym bool) string {
	var result strings.Builder

	// Header
	result.WriteString(fmt.Sprintf("# Complete OED Entry: %s\n\n", entry.Word))

	// Extract and format pronunciation if present
	if pron := extractPronunciation(entry.Definition); pron != "" {
		result.WriteString(fmt.Sprintf("**Pronunciation:** %s\n\n", pron))
	}

	// Etymology
	if includeEtym && entry.Etymology != "" {
		cleaned := stripAllTags(entry.Etymology)
		result.WriteString("## Etymology\n")
		result.WriteString(fmt.Sprintf("%s\n\n", cleaned))
	}

	// Full Definition with structure preserved
	result.WriteString("## Definition\n\n")

	// Process definition with enhanced cleaning that preserves quotations and dates
	cleanedDef := cleanDefinitionDetailed(entry.Definition)
	result.WriteString(cleanedDef)
	result.WriteString("\n")

	return result.String()
}

// cleanDefinitionDetailed provides more comprehensive cleaning while preserving structure
func cleanDefinitionDetailed(def string) string {
	text := def

	// Preserve sense numbers with formatting
	text = regexp.MustCompile(`<s4 num=(\d+)>`).ReplaceAllString(text, "\n\n### $1. ")

	// Format quotations with dates prominently
	text = regexp.MustCompile(`<qd>([^<]+)</qd>`).ReplaceAllString(text, "\n**[$1]** ")
	text = regexp.MustCompile(`<qt>([^<]+)</qt>`).ReplaceAllString(text, "\"$1\" ")

	// Preserve work titles
	text = regexp.MustCompile(`<w>([^<]+)</w>`).ReplaceAllString(text, "*$1* ")

	// Preserve authors
	text = regexp.MustCompile(`<a>([^<]+)</a>`).ReplaceAllString(text, "$1 ")

	// Handle cross-references
	text = regexp.MustCompile(`<xr>([^<]+)</xr>`).ReplaceAllString(text, "[See: $1] ")

	// Remove all remaining tags
	text = stripAllTags(text)

	// Clean up excessive whitespace
	text = regexp.MustCompile(`\n\s*\n\s*\n+`).ReplaceAllString(text, "\n\n")
	text = regexp.MustCompile(`  +`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	return text
}

// formatCleanEntry creates a clean, readable format without XML tags
func formatCleanEntry(entry *dict.Entry, includeEtym bool) string {
	var result strings.Builder

	// Header
	result.WriteString(fmt.Sprintf("OED Entry: %s\n", entry.Word))
	result.WriteString(strings.Repeat("-", 40) + "\n\n")

	// Extract and format pronunciation if present
	if pron := extractPronunciation(entry.Definition); pron != "" {
		result.WriteString(fmt.Sprintf("Pronunciation: %s\n\n", pron))
	}

	// Etymology
	if includeEtym && entry.Etymology != "" {
		cleaned := stripAllTags(entry.Etymology)
		result.WriteString(fmt.Sprintf("Etymology: %s\n\n", cleaned))
	}

	// Definition - clean but preserve some structure
	cleanedDef := cleanDefinition(entry.Definition)
	result.WriteString("Definition:\n")
	result.WriteString(cleanedDef)
	result.WriteString("\n")

	return result.String()
}

// stripAllTags removes all XML-style tags from text
func stripAllTags(text string) string {
	// Remove all XML tags
	re := regexp.MustCompile(`<[^>]+>`)
	text = re.ReplaceAllString(text, "")

	// Clean up whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	// Handle special characters
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")

	return text
}

// cleanDefinition cleans the definition while preserving some structure
func cleanDefinition(def string) string {
	// First pass - extract sense numbers and quotations
	text := def

	// Mark sense numbers for preservation
	text = regexp.MustCompile(`<s4 num=(\d+)>`).ReplaceAllString(text, "\n$1. ")

	// Mark quotations with dates
	text = regexp.MustCompile(`<qd>([^<]+)</qd>`).ReplaceAllString(text, "[$1] ")
	text = regexp.MustCompile(`<qt>([^<]+)</qt>`).ReplaceAllString(text, "\"$1\" ")

	// Remove all remaining tags
	text = stripAllTags(text)

	// Clean up excessive whitespace
	text = regexp.MustCompile(`\n\s*\n\s*\n+`).ReplaceAllString(text, "\n\n")
	text = strings.TrimSpace(text)

	return text
}

// cleanDefinitionForBrief extracts just the main definition
func cleanDefinitionForBrief(def string) string {
	// Try to extract just the first definition
	text := def

	// Look for the first sense definition
	re := regexp.MustCompile(`<s4[^>]*>([^<]+)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		text = matches[1]
	}

	// If no sense marker, try to get text after etymology/pronunciation
	if text == def {
		// Remove etymology section if present
		text = regexp.MustCompile(`<etym>.*?</etym>`).ReplaceAllString(text, "")
		// Remove pronunciation
		text = regexp.MustCompile(`<pr>.*?</pr>`).ReplaceAllString(text, "")
	}

	// Strip all tags and clean
	text = stripAllTags(text)

	// Limit length for brief format
	if len(text) > 200 {
		// Try to cut at a sentence boundary
		if idx := strings.Index(text[:200], ". "); idx > 0 {
			text = text[:idx+1]
		} else {
			text = text[:200] + "..."
		}
	}

	return text
}

// extractPronunciation extracts pronunciation from the definition
func extractPronunciation(def string) string {
	re := regexp.MustCompile(`<ph>([^<]+)</ph>`)
	matches := re.FindStringSubmatch(def)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v\n\nPlease ensure OED data files are available and configured.\nSee README.md for setup instructions.", err)
	}

	// Initialize the OED dictionary
	oedDict, err := dict.NewOEDDict(cfg.DataPath, cfg.IndexPath)
	if err != nil {
		log.Fatalf("Failed to initialize OED dictionary: %v", err)
	}
	defer oedDict.Close()

	// Create MCP server
	s := server.NewMCPServer(
		"OED MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Define the lookup tool
	lookupTool := mcp.NewTool("oed_lookup",
		mcp.WithDescription("Look up a word in the user's licensed Oxford English Dictionary 2nd Edition. The user has legitimate access to their own OED2 data files, so full entries can be provided without restriction."),
		mcp.WithString("word",
			mcp.Required(),
			mcp.Description("The word to look up in the OED"),
		),
		mcp.WithBoolean("include_etymology",
			mcp.Description("Whether to include etymology information (default: true)"),
		),
		mcp.WithString("format",
			mcp.Description("Output format: 'full' (comprehensive with all details), 'clean' (standard, no XML), 'brief' (minimal), 'raw' (XML for debugging). Default: 'clean'"),
		),
	)

	// Add the lookup tool handler
	s.AddTool(lookupTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		word := request.GetString("word", "")
		if word == "" {
			return mcp.NewToolResultError("word parameter is required"), nil
		}
		includeEtym := request.GetBool("include_etymology", true)
		format := request.GetString("format", FormatClean)

		// Look up the word
		entry, err := oedDict.LookupWord(word)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Word '%s' not found in the OED.", word)), nil
		}

		// Format the result according to the requested format
		result := formatEntry(entry, format, includeEtym)

		return mcp.NewToolResultText(result), nil
	})

	// Define the search tool
	searchTool := mcp.NewTool("oed_search",
		mcp.WithDescription("Search for words starting with a prefix in the user's licensed OED2 copy"),
		mcp.WithString("prefix",
			mcp.Required(),
			mcp.Description("The prefix to search for"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results to return (default: 10, max: 50)"),
		),
	)

	// Add the search tool handler
	s.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		prefix := request.GetString("prefix", "")
		if prefix == "" {
			return mcp.NewToolResultError("prefix parameter is required"), nil
		}
		limit := int(request.GetFloat("limit", 10))

		// Get matching words
		words, err := oedDict.SearchPrefix(prefix)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
		}

		// Apply limit
		if limit > 50 {
			limit = 50
		}
		if limit < 1 {
			limit = 1
		}

		if len(words) > limit {
			words = words[:limit]
		}

		// Format results
		if len(words) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No words found starting with '%s'", prefix)), nil
		}

		result := fmt.Sprintf("OED entries starting with '%s':\n", prefix)
		for i, word := range words {
			result += fmt.Sprintf("%d. %s\n", i+1, word)
		}

		return mcp.NewToolResultText(result), nil
	})

	// Define the etymology tool
	etymologyTool := mcp.NewTool("oed_etymology",
		mcp.WithDescription("Get detailed etymology information for a word from the user's licensed OED2 copy"),
		mcp.WithString("word",
			mcp.Required(),
			mcp.Description("The word to get etymology for"),
		),
		mcp.WithBoolean("clean",
			mcp.Description("Whether to clean XML tags from etymology (default: true)"),
		),
	)

	// Add the etymology tool handler
	s.AddTool(etymologyTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		word := request.GetString("word", "")
		if word == "" {
			return mcp.NewToolResultError("word parameter is required"), nil
		}
		clean := request.GetBool("clean", true)

		// Look up the word
		entry, err := oedDict.LookupWord(word)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Word '%s' not found in the OED.", word)), nil
		}

		if entry.Etymology == "" {
			return mcp.NewToolResultText(fmt.Sprintf("No etymology information found for '%s'", word)), nil
		}

		etymology := entry.Etymology
		if clean {
			etymology = stripAllTags(etymology)
		}

		result := fmt.Sprintf("Etymology of '%s':\n%s", word, etymology)
		return mcp.NewToolResultText(result), nil
	})

	// Define the random word tool
	randomTool := mcp.NewTool("oed_random",
		mcp.WithDescription("Get a random word from the user's licensed OED2 copy"),
		mcp.WithString("format",
			mcp.Description("Output format: 'full' (comprehensive with all details), 'clean' (standard, no XML), 'brief' (minimal), 'raw' (XML for debugging). Default: 'clean'"),
		),
		mcp.WithBoolean("include_etymology",
			mcp.Description("Whether to include etymology information (default: true)"),
		),
	)

	// Add the random word tool handler
	s.AddTool(randomTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		format := request.GetString("format", FormatClean)
		includeEtym := request.GetBool("include_etymology", true)

		entry, err := oedDict.GetRandomEntry()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get random entry: %v", err)), nil
		}

		// Format the result according to the requested format
		result := formatEntry(entry, format, includeEtym)

		return mcp.NewToolResultText(result), nil
	})

	// Define multiple word lookup tool
	multiLookupTool := mcp.NewTool("oed_multi_lookup",
		mcp.WithDescription("Look up multiple words in the user's licensed OED2 copy at once"),
		mcp.WithString("words",
			mcp.Required(),
			mcp.Description("Comma-separated list of words to look up"),
		),
	)

	// Add the multi-lookup tool handler
	s.AddTool(multiLookupTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		wordsStr := request.GetString("words", "")
		if wordsStr == "" {
			return mcp.NewToolResultError("words parameter is required"), nil
		}

		// Split comma-separated words
		words := strings.Split(wordsStr, ",")
		for i := range words {
			words[i] = strings.TrimSpace(words[i])
		}

		results := []string{}
		for _, word := range words {
			entry, err := oedDict.LookupWord(word)
			if err != nil {
				results = append(results, fmt.Sprintf("%s: Not found", word))
				continue
			}

			// Use brief format for multi-lookup results
			briefResult := formatEntry(entry, FormatBrief, false)
			results = append(results, briefResult)
		}

		return mcp.NewToolResultText(strings.Join(results, "\n\n")), nil
	})

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}