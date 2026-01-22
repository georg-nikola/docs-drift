package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// CodeBlock represents an extracted code block from markdown
type CodeBlock struct {
	Language   string
	Code       string
	FilePath   string
	LineNumber int
	Skip       bool
}

// Parser extracts code blocks from markdown files
type Parser struct {
	// fenceRegex matches opening code fences with optional language
	fenceRegex *regexp.Regexp
}

// New creates a new markdown parser
func New() *Parser {
	return &Parser{
		// Matches ``` or ~~~ optionally followed by language and info string
		// Supports docs-drift:skip directive in info string
		fenceRegex: regexp.MustCompile(`^(\x60{3,}|~{3,})(\w+)?(.*)$`),
	}
}

// ParseFile extracts code blocks from a markdown file
func (p *Parser) ParseFile(filePath string) ([]CodeBlock, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var blocks []CodeBlock
	scanner := bufio.NewScanner(file)

	lineNum := 0
	var currentBlock *CodeBlock
	var fenceChar string
	var fenceLen int
	var codeLines []string

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if currentBlock == nil {
			// Look for opening fence
			matches := p.fenceRegex.FindStringSubmatch(line)
			if matches != nil {
				fence := matches[1]
				lang := strings.TrimSpace(matches[2])
				infoString := strings.TrimSpace(matches[3])

				// Check for skip directive
				skip := strings.Contains(infoString, "docs-drift:skip")

				currentBlock = &CodeBlock{
					Language:   lang,
					FilePath:   filePath,
					LineNumber: lineNum,
					Skip:       skip,
				}
				fenceChar = string(fence[0])
				fenceLen = len(fence)
				codeLines = nil
			}
		} else {
			// Look for closing fence
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, fenceChar) &&
				len(trimmed) >= fenceLen &&
				strings.Trim(trimmed, fenceChar) == "" {
				// Found closing fence
				currentBlock.Code = strings.Join(codeLines, "\n")
				blocks = append(blocks, *currentBlock)
				currentBlock = nil
				fenceChar = ""
				fenceLen = 0
			} else {
				codeLines = append(codeLines, line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Handle unclosed fence (shouldn't happen in valid markdown)
	if currentBlock != nil {
		currentBlock.Code = strings.Join(codeLines, "\n")
		blocks = append(blocks, *currentBlock)
	}

	return blocks, nil
}

// FilterByLanguages returns only blocks with specified languages
func FilterByLanguages(blocks []CodeBlock, languages []string) []CodeBlock {
	if len(languages) == 0 {
		return blocks
	}

	langSet := make(map[string]bool)
	for _, lang := range languages {
		langSet[normalizeLanguage(lang)] = true
	}

	var filtered []CodeBlock
	for _, block := range blocks {
		if langSet[normalizeLanguage(block.Language)] {
			filtered = append(filtered, block)
		}
	}

	return filtered
}

// FilterSkipped removes blocks marked with docs-drift:skip
func FilterSkipped(blocks []CodeBlock) []CodeBlock {
	var filtered []CodeBlock
	for _, block := range blocks {
		if !block.Skip {
			filtered = append(filtered, block)
		}
	}
	return filtered
}

func normalizeLanguage(lang string) string {
	switch strings.ToLower(lang) {
	case "js", "javascript":
		return "javascript"
	case "py", "python":
		return "python"
	default:
		return strings.ToLower(lang)
	}
}
