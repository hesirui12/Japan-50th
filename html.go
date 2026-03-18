package main

import (
	"bytes"
	"fmt"
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
	District string
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

	fmt.Printf("处理完成，共提取 %d 个表格\n", len(allResults))
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

	district := extractDistrict(contentStr)
	tableContent := extractFirstWikitable(contentStr)
	if tableContent == "" {
		fmt.Printf("Worker %d: 文件 %s 包含目标文本但未找到wikitable\n", workerID, filepath.Base(task.FilePath))
		return
	}

	results <- Result{
		FileName: filepath.Base(task.FilePath),
		Content:  tableContent,
		Index:    task.Index,
		District: district,
	}
}

func extractDistrict(content string) string {
	// 找到目标文本的位置
	targetIndex := strings.Index(content, targetText)
	if targetIndex == -1 {
		return ""
	}

	// 从目标文本位置开始搜索</small>标签
	smallEnd := strings.Index(content[targetIndex:], "</small>")
	if smallEnd == -1 {
		return ""
	}
	smallEnd += targetIndex

	// 从</small>标签位置开始搜索</div>标签
	divEnd := strings.Index(content[smallEnd:], "</div>")
	if divEnd == -1 {
		return ""
	}
	divEnd += smallEnd

	// 提取</small>到</div>之间的内容
	districtPart := strings.TrimSpace(content[smallEnd+8 : divEnd])
	return districtPart
}

func extractFirstWikitable(content string) string {
	targetIndex := strings.Index(content, targetText)
	if targetIndex == -1 {
		return ""
	}

	contentAfterTarget := content[targetIndex:]

	tableStart := strings.Index(contentAfterTarget, `<table class="wikitable">`)
	if tableStart == -1 {
		tableStart = strings.Index(contentAfterTarget, `<table class="wikitable"`)
	}
	if tableStart == -1 {
		return ""
	}

	contentAfterTableStart := contentAfterTarget[tableStart:]

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

	return contentAfterTableStart[:tableEnd]
}

func writeResults(results []Result) error {
	if err := os.WriteFile(outputFile, []byte(""), 0644); err != nil {
		return fmt.Errorf("清空输出文件失败: %w", err)
	}

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
		title := result.District
		if title == "" {
			title = result.FileName
		}
		buffer.WriteString(fmt.Sprintf("<div class=\"section\">\n"))
		buffer.WriteString(fmt.Sprintf("<div class=\"filename\">%s</div>\n", title))
		buffer.WriteString(result.Content)
		buffer.WriteString("</div>\n")
	}

	buffer.WriteString("</body>\n</html>")

	if err := os.WriteFile(outputFile, buffer.Bytes(), 0644); err != nil {
		return fmt.Errorf("写入输出文件失败: %w", err)
	}

	return nil
}
