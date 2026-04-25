package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"unicode"
)

// ==========================================================
// Diff 过滤与压缩逻辑
// ==========================================================

// getStagedChanges 获取暂存区的变更并进行处理
// 返回:
//   - 处理后的变更内容
//   - 错误信息
func getStagedChanges() (string, error) {
	statSummary := getStatSummary()

	cleanDiff, err := getFilteredDiff()
	if err != nil {
		return "", fmt.Errorf("获取过滤后的Diff失败: %v", err)
	}

	compressedDiff := compressDiffUniversal(cleanDiff, 1, 10)

	var sb strings.Builder
	sb.Grow(len(statSummary) + len(compressedDiff) + 100) // 预分配缓冲区

	if statSummary != "" {
		sb.WriteString("【变更统计】\n")
		sb.WriteString(statSummary)
		sb.WriteString("\n\n")
	}

	sb.WriteString("【详细 Diff 摘要】\n")
	sb.WriteString(compressedDiff)

	content := sb.String()
	maxChars := 4000
	if len(content) > maxChars {
		content = content[:maxChars] + "\n... [内容过长已截断，请根据现有信息推断]"
	}

	return content, nil
}

// getStatSummary 获取变更统计信息
// 返回:
//   - 变更统计信息字符串
func getStatSummary() string {
	cmd := exec.Command("git", "diff", "--cached", "--stat")
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取变更统计失败: %+v\n", err)
		return ""
	}
	return string(out)
}

// getFilteredDiff 获取过滤后的Diff内容
// 返回:
//   - 过滤后的Diff内容
//   - 错误信息
func getFilteredDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-status")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取文件状态失败: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	var keptFiles []string
	var ignoredFilesBuilder strings.Builder
	ignoredFilesBuilder.Grow(1000) // 预分配缓冲区

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			continue
		}
		status, path := parts[0], parts[1]

		if IsIgnored(path) {
			if ignoredFilesBuilder.Len() > 0 {
				ignoredFilesBuilder.WriteString("\n")
			}
			ignoredFilesBuilder.WriteString("[")
			ignoredFilesBuilder.WriteString(status)
			ignoredFilesBuilder.WriteString("] ")
			ignoredFilesBuilder.WriteString(path)
			ignoredFilesBuilder.WriteString(" (已过滤)")
		} else {
			keptFiles = append(keptFiles, path)
		}
	}

	var sb strings.Builder
	sb.Grow(2000 + ignoredFilesBuilder.Len()) // 预分配缓冲区

	if ignoredFilesBuilder.Len() > 0 {
		sb.WriteString("【变更说明：以下文件已变更但详细内容已过滤（锁文件/二进制等）】\n")
		sb.WriteString(ignoredFilesBuilder.String())
		sb.WriteString("\n\n")
	}

	if len(keptFiles) > 0 {
		args := append([]string{"diff", "--cached", "--"}, keptFiles...)
		cmdDiff := exec.Command("git", args...)
		diffOut, err := cmdDiff.Output()
		if err != nil {
			return "", fmt.Errorf("获取Diff内容失败: %v", err)
		}
		sb.Write(diffOut)
	}

	return sb.String(), nil
}

var structuralSymbolMap = map[rune]struct{}{
	'{': {}, '}': {}, '(': {}, ')': {},
	':': {}, '=': {}, '<': {}, '>': {},
}

var importantPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^[-+]?func\s+`),
	regexp.MustCompile(`^[-+]?type\s+\w+\s+struct`),
	regexp.MustCompile(`^[-+]?type\s+\w+\s+interface`),
	regexp.MustCompile(`^[-+]?const\s+`),
	regexp.MustCompile(`^[-+]?var\s+`),
	regexp.MustCompile(`^[-+]?(public|private|protected|internal)\s+`),
	regexp.MustCompile(`^[-+]?class\s+`),
	regexp.MustCompile(`^[-+]?def\s+\w+\s*\(`),
	regexp.MustCompile(`^[-+]?async\s+func`),
	regexp.MustCompile(`^[-+]?import\s+`),
	regexp.MustCompile(`^[-+]?from\s+\w+\s+import`),
	regexp.MustCompile(`^[-+]?package\s+`),
	regexp.MustCompile(`^[-+]?namespace\s+`),
	regexp.MustCompile(`^\s*#.*import\s+`),
	regexp.MustCompile(`^\s*@\w+\s*(\(|$)`),
	regexp.MustCompile(`^\s*//.*FIXME|//.*TODO|//.*BUG`),
}

var importantSymbols = map[rune]int{
	'{': 2, '}': 2, '(': 1, ')': 1,
	'[': 1, ']': 1,
}

func matchesImportantPattern(line string) bool {
	stripped := strings.TrimPrefix(line, "+")
	stripped = strings.TrimPrefix(stripped, "-")
	for _, pattern := range importantPatterns {
		if pattern.MatchString(stripped) {
			return true
		}
	}
	return false
}

func calculateHunkImportance(lines []string, start, end int) int {
	importance := 0
	for i := start; i < end && i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			if matchesImportantPattern(line) {
				importance += 5
			}
			score := calculateLineScore(line)
			if score >= 3 {
				importance += score
			}
			symbolCount := countImportantSymbols(line)
			importance += symbolCount
		}
	}
	return importance
}

func countImportantSymbols(line string) int {
	startIdx := 0
	if len(line) > 0 && (line[0] == '+' || line[0] == '-') {
		startIdx = 1
	}
	count := 0
	for i := startIdx; i < len(line); i++ {
		if weight, ok := importantSymbols[rune(line[i])]; ok {
			count += weight
		}
	}
	return count
}

// compressDiffUniversal 通用Diff压缩算法
// 功能说明：
//  1. 将Diff内容按行分割
//  2. 提取并分析每个代码块（hunk）的重要性
//  3. 根据代码重要性智能过滤内容：
//     - 头部信息（diff --git, index, ---, +++）完整保留
//     - 代码块标记（@@）根据重要性决定是否添加说明
//     - 上下文行（未变更行）限制在 maxContext 行内
//     - 变更行（+/-开头）根据重要性分数决定保留：
//     * 匹配重要模式（函数声明、类型定义等）始终保留
//     * 重要性分数 >= 3 的行始终保留
//     * 其他变更行在不超过 maxDetailLines 限制时保留
//
// 参数说明：
//   - diff: 原始Diff内容
//   - maxContext: 每个代码块保留的最大上下文行数（未变更行）
//   - maxDetailLines: 每个代码块保留的最大变更行数（+/-开头的行）
//
// 返回值：压缩后的Diff内容字符串
func compressDiffUniversal(diff string, maxContext int, maxDetailLines int) string {
	lines := strings.Split(diff, "\n")
	var sb strings.Builder
	sb.Grow(len(diff) / 2)

	hunks := extractHunks(lines)
	hunkStats := make([]hunkStat, len(hunks))

	for i, hunk := range hunks {
		hunkStats[i] = analyzeHunk(lines, hunk)
	}

	for i, line := range lines {
		if isHeaderLine(line) {
			if i > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString(line)
			continue
		}

		if strings.HasPrefix(line, "@@ ") {
			hunkIdx := findHunkIndexForLine(i, hunks)
			shouldTruncate := false
			if hunkIdx >= 0 && hunkIdx < len(hunkStats) {
				shouldTruncate = !hunkStats[hunkIdx].keepFull
			}
			if shouldTruncate && hunkIdx >= 0 && hunkIdx < len(hunkStats) {
				sb.WriteString("... [代码块重要性高，保留更多细节]\n")
			}
			sb.WriteString(line)
			if i < len(lines)-1 {
				sb.WriteByte('\n')
			}
			continue
		}

		hunkIdx := findHunkIndexForLine(i, hunks)
		shouldTruncate := false
		detailCount := 0
		if hunkIdx >= 0 && hunkIdx < len(hunkStats) {
			shouldTruncate = !hunkStats[hunkIdx].keepFull
			detailCount = hunkStats[hunkIdx].detailCount
		}

		if !strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "-") {
			contextLines := hunkStats[hunkIdx].contextCount
			if contextLines < maxContext {
				sb.WriteString(line)
				if i < len(lines)-1 {
					sb.WriteByte('\n')
				}
				if hunkIdx >= 0 && hunkIdx < len(hunkStats) {
					hunkStats[hunkIdx].contextCount++
				}
			} else if contextLines == maxContext {
				sb.WriteString("... [上下文已省略]")
				if i < len(lines)-1 {
					sb.WriteByte('\n')
				}
			}
			continue
		}

		if matchesImportantPattern(line) {
			sb.WriteString(line)
			if i < len(lines)-1 {
				sb.WriteByte('\n')
			}
			if hunkIdx >= 0 && hunkIdx < len(hunkStats) {
				hunkStats[hunkIdx].detailCount++
			}
			continue
		}

		score := calculateLineScore(line)
		if score >= 3 {
			sb.WriteString(line)
			if i < len(lines)-1 {
				sb.WriteByte('\n')
			}
			if hunkIdx >= 0 && hunkIdx < len(hunkStats) {
				hunkStats[hunkIdx].detailCount++
			}
			continue
		}

		if !shouldTruncate || detailCount < maxDetailLines {
			sb.WriteString(line)
			if i < len(lines)-1 {
				sb.WriteByte('\n')
			}
			if hunkIdx >= 0 && hunkIdx < len(hunkStats) {
				hunkStats[hunkIdx].detailCount++
			}
		} else if detailCount == maxDetailLines {
			sb.WriteString("... [冗余细节已省略]")
			if i < len(lines)-1 {
				sb.WriteByte('\n')
			}
			if hunkIdx >= 0 && hunkIdx < len(hunkStats) {
				hunkStats[hunkIdx].detailCount++
			}
		}
	}

	return sb.String()
}

type hunkStat struct {
	keepFull     bool
	detailCount  int
	contextCount int
	importance   int
}

func extractHunks(lines []string) []hunkRange {
	var hunks []hunkRange
	for i, line := range lines {
		if strings.HasPrefix(line, "@@ ") {
			hunks = append(hunks, hunkRange{start: i})
		} else if len(hunks) > 0 && i > 0 && lines[i-1] == "" && !strings.HasPrefix(line, "@@ ") && !strings.HasPrefix(line, "diff ") && !strings.HasPrefix(line, "index ") && !strings.HasPrefix(line, "--- ") && !strings.HasPrefix(line, "+++ ") {
			hunks[len(hunks)-1].end = i - 1
		}
	}
	if len(hunks) > 0 {
		hunks[len(hunks)-1].end = len(lines)
	}
	return hunks
}

type hunkRange struct {
	start int
	end   int
}

func analyzeHunk(lines []string, hunk hunkRange) hunkStat {
	stats := hunkStat{
		keepFull:    false,
		detailCount: 0,
		importance:  0,
	}

	additions := 0
	deletions := 0
	for i := hunk.start; i < hunk.end && i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "+") {
			additions++
			if matchesImportantPattern(lines[i]) {
				stats.importance += 10
			}
		} else if strings.HasPrefix(lines[i], "-") {
			deletions++
		}
	}

	stats.importance = calculateHunkImportance(lines, hunk.start, hunk.end)

	balanceRatio := 0.0
	if additions+deletions > 0 {
		balanceRatio = float64(abs(additions-deletions)) / float64(additions+deletions)
	}

	if stats.importance > 15 || balanceRatio < 0.3 {
		stats.keepFull = true
	}

	return stats
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func isHeaderLine(line string) bool {
	return strings.HasPrefix(line, "diff --git ") ||
		strings.HasPrefix(line, "index ") ||
		strings.HasPrefix(line, "--- ") ||
		strings.HasPrefix(line, "+++")
}

func findHunkIndexForLine(lineIdx int, hunks []hunkRange) int {
	for i, hunk := range hunks {
		if lineIdx >= hunk.start && lineIdx < hunk.end {
			return i
		}
	}
	return -1
}

// calculateLineScore 计算代码行的重要性分数
// 参数:
//   - line: 代码行
//
// 返回:
//   - 分数，越高越重要
func calculateLineScore(line string) int {
	score := 0

	startIdx := 0
	if len(line) > 0 && (line[0] == '+' || line[0] == '-') {
		startIdx = 1
	}

	leadingSpaces := 0
	for i := startIdx; i < len(line); i++ {
		if unicode.IsSpace(rune(line[i])) {
			leadingSpaces++
		} else {
			break
		}
	}

	contentStart := startIdx + leadingSpaces
	contentLen := len(line) - contentStart

	if contentLen > 150 {
		score -= 2
	} else if contentLen > 80 {
		score -= 1
	}

	foundSymbols := 0
	for i := contentStart; i < len(line); i++ {
		if _, ok := structuralSymbolMap[rune(line[i])]; ok {
			foundSymbols++
		}
	}
	score += foundSymbols * 2

	if leadingSpaces == 0 {
		score += 3
	} else if leadingSpaces <= 4 {
		score += 1
	}

	if contentLen >= 2 {
		firstTwo := line[contentStart : contentStart+2]
		if firstTwo == "//" || firstTwo == "/*" {
			score -= 1
		}
	} else if contentLen >= 1 {
		if line[contentStart] == '#' {
			score -= 1
		}
	}

	score += calculateKeywordScore(line)
	score += calculateOperatorScore(line, contentStart)
	score += calculatePatternScore(line)

	return score
}

var languageKeywords = map[string]int{
	"func": 5, "function": 5, "def": 5, "fn": 5,
	"type": 4, "class": 4, "struct": 4, "interface": 4,
	"const": 3, "var": 3, "let": 3, "static": 3,
	"return": 4, "yield": 3, "throw": 3, "raise": 3,
	"if": 2, "else": 2, "switch": 2, "case": 2,
	"for": 2, "while": 2, "do": 2, "foreach": 2,
	"try": 3, "catch": 3, "finally": 3,
	"import": 4, "export": 4, "from": 3, "require": 3,
	"public": 3, "private": 3, "protected": 3, "internal": 3,
	"async": 4, "await": 3, "promise": 3,
	"new": 3, "delete": 3, "this": 2, "self": 2, "super": 2,
	"true": 1, "false": 1, "nil": 1, "null": 1, "none": 1,
	"get": 2, "set": 2,
}

var importantOperators = map[rune]int{
	'=': 2, '+': 1, '-': 1, '*': 1, '/': 1,
	'<': 1, '>': 1, '!': 1, '&': 2, '|': 2,
	':': 2, '.': 2, ',': 1,
}

type patternScore struct {
	pattern *regexp.Regexp
	score   int
}

var lineScoringPatterns = []patternScore{
	{regexp.MustCompile(`^\s*(public|private|protected|internal)\s+(static\s+)?`), 4},
	{regexp.MustCompile(`^\s*(override|virtual|abstract|final)\s+`), 4},
	{regexp.MustCompile(`^\s*@\w+`), 3},
	{regexp.MustCompile(`\[\w+\]`), 1},
	{regexp.MustCompile(`\{\w+\}`), 1},
	{regexp.MustCompile(`\(\w+:`), 2},
	{regexp.MustCompile(`->\w+`), 2},
	{regexp.MustCompile(`::\w+`), 3},
	{regexp.MustCompile(`\w+\(\)`), 2},
	{regexp.MustCompile(`^\s*#include`), 4},
	{regexp.MustCompile(`^\s*#define`), 3},
	{regexp.MustCompile(`^\s*using\s+`), 3},
	{regexp.MustCompile(`^\s*namespace\s+`), 4},
}

func calculateKeywordScore(line string) int {
	score := 0
	lowerLine := strings.ToLower(line)

	for keyword, weight := range languageKeywords {
		pattern := regexp.MustCompile(`\b` + keyword + `\b`)
		if pattern.MatchString(lowerLine) {
			score += weight
		}
	}

	return score
}

func calculateOperatorScore(line string, contentStart int) int {
	score := 0
	for i := contentStart; i < len(line); i++ {
		if weight, ok := importantOperators[rune(line[i])]; ok {
			score += weight
		}
	}
	return score
}

func calculatePatternScore(line string) int {
	score := 0
	for _, ip := range lineScoringPatterns {
		if ip.pattern.MatchString(line) {
			score += ip.score
		}
	}
	return score
}
