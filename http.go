package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/xuri/excelize/v2"
	"golang.org/x/net/html"
)

type Task struct {
	Number int
	Path   string
}

func main() {
	tempDir := "Temp"
	excelDir := `c:\Users\jcsyh\project\golang\日本选区爬取\excel\`

	err := os.MkdirAll(excelDir, 0755)
	if err != nil {
		fmt.Printf("创建 excel 目录失败: %v\n", err)
		return
	}

	files, err := os.ReadDir(tempDir)
	if err != nil {
		fmt.Printf("读取 Temp 目录失败: %v\n", err)
		return
	}

	var tasks []Task
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".html") {
			numStr := strings.TrimSuffix(file.Name(), ".html")
			num, err := strconv.Atoi(numStr)
			if err != nil {
				continue
			}
			tasks = append(tasks, Task{
				Number: num,
				Path:   filepath.Join(tempDir, file.Name()),
			})
		}
	}

	fmt.Printf("找到 %d 个 HTML 文件待处理\n", len(tasks))

	taskChan := make(chan Task, len(tasks))
	var wg sync.WaitGroup

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go worker(taskChan, excelDir, &wg)
	}

	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	wg.Wait()
	fmt.Println("所有文件处理完成")
}

func worker(taskChan <-chan Task, excelDir string, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range taskChan {
		fmt.Printf("线程处理 [%d]: %s\n", task.Number, task.Path)
		err := processHTML(task.Path, task.Number, excelDir)
		if err != nil {
			fmt.Printf("处理文件 %s 失败: %v\n", task.Path, err)
		}
	}
}

func processHTML(filePath string, number int, excelDir string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	doc, err := html.Parse(strings.NewReader(string(content)))
	if err != nil {
		return fmt.Errorf("解析 HTML 失败: %w", err)
	}

	var targetTable *html.Node
	var foundSection bool

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if foundSection && targetTable != nil {
			return
		}

		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if strings.Contains(text, "第50届日本众议院议员总选举") || strings.Contains(text, "第50屆日本眾議院議員總選舉") {
				foundSection = true
			}
		}

		if foundSection && targetTable == nil {
			if n.Type == html.ElementNode && n.Data == "table" {
				for _, attr := range n.Attr {
					if attr.Key == "class" && strings.Contains(attr.Val, "wikitable") {
						targetTable = n
						return
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	if targetTable == nil {
		return fmt.Errorf("未找到第50届日本众议院议员总选举的 wikitable")
	}

	data := parseTable(targetTable)
	if len(data) == 0 {
		return fmt.Errorf("表格数据为空")
	}

	excelPath := filepath.Join(excelDir, fmt.Sprintf("%d.xlsx", number))
	err = createExcel(data, excelPath)
	if err != nil {
		return fmt.Errorf("创建 Excel 失败: %w", err)
	}

	return nil
}

func parseTable(table *html.Node) [][]string {
	var data [][]string

	var traverse func(*html.Node, *[]string)
	traverse = func(n *html.Node, row *[]string) {
		if n.Type == html.ElementNode && (n.Data == "tr") {
			if len(*row) > 0 {
				data = append(data, *row)
			}
			newRow := []string{}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				traverse(c, &newRow)
			}
			if len(newRow) > 0 {
				data = append(data, newRow)
			}
			return
		}

		if n.Type == html.ElementNode && (n.Data == "th" || n.Data == "td") {
			text := extractText(n)
			text = strings.TrimSpace(text)
			text = cleanText(text)
			*row = append(*row, text)
			return
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c, row)
		}
	}

	var firstRow []string
	traverse(table, &firstRow)

	return data
}

func extractText(n *html.Node) string {
	var text strings.Builder

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)
	return text.String()
}

func cleanText(text string) string {
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func createExcel(data [][]string, filePath string) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("关闭 Excel 文件失败: %v\n", err)
		}
	}()

	for i, row := range data {
		for j, cell := range row {
			cellName, err := excelize.CoordinatesToCellName(j+1, i+1)
			if err != nil {
				return err
			}
			f.SetCellValue("Sheet1", cellName, cell)
		}
	}

	if err := f.SaveAs(filePath); err != nil {
		return err
	}

	return nil
}
