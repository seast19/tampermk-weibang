package spider_zxw

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
)

type JsonData struct {
	TopicList []Question `json:"topicList"`
}

type Question struct {
	ID       string `json:"ttop001"`
	Type     int    `json:"basetype"`
	TypeText string `json:"ttop010"`
	Question string `json:"ttop011"`
	Opt      string `json:"ttop018"`

	Ans string `json:"ttop022"`
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

func writeExcel(d []Question) {
	// count := 0
	// repeatCount := 0

	f, err := excelize.OpenFile("./spider_zxw/data.xlsx")
	if err != nil {
		fmt.Println("openxlsx", err)
		return
	}

restart:
	for qi, qs := range d {
		//重复次数过多则停止
		// if repeatCount > 500 {
		// 	fmt.Println("重复次数过多，退出")
		// 	break
		// }

		//每50道保存一次
		// if count > 0 && count%50 == 0 {
		// 	fmt.Println("临时保存")
		// 	f.Save()
		// }

		repeat := false
		//查看当前题目是否存在

		rows := f.GetRows("Sheet1")
		maxRow := len(rows)

		for i := 1; i <= maxRow; i++ {
			if f.GetCellValue("Sheet1", "A"+strconv.Itoa(i)) == trimSpace(qs.ID) {
				// repeatCount++
				repeat = true
				fmt.Printf("重复了\n")
				continue restart
			}
		}

		//不重复则添加
		if !repeat {
			// count++
			// repeatCount = 0
			nowRow := maxRow + 1
			// 写id
			f.SetCellValue("Sheet1", "A"+strconv.Itoa(nowRow), trimSpace(qs.ID))
			//写入题目
			f.SetCellValue("Sheet1", "B"+strconv.Itoa(nowRow), trimSpace(qs.Question))

			// 写入选项
			if qs.Type == 1 {
				// 单选
				for qi, qv := range strings.Split(qs.Opt, "$$") {
					switch qi {
					case 0:
						f.SetCellValue("Sheet1", "C"+strconv.Itoa(nowRow), trimSpace(qv))
					case 1:
						f.SetCellValue("Sheet1", "D"+strconv.Itoa(nowRow), trimSpace(qv))
					case 2:
						f.SetCellValue("Sheet1", "E"+strconv.Itoa(nowRow), trimSpace(qv))
					case 3:
						f.SetCellValue("Sheet1", "F"+strconv.Itoa(nowRow), trimSpace(qv))
					case 4:
						f.SetCellValue("Sheet1", "G"+strconv.Itoa(nowRow), trimSpace(qv))
					}
				}
			} else if qs.Type == 3 {
				// 判断
				f.SetCellValue("Sheet1", "C"+strconv.Itoa(nowRow), "正确")
				f.SetCellValue("Sheet1", "D"+strconv.Itoa(nowRow), "错误")

			}

			f.SetCellValue("Sheet1", "H"+strconv.Itoa(nowRow), trimSpace(qs.Ans))
			f.SetCellValue("Sheet1", "I"+strconv.Itoa(nowRow), trimSpace(qs.TypeText))
			f.SetCellValue("Sheet1", "L"+strconv.Itoa(nowRow), "众学网")
			fmt.Println("写入了", qi)

		}
	}

	//结束后保存
	if err := f.Save(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("保存文件成功")
}

func Start() {
	// 读取json数据
	f, err := ioutil.ReadFile("./spider_zxw/data.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	jsonD := JsonData{}

	err = json.Unmarshal(f, &jsonD)
	if err != nil {
		fmt.Println(err)
		return
	}

	// fmt.Println(jsonD)

	writeExcel(jsonD.TopicList)
	//写入xlsx文件
}
