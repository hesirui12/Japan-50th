package main

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	targetText  = "第50屆日本眾議院議員總選舉"
	tempDir     = "Temp"
	outputFile  = "all.html"
	workerCount = 16
)

type Result struct {
	FileName string
	Content  string
	Index    int
}

type FileTask struct {
	FilePath string
	Index    int
}

func main() {
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		return
	}

	var htmlFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") {
			htmlFiles = append(htmlFiles, filepath.Join(tempDir, entry.Name()))
		}
	}

	if len(htmlFiles) == 0 {
		fmt.Println("未找到HTML文件")
		return
	}

	sort.Slice(htmlFiles, func(i, j int) bool {
		numI := extractNumber(filepath.Base(htmlFiles[i]))
		numJ := extractNumber(filepath.Base(htmlFiles[j]))
		return numI < numJ
	})

	fmt.Printf("找到 %d 个HTML文件，开始处理...\n", len(htmlFiles))

	results := make(chan Result, len(htmlFiles))
	var wg sync.WaitGroup

	taskQueue := make(chan FileTask, len(htmlFiles))
	for idx, file := range htmlFiles {
		taskQueue <- FileTask{FilePath: file, Index: idx}
	}
	close(taskQueue)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range taskQueue {
				processFile(task, workerID, results)
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var allResults []Result
	for result := range results {
		if result.Content != "" {
			allResults = append(allResults, result)
		}
	}

	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Index < allResults[j].Index
	})

	if err := writeResults(allResults); err != nil {
		fmt.Printf("写入结果失败: %v\n", err)
		return
	}

	fmt.Printf("处理完成，共提取 %d 个区块\n", len(allResults))
}

func extractNumber(filename string) int {
	numStr := strings.TrimSuffix(filename, ".html")
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 999999
	}
	return num
}

func processFile(task FileTask, workerID int, results chan<- Result) {
	content, err := os.ReadFile(task.FilePath)
	if err != nil {
		fmt.Printf("Worker %d: 读取文件 %s 失败: %v\n", workerID, filepath.Base(task.FilePath), err)
		return
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, targetText) {
		return
	}

	sectionContent := extractSection(contentStr)
	if sectionContent == "" {
		fmt.Printf("Worker %d: 文件 %s 包含目标文本但未找到完整区块\n", workerID, filepath.Base(task.FilePath))
		return
	}

	results <- Result{
		FileName: filepath.Base(task.FilePath),
		Content:  sectionContent,
		Index:    task.Index,
	}
}

func extractSection(content string) string {
	// 查找第三次出现目标文本的位置
	searchStart := 0
	var targetIndex int = -1
	count := 0
	for {
		idx := strings.Index(content[searchStart:], targetText)
		if idx == -1 {
			break
		}
		idx += searchStart
		count++
		if count == 3 {
			targetIndex = idx
			break
		}
		searchStart = idx + len(targetText)
	}

	if targetIndex == -1 {
		return ""
	}

	lineStart := strings.LastIndex(content[:targetIndex], "\n")
	if lineStart == -1 {
		lineStart = 0
	} else {
		lineStart++
	}

	contentAfterLine := content[lineStart:]

	tableStart := strings.Index(contentAfterLine, `<table class="wikitable">`)
	if tableStart == -1 {
		tableStart = strings.Index(contentAfterLine, `<table class="wikitable"`)
	}
	if tableStart == -1 {
		return ""
	}

	contentAfterTableStart := contentAfterLine[tableStart:]

	var tableEnd int
	depth := 0
	pos := 0

	for pos < len(contentAfterTableStart) {
		if strings.HasPrefix(contentAfterTableStart[pos:], `<table`) {
			depth++
			pos += 6
			continue
		}
		if strings.HasPrefix(contentAfterTableStart[pos:], `</table>`) {
			depth--
			if depth == 0 {
				tableEnd = pos + 8
				break
			}
			pos += 8
			continue
		}
		pos++
	}

	if tableEnd == 0 {
		return ""
	}

	return contentAfterLine[:tableStart+tableEnd]
}

func writeResults(results []Result) error {
	var buffer bytes.Buffer
	buffer.WriteString("<!DOCTYPE html>\n<html>\n<head>\n<meta charset=\"UTF-8\">\n<title>第50屆日本眾議院議員總選舉選舉結果</title>\n<style>\n")
	buffer.WriteString("body { font-family: Arial, sans-serif; margin: 20px; }\n")
	buffer.WriteString("table.wikitable { border-collapse: collapse; margin: 20px 0; width: 100%; }\n")
	buffer.WriteString("table.wikitable th, table.wikitable td { border: 1px solid #ccc; padding: 8px; text-align: left; }\n")
	buffer.WriteString("table.wikitable th { background-color: #f2f2f2; }\n")
	buffer.WriteString(".section { margin-bottom: 40px; border-bottom: 2px solid #333; padding-bottom: 20px; }\n")
	buffer.WriteString(".filename { font-size: 18px; font-weight: bold; color: #333; margin-bottom: 10px; }\n")
	buffer.WriteString("</style>\n</head>\n<body>\n")
	buffer.WriteString("<h1>第50屆日本眾議院議員總選舉選舉結果</h1>\n")

	for _, result := range results {
		buffer.WriteString("<div class=\"section\">\n")
		buffer.WriteString(fmt.Sprintf("<div class=\"filename\">%s</div>\n", html.EscapeString(result.FileName)))
		buffer.WriteString(result.Content)
		buffer.WriteString("</div>\n")
	}

	buffer.WriteString("</body>\n</html>")

	if err := os.WriteFile(outputFile, buffer.Bytes(), 0644); err != nil {
		return fmt.Errorf("写入输出文件失败: %w", err)
	}

	return nil
}
