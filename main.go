package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pdfinn/oedmcp/config"
	"github.com/pdfinn/oedmcp/dict"
)

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
	)

	// Add the lookup tool handler
	s.AddTool(lookupTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		word := request.GetString("word", "")
		if word == "" {
			return mcp.NewToolResultError("word parameter is required"), nil
		}
		includeEtym := request.GetBool("include_etymology", true)

		// Look up the word
		entry, err := oedDict.LookupWord(word)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Word '%s' not found in the OED.", word)), nil
		}

		// Format the result
		result := fmt.Sprintf("OED Entry for '%s':\n\n", word)

		if entry.Definition != "" {
			result += fmt.Sprintf("Definition:\n%s\n", entry.Definition)
		}

		if includeEtym && entry.Etymology != "" {
			result += fmt.Sprintf("\nEtymology:\n%s\n", entry.Etymology)
		}

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
	)

	// Add the etymology tool handler
	s.AddTool(etymologyTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		word := request.GetString("word", "")
		if word == "" {
			return mcp.NewToolResultError("word parameter is required"), nil
		}

		// Look up the word
		entry, err := oedDict.LookupWord(word)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Word '%s' not found in the OED.", word)), nil
		}

		if entry.Etymology == "" {
			return mcp.NewToolResultText(fmt.Sprintf("No etymology information found for '%s'", word)), nil
		}

		result := fmt.Sprintf("Etymology of '%s':\n%s", word, entry.Etymology)
		return mcp.NewToolResultText(result), nil
	})

	// Define the random word tool
	randomTool := mcp.NewTool("oed_random",
		mcp.WithDescription("Get a random word from the user's licensed OED2 copy"),
	)

	// Add the random word tool handler
	s.AddTool(randomTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		entry, err := oedDict.GetRandomEntry()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get random entry: %v", err)), nil
		}

		result := fmt.Sprintf("Random OED Entry:\n\n")
		if entry.Word != "" {
			result += fmt.Sprintf("Word: %s\n\n", entry.Word)
		}
		if entry.Definition != "" {
			result += fmt.Sprintf("Definition:\n%s\n", entry.Definition)
		}
		if entry.Etymology != "" {
			result += fmt.Sprintf("\nEtymology:\n%s\n", entry.Etymology)
		}

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
				results = append(results, fmt.Sprintf("'%s': Not found", word))
				continue
			}

			wordResult := fmt.Sprintf("'%s':", word)
			if entry.Definition != "" {
				// Truncate long definitions
				def := entry.Definition
				if len(def) > 200 {
					def = def[:200] + "..."
				}
				wordResult += fmt.Sprintf(" %s", def)
			}
			results = append(results, wordResult)
		}

		return mcp.NewToolResultText(strings.Join(results, "\n\n")), nil
	})

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}