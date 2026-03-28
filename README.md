# 第50屆日本眾議院議員總選舉選舉結果提取器

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)

---

## 项目简介 | Project Overview

这是一个用于提取日本众议院议员总选举结果数据的 Go 语言工具。

This is a Go-based tool for extracting data from the results of the Japanese House of Representatives general election.

> ⚠️ **自用项目 | Personal Use Only**
> 
> 本项目为个人爬虫作业练习项目，仅供学习交流使用。
> 
> This project is for personal web scraping practice and educational purposes only.

---

## 功能说明 | Features

- **多线程处理**: 使用 16 个 worker 并发处理 HTML 文件
  
  **Multi-threading**: Uses 16 workers to process HTML files concurrently

- **智能提取**: 从维基百科格式的 HTML 中提取选举结果表格
  
  **Smart Extraction**: Extracts election result tables from Wikipedia-formatted HTML

- **结果合并**: 将所有提取的表格合并为一个 HTML 文件
  
  **Result Merging**: Merges all extracted tables into a single HTML file

---

## 使用方法 | Usage

```bash
# 编译 | Build
go build -o html.exe html.go

# 运行 | Run
./html.exe
```

---

## 目录结构 | Directory Structure

```
.
├── html.go          # 主程序 | Main program
├── html.exe         # 编译后的可执行文件 | Compiled executable
├── Temp/            # 存放源 HTML 文件 | Source HTML files directory
│   ├── 1.html
│   ├── 2.html
│   └── ...
├── all.html         # 输出文件 | Output file
└── README.md        # 本文件 | This file
```

---

## 技术细节 | Technical Details

### 提取逻辑 | Extraction Logic

程序查找包含 `"第50屆日本眾議院議員總選舉"` 的目标文本第三次出现的位置，然后提取：

The program finds the 3rd occurrence of the target text `"第50屆日本眾議院議員總選舉"`, then extracts:

1. 从目标文本所在行的行首开始
   
   From the beginning of the line where the target text appears

2. 到该行后面第一个 `<table class="wikitable">` 结束
   
   To the end of the first `<table class="wikitable">` after that line

### 依赖 | Dependencies

- Go 1.21+
- 标准库 (Standard Library Only):
  - `bytes`
  - `fmt`
  - `html`
  - `os`
  - `path/filepath`
  - `sort`
  - `strconv`
  - `strings`
  - `sync`

---

## 数据来源 | Data Source

数据来源于维基百科各选区页面，仅供学习参考。

Data sourced from Wikipedia constituency pages, for educational reference only.

---

## 许可证 | License

本项目为个人学习项目，不提供任何许可证。

This is a personal learning project, no license is provided.

---

## 作者 | Author

个人爬虫作业练习

Personal web scraping practice project
