package drift

import (
	"context"
	"fmt"
	"sync"

	"github.com/georg-nikola/docs-drift/internal/config"
	"github.com/georg-nikola/docs-drift/internal/parser"
	"github.com/georg-nikola/docs-drift/internal/runner"
)

// Results contains the outcome of checking all code blocks
type Results struct {
	TotalFiles  int
	TotalBlocks int
	Passed      int
	Skipped     int
	Failed      []BlockResult
}

// BlockResult represents the result of checking a single code block
type BlockResult struct {
	FilePath   string
	LineNumber int
	Language   string
	Success    bool
	Output     string
	Error      string
}

// HasDrift returns true if any code blocks failed
func (r *Results) HasDrift() bool {
	return len(r.Failed) > 0
}

// Checker validates code blocks in markdown files
type Checker struct {
	config  *config.Config
	parser  *parser.Parser
	verbose bool
}

// NewChecker creates a new drift checker
func NewChecker(cfg *config.Config, verbose bool) *Checker {
	return &Checker{
		config:  cfg,
		parser:  parser.New(),
		verbose: verbose,
	}
}

// Check validates code blocks in the given files
func (c *Checker) Check(files []string) (*Results, error) {
	results := &Results{
		TotalFiles: len(files),
	}

	// Verify runtimes are available
	for _, lang := range c.config.Checks.CodeBlocks.Languages {
		if err := runner.CheckRuntime(lang); err != nil {
			return nil, fmt.Errorf("runtime check failed for %s: %w", lang, err)
		}
	}

	// Process each file
	for _, filePath := range files {
		if c.verbose {
			fmt.Printf("Checking %s...\n", filePath)
		}

		blocks, err := c.parser.ParseFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
		}

		// Filter blocks
		blocks = parser.FilterByLanguages(blocks, c.config.Checks.CodeBlocks.Languages)

		for _, block := range blocks {
			results.TotalBlocks++

			if block.Skip {
				results.Skipped++
				if c.verbose {
					fmt.Printf("  Skipped block at line %d (docs-drift:skip)\n", block.LineNumber)
				}
				continue
			}

			// Skip blocks without recognized language
			if block.Language == "" {
				results.Skipped++
				continue
			}

			// Run the code block
			result := c.runBlock(block)

			if result.Success {
				results.Passed++
				if c.verbose {
					fmt.Printf("  Line %d: %s - passed\n", block.LineNumber, block.Language)
				}
			} else {
				results.Failed = append(results.Failed, result)
				if c.verbose {
					fmt.Printf("  Line %d: %s - FAILED\n", block.LineNumber, block.Language)
				}
			}
		}
	}

	return results, nil
}

// runBlock executes a single code block and returns the result
func (c *Checker) runBlock(block parser.CodeBlock) BlockResult {
	result := BlockResult{
		FilePath:   block.FilePath,
		LineNumber: block.LineNumber,
		Language:   block.Language,
	}

	r, err := runner.NewRunner(block.Language, c.config.Checks.CodeBlocks.Timeout)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}

	ctx := context.Background()
	runResult := r.Run(ctx, block.Code)

	result.Success = runResult.Success
	result.Output = runResult.Output
	result.Error = runResult.Error

	return result
}

// CheckConcurrent validates code blocks using multiple goroutines
// This is an optimization for when there are many code blocks
func (c *Checker) CheckConcurrent(files []string, workers int) (*Results, error) {
	if workers <= 0 {
		workers = 4
	}

	results := &Results{
		TotalFiles: len(files),
	}

	// Collect all blocks first
	var allBlocks []parser.CodeBlock
	for _, filePath := range files {
		blocks, err := c.parser.ParseFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
		}
		blocks = parser.FilterByLanguages(blocks, c.config.Checks.CodeBlocks.Languages)
		allBlocks = append(allBlocks, blocks...)
	}

	results.TotalBlocks = len(allBlocks)

	// Process blocks concurrently
	type blockJob struct {
		block parser.CodeBlock
		index int
	}

	jobs := make(chan blockJob, len(allBlocks))
	resultsChan := make(chan BlockResult, len(allBlocks))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				if job.block.Skip || job.block.Language == "" {
					resultsChan <- BlockResult{
						FilePath:   job.block.FilePath,
						LineNumber: job.block.LineNumber,
						Language:   job.block.Language,
						Success:    true, // Skipped blocks are considered successful
					}
					continue
				}
				resultsChan <- c.runBlock(job.block)
			}
		}()
	}

	// Send jobs
	for i, block := range allBlocks {
		jobs <- blockJob{block: block, index: i}
	}
	close(jobs)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		if result.Success {
			// Check if it was skipped
			for _, block := range allBlocks {
				if block.FilePath == result.FilePath &&
					block.LineNumber == result.LineNumber &&
					(block.Skip || block.Language == "") {
					results.Skipped++
					continue
				}
			}
			results.Passed++
		} else {
			results.Failed = append(results.Failed, result)
		}
	}

	return results, nil
}
