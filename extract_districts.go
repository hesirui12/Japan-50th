package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

func main() {
	// 打开文件
	file, err := os.Open("index.html")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 创建一个scanner来逐行读取文件
	scanner := bufio.NewScanner(file)

	// 跳过前2089行
	for i := 0; i < 2089; i++ {
		if !scanner.Scan() {
			fmt.Println("File ended before reaching line 2090")
			return
		}
	}

	// 定义正则表达式来匹配"第[数字]区"的链接
	re := regexp.MustCompile(`<a href="(/wiki/[^" ]+)" title="[^"]+第\d+区">第\d+区</a>`)

	// 存储提取的URL
	var urls []string

	// 遍历文件的其余部分
	for scanner.Scan() {
		line := scanner.Text()

		// 查找所有匹配的链接
		matches := re.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				// 构建完整的URL
				url := "https://zh.wikipedia.org" + match[1]
				urls = append(urls, url)
			}
		}
	}

	// 检查是否有扫描错误
	if err := scanner.Err(); err != nil {
		fmt.Println("Error scanning file:", err)
		return
	}

	// 将URL写入到url.json文件
	jsonFile, err := os.Create("url.json")
	if err != nil {
		fmt.Println("Error creating json file:", err)
		return
	}
	defer jsonFile.Close()

	// 编码为JSON
	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(urls); err != nil {
		fmt.Println("Error encoding to json:", err)
		return
	}

	fmt.Println("URLs have been written to url.json")
}
