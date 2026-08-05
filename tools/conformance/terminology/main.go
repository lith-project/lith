package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// Violation records an undefined term found in a scanned file.
type Violation struct {
	File string
	Line int
	Term string
}

// loadGlossary reads docs/glossary.md and extracts level-3 headings as canonical terms.
// Parenthetical qualifiers are stripped — "### Note (Document)" → "Note".
func loadGlossary(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening glossary: %w", err)
	}
	defer func() { _ = f.Close() }()

	terms := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "### ") {
			term := strings.TrimSpace(strings.TrimPrefix(line, "###"))
			if idx := strings.Index(term, "("); idx >= 0 {
				term = strings.TrimSpace(term[:idx])
			}
			if term != "" {
				terms[term] = true
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading glossary: %w", err)
	}
	return terms, nil
}

// loadAllowlist reads the allowlist file and returns the set of allowed terms.
func loadAllowlist(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening allowlist: %w", err)
	}
	defer func() { _ = f.Close() }()

	terms := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		term := line
		if idx := strings.Index(line, "#"); idx >= 0 {
			term = strings.TrimSpace(line[:idx])
		}
		if term != "" {
			terms[term] = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading allowlist: %w", err)
	}
	return terms, nil
}

// validWordRe matches a single capitalized word: [A-Z][a-z]+
var validWordRe = regexp.MustCompile(`^[A-Z][a-z]+$`)

// stripMarkdown removes common markdown formatting for candidate extraction.
func stripMarkdown(line string) string {
	// Remove bold: **text** → text
	line = boldRe.ReplaceAllString(line, "$1")
	// Remove italic: *text* → text
	line = italicRe.ReplaceAllString(line, "$1")
	// Remove link destinations: [text](URL) → text
	line = linkDestRe.ReplaceAllString(line, "$1")
	// Remove inline code: `text` → text
	line = inlineCodeRe.ReplaceAllString(line, "$1")
	return line
}

var (
	boldRe       = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	italicRe     = regexp.MustCompile(`\*([^*]+)\*`)
	linkDestRe   = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	inlineCodeRe = regexp.MustCompile("`[^`]+`")
)

// isSentenceBoundaryWord checks if a word at position i in the rune slice
// is the first content word on the line (accounting for markdown list prefixes,
// blockquote prefixes, digits, and punctuation).
func isSentenceBoundaryWord(runes []rune, i int) bool {
	// Check that everything before position i is only:
	// whitespace, digits, periods, hyphens, asterisks, plus signs, pipes, colons
	for k := 0; k < i; k++ {
		ch := runes[k]
		if unicode.IsSpace(ch) {
			continue
		}
		if unicode.IsDigit(ch) || ch == '.' || ch == '-' || ch == '*' ||
			ch == '+' || ch == '|' || ch == ':' || ch == '>' ||
			ch == '[' || ch == ']' || ch == '(' || ch == ')' {
			continue
		}
		return false
	}
	return true
}

// extractCandidates extracts candidate terms from a line of prose.
// Multi-word phrases of consecutive capitalized words are treated as single terms.
// A single capitalized word at sentence boundary is skipped.
func extractCandidates(line string) []string {
	cleaned := stripMarkdown(line)
	runes := []rune(cleaned)

	// Find content start (first non-space position)
	contentStart := 0
	for contentStart < len(runes) && unicode.IsSpace(runes[contentStart]) {
		contentStart++
	}

	var candidates []string
	i := contentStart

	for i < len(runes) {
		if !unicode.IsUpper(runes[i]) {
			i++
			continue
		}

		// Find end of this word
		j := i + 1
		for j < len(runes) && unicode.IsLetter(runes[j]) {
			j++
		}
		word := string(runes[i:j])

		if !validWordRe.MatchString(word) {
			i = j
			continue
		}

		// Skip single word at sentence boundary BEFORE building phrase
		if isSentenceBoundaryWord(runes, i) {
			i = j
			continue
		}

		// Try to build a phrase of consecutive capitalized words separated by single spaces
		phraseWords := []string{word}
		probe := j
		for probe < len(runes) {
			wsCount := 0
			tmp := probe
			for tmp < len(runes) && unicode.IsSpace(runes[tmp]) {
				wsCount++
				tmp++
			}
			if tmp == probe || tmp >= len(runes) {
				break
			}
			if wsCount != 1 {
				break
			}
			if !unicode.IsUpper(runes[tmp]) {
				break
			}
			nextStart := tmp
			tmp++
			for tmp < len(runes) && unicode.IsLetter(runes[tmp]) {
				tmp++
			}
			nextWord := string(runes[nextStart:tmp])
			if !validWordRe.MatchString(nextWord) {
				break
			}
			phraseWords = append(phraseWords, nextWord)
			probe = tmp
		}

		if len(phraseWords) >= 2 {
			candidates = append(candidates, strings.Join(phraseWords, " "))
			i = probe
			continue
		}

		// Single word: skip if at sentence boundary
		if isSentenceBoundaryWord(runes, i) {
			i = j
			continue
		}

		candidates = append(candidates, word)
		i = j
	}

	return candidates
}

// stripDeterminers removes leading English determiners from a phrase.
func stripDeterminers(phrase string) string {
	for _, det := range []string{"A ", "An ", "The "} {
		if strings.HasPrefix(phrase, det) {
			stripped := strings.TrimSpace(phrase[len(det):])
			if stripped != "" {
				return stripped
			}
		}
	}
	return phrase
}

// scanFile scans a single file for terminology violations.
func scanFile(filePath string, glossary, allowlist map[string]bool) ([]Violation, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", filePath, err)
	}
	defer func() { _ = f.Close() }()

	var violations []Violation
	lineNum := 0
	inFrontmatter := false
	frontmatterCount := 0
	inCodeBlock := false
	inReferences := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Track YAML frontmatter (between opening --- and closing ---)
		if trimmed == "---" {
			frontmatterCount++
			if frontmatterCount == 1 {
				inFrontmatter = true
				continue
			}
			if frontmatterCount == 2 {
				inFrontmatter = false
				continue
			}
		}

		if inFrontmatter {
			continue
		}

		// Track code fences
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			continue
		}

		// References section and everything after — skip (BEFORE heading check!)
		if strings.HasPrefix(trimmed, "## References") || strings.HasPrefix(trimmed, "## Appendix") {
			inReferences = true
			continue
		}
		if inReferences {
			continue
		}

		// Headings — skip
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Table rows — skip
		if strings.HasPrefix(trimmed, "|") {
			continue
		}

		// Empty lines — skip
		if trimmed == "" {
			continue
		}

		// Extract candidates from prose
		candidates := extractCandidates(line)

		// Deduplicate within line
		seen := make(map[string]bool)
		for _, candidate := range candidates {
			if seen[candidate] {
				continue
			}
			seen[candidate] = true

			lookup := stripDeterminers(candidate)

			if glossary[lookup] || allowlist[lookup] || allowlist[candidate] {
				continue
			}

			violations = append(violations, Violation{
				File: filePath,
				Line: lineNum,
				Term: candidate,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}

	return violations, nil
}

func main() {
	repoRoot, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: determining working directory: %v\n", err)
		os.Exit(2)
	}

	glossaryPath := filepath.Join(repoRoot, "docs", "glossary.md")
	glossary, err := loadGlossary(glossaryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: loading glossary: %v\n", err)
		os.Exit(2)
	}

	allowlistPath := filepath.Join(repoRoot, "tools", "conformance", "terminology-allowlist.txt")
	allowlist, err := loadAllowlist(allowlistPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: loading allowlist: %v\n", err)
		os.Exit(2)
	}

	rfcDir := filepath.Join(repoRoot, "rfcs")
	entries, err := os.ReadDir(rfcDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: reading rfcs directory: %v\n", err)
		os.Exit(2)
	}

	exclude := map[string]bool{
		"index.md":  true,
		"README.md": true,
	}

	var allViolations []Violation
	for _, entry := range entries {
		if entry.IsDir() || exclude[entry.Name()] || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		rfcPath := filepath.Join(rfcDir, entry.Name())
		violations, err := scanFile(rfcPath, glossary, allowlist)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: scanning %s: %v\n", rfcPath, err)
			os.Exit(2)
		}
		allViolations = append(allViolations, violations...)
	}

	for _, v := range allViolations {
		fmt.Fprintf(os.Stderr, "%s:%d: undefined term %q\n", v.File, v.Line, v.Term)
	}

	if len(allViolations) > 0 {
		os.Exit(1)
	}
}
