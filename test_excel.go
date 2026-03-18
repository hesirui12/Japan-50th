package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func main() {
	f, err := excelize.OpenFile("excel/1.xlsx")
	if err != nil {
		fmt.Printf("打开文件失败: %v\n", err)
		return
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		fmt.Printf("读取行失败: %v\n", err)
		return
	}

	fmt.Printf("总共 %d 行\n", len(rows))
	for i, row := range rows {
		if i < 10 {
			fmt.Printf("Row %d: %v\n", i, row)
		}
	}
}
