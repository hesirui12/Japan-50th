package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	data, err := os.ReadFile("url.json")
	if err != nil {
		fmt.Printf("读取 url.json 失败: %v\n", err)
		return
	}

	var urls []string
	err = json.Unmarshal(data, &urls)
	if err != nil {
		fmt.Printf("解析 url.json 失败: %v\n", err)
		return
	}

	err = os.MkdirAll("Temp", 0755)
	if err != nil {
		fmt.Printf("创建 Temp 目录失败: %v\n", err)
		return
	}

	client := &http.Client{}

	for i, url := range urls {
		fmt.Printf("正在获取 [%d/%d]: %s\n", i+1, len(urls), url)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("创建请求失败: %v\n", err)
			continue
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("请求失败: %v\n", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("读取响应失败: %v\n", err)
			continue
		}

		if len(body) < 3072 {
			fmt.Printf("响应大小 %d 字节 (< 3KB), 等待 1 秒...\n", len(body))
			time.Sleep(1 * time.Second)
		}

		filename := filepath.Join("Temp", strconv.Itoa(i+1)+".html")
		err = os.WriteFile(filename, body, 0644)
		if err != nil {
			fmt.Printf("保存文件失败: %v\n", err)
			continue
		}

		fmt.Printf("已保存到: %s (大小: %d 字节)\n", filename, len(body))
	}

	fmt.Println("所有 URL 处理完成")
}
