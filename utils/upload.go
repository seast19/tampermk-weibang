package utils

import (
	"encoding/json"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/parnurzeal/gorequest"
	"strings"
)

// 过滤特殊字符
func filterSymbol(old string) string {
	newS := strings.ReplaceAll(old, " ", "")
	newS = strings.ReplaceAll(newS, "", "")
	newS = strings.ReplaceAll(newS, "\n", "")
	newS = strings.ReplaceAll(newS, "\r", "")
	newS = strings.TrimSpace(newS)
	return newS
}

// 手动上传题库到learncloud
func Upload() {
	//读取文件
	f, err := excelize.OpenFile("./q.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}

	rows := f.GetRows("Sheet1")
	for rowIndex, row := range rows {
		if rowIndex == 0 || rowIndex == 1 {
			continue
		}

		//检测重复
		query := fmt.Sprintf(`?where={"question":"%s"}`, filterSymbol(row[3]))
		request := gorequest.New()
		_, body, errs := request.Get("https://lc-api.seast.net/1.1/classes/wb_questions"+query).
			Set("X-LC-Id", "hYVRtO7xCsS9k7ac4o9bfjKn-gzGzoHsz").
			Set("X-LC-Key", "u8XIvYFinbdemgmcSeFrLf87").
			End()
		if errs != nil {
			fmt.Println(errs)
			continue
		}

		resSearch := struct {
			Results []interface{} `json:"results"`
		}{}
		err = json.Unmarshal([]byte(body), &resSearch)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if len(resSearch.Results) > 0 {
			fmt.Printf("题目重复 %d",rowIndex)
			continue
		}

		//构建请求体
		selects := []string{}
		if len(row[5]) > 0 {
			selects = append(selects, filterSymbol(row[5]))
		}
		if len(row[6]) > 0 {
			selects = append(selects, filterSymbol(row[6]))
		}
		if len(row[7]) > 0 {
			selects = append(selects, filterSymbol(row[7]))
		}
		if len(row[8]) > 0 {
			selects = append(selects, filterSymbol(row[8]))
		}
		if len(row[9]) > 0 {
			selects = append(selects, filterSymbol(row[9]))
		}

		data := struct {
			QID      string   `json:"qid"`
			Question string   `json:"question"`
			Answer   string   `json:"answer"`
			Type     string   `json:"type"`
			Selects  []string `json:"selects"`
			Title    string   `json:"title"`
			Website  string   `json:"website"`
		}{
			Question: filterSymbol(row[3]),
			Answer:   filterSymbol(row[4]),
			Type:     filterSymbol(row[0]),
			Selects:  selects,
			Title:    "手动添加",
			Website:  "手动添加",
		}

		//构建请求体
		request = gorequest.New()
		_, body, errs = request.Post("https://lc-api.seast.net/1.1/classes/wb_questions").
			Set("X-LC-Id", "hYVRtO7xCsS9k7ac4o9bfjKn-gzGzoHsz").
			Set("X-LC-Key", "u8XIvYFinbdemgmcSeFrLf87").
			Send(data).
			End()
		if errs != nil {
			fmt.Println(errs)
			continue
		}

		fmt.Printf("添加成功 %d\n",rowIndex)



	}

	fmt.Println("手动添加题库结束")

}
