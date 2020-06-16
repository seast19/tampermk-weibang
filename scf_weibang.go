package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
)

/*
微邦答题助手
腾讯云SCF 后台 v2
*/

// DefineEvent 入参参数
type DefineEvent struct {
	Body string `json:"body"` // http请求的 body 参数
}

// Question 输入的问题结构
type Question struct {
	Q    string `json:"q"`    //题目
	OptA string `json:"opta"` //选项A
	Type string `json:"type"` // 类型 单选、多选、判断
}

// Answer 返回的答案结构
type Answer struct {
	Q    string   `json:"q"`    //题目
	A    []string `json:"a"`    //答案列表 ['ans1','ans2','ans3']
	Opt  string   `json:"opt"`  // 答案选项 AB
	Type string   `json:"type"` //类型
}

// 全局题目 答案 列表
var questionsList = []Question{}
var answerList = []Answer{}

// 过滤特殊字符
func filterSymbol(old string) string {
	newS := strings.ReplaceAll(old, " ", "")
	newS = strings.ReplaceAll(newS, "(", "")
	newS = strings.ReplaceAll(newS, ")", "")
	newS = strings.ReplaceAll(newS, "（", "")
	newS = strings.ReplaceAll(newS, "）", "")
	newS = strings.ReplaceAll(newS, "_", "")
	newS = strings.ReplaceAll(newS, "", "")
	newS = strings.ReplaceAll(newS, "\n", "")
	newS = strings.ReplaceAll(newS, "\r", "")

	newS = strings.TrimSpace(newS)

	return newS
}

// 添加答案至全局列表
func appendAns(ques Question, row []string) {
	//临时答案
	tempAns := []string{}

	//切割答案，以匹配多选
	ansOptList := strings.Split(row[7], "")
	for _, ansOpt := range ansOptList {
		switch ansOpt {
		case "A":
			tempAns = append(tempAns, strings.TrimSpace(row[2]))
		case "B":
			tempAns = append(tempAns, strings.TrimSpace(row[3]))
		case "C":
			tempAns = append(tempAns, strings.TrimSpace(row[4]))
		case "D":
			tempAns = append(tempAns, strings.TrimSpace(row[5]))
		case "E":
			tempAns = append(tempAns, strings.TrimSpace(row[6]))
		default:

		}
	}

	//构建答案
	answerList = append(answerList, Answer{
		Q:    ques.Q, //题目原样返回至前端
		A:    tempAns,
		Opt:  row[7],
		Type: row[8],
	})

}

// MatchAns 生成答案
func MatchAns() error {
	// 打开题库文件
	f, err := excelize.OpenFile("./lib2020-06-13.xlsx")
	if err != nil {
		fmt.Println("[MatchAns]打开xlsx失败->", err)
		return errors.New("打开题库文件失败")
	}
	rows := f.GetRows("Sheet1")

	// 逆序题目，保证匹配到最新的题目
	length := len(rows)
	for i := 0; i < length/2; i++ {
		temp := rows[length-1-i]
		rows[length-1-i] = rows[i]
		rows[i] = temp
	}

	//循环题目，获取答案
repeat:
	for quesNum, ques := range questionsList {
		// 首先查找type opta q 参数均匹配的
		for _, row := range rows {
			// 将题目中的特殊字符去掉后进行对比，题目一致且选项A也一致才认为相同
			if filterSymbol(ques.Q) == filterSymbol(row[1]) && filterSymbol(ques.OptA) == filterSymbol(row[2]) && filterSymbol(ques.Type) == filterSymbol(row[8]) {
				fmt.Printf("[%s]\t[第%d题]->%s\n", row[7], quesNum+1, row[1])

				appendAns(ques, row)

				//匹配完答案则下一题
				continue repeat
			}
		}

		// 没有的话查找 opta q 参数匹配的
		for _, row := range rows {
			// 将题目中的特殊字符去掉后进行对比，题目一致且选项A也一致才认为相同
			if filterSymbol(ques.Q) == filterSymbol(row[1]) && filterSymbol(ques.OptA) == filterSymbol(row[2]) {
				fmt.Printf("[%s]\t[第%d题]->%s\n", row[7], quesNum+1, row[1])

				appendAns(ques, row)

				//匹配完答案则下一题
				continue repeat
			}
		}

		// 没有的话查找 q 参数匹配的
		for _, row := range rows {
			// 将题目中的特殊字符去掉后进行对比，题目一致且选项A也一致才认为相同
			if filterSymbol(ques.Q) == filterSymbol(row[1]) {
				fmt.Printf("[%s]\t[第%d题]->%s\n", row[7], quesNum+1, row[1])

				appendAns(ques, row)

				//匹配完答案则下一题
				continue repeat
			}
		}
	}

	return nil
}

// Scf 云函数入口
func Scf(event DefineEvent) ([]Answer, error) {
	// 初始化题目和答案列表
	questionsList = []Question{}
	answerList = []Answer{}

	//反序列化json获取题目list
	err := json.Unmarshal([]byte(event.Body), &questionsList)
	if err != nil {
		fmt.Println("反序列化输入失败", err)
		return nil, err
	}

	fmt.Println("题目输入:")
	if len(questionsList) > 0 {
		for i, v := range questionsList {
			fmt.Printf("[%d]%s\n", i+1, v.Q)
		}
	}
	fmt.Println("************************")

	// 查询答案
	MatchAns()

	return answerList, nil
}

func main() {
	cloudfunction.Start(Scf)
}
