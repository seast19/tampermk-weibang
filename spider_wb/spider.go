package spider

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/parnurzeal/gorequest"
)

// 自定义爬虫参数
const (
	CusUserAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1"
	TargetURL    = "https://weibang.youth.cn/sapi/v1/share_examination/getShareExaminationDetail"
)

// DataStruct 数据结构体
type DataStruct struct {
	Data struct {
		QuestionList []Question `json:"questionList"`
	} `json:"data"`
}

// Question 题目结构
type Question struct {
	Content       string `json:"content"`
	QuestionID    string `json:"questionId"`
	CorrectAnswer string `json:"correctAnswer"`
	Selects       []struct {
		Content string `json:"content"`
	}
}

var questionCHAN chan Question
var phoneList []string

func getPhoneList() {
	f, err := os.Open("./spider/phone.txt")
	if err != nil {
		fmt.Println("打开文件错误", err)
		return
	}
	defer f.Close()

	r := bufio.NewReader(f)
repeat:
	for {
		phone, _, err := r.ReadLine()
		if err != nil && err == io.EOF {
			fmt.Println("读取结束", err)
			break
		}
		if err != nil && err != io.EOF {
			fmt.Println("读取错误", err)
			continue
		}

		//无重复则添加
		for _, v := range phoneList {
			if v == string(phone) {
				fmt.Println("号码重复，跳过", string(phone))
				continue repeat
			}
		}

		phoneList = append(phoneList, string(phone))
	}

	// fmt.Println(phoneList)
}

//获取含有答案的题目
func getQuesWithAns() {
	request := gorequest.New()

	for _, phone := range phoneList {

		_, body, errs := request.Post("https://weibang.youth.cn/sapi/v1/share_examination/getAnswerInfo").
			Set("User-Agent", CusUserAgent).
			Send(map[string]string{
				"phone":                phone,
				"share_examination_id": "XiY4AHhPM66BHGry",
			}).
			End()

		if errs != nil {
			fmt.Println("发送请求失败", errs)
			return
		}

		//fmt.Println(body)

		aaa := DataStruct{}

		err := json.Unmarshal([]byte(body), &aaa)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, v := range aaa.Data.QuestionList {
			questionCHAN <- v
		}
	}
	close(questionCHAN)

}

//获取题目(无答案)
func getQues() {
	request := gorequest.New()

	for true {
		_, body, errs := request.Post(TargetURL).
			Set("User-Agent", CusUserAgent).
			Send(map[string]string{
				"my_uid":               "0",
				"share_examination_id": "XiY4AHhPM66BHGry",
			}).
			End()

		if errs != nil {
			fmt.Println(errs)
			return
		}

		aaa := DataStruct{}

		err := json.Unmarshal([]byte(body), &aaa)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, v := range aaa.Data.QuestionList {
			questionCHAN <- v
		}
	}

}

//写入xlsx
func writeExcel() {
	count := 0
	repeatCount := 0

	f, err := excelize.OpenFile("./spider/lib.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}

restart:
	for qs := range questionCHAN {
		//重复次数过多则停止
		if repeatCount > 500 {
			fmt.Println("重复次数过多，退出")
			break
		}

		//每50道保存一次
		if count > 0 && count%50 == 0 {
			fmt.Println("临时保存")
			f.Save()
		}

		repeat := false
		//查看当前题目是否存在

		rows := f.GetRows("Sheet1")
		maxRow := len(rows)

		for i := 1; i <= maxRow; i++ {
			if f.GetCellValue("Sheet1", "H"+strconv.Itoa(i)) == qs.QuestionID {
				repeatCount++
				repeat = true
				fmt.Printf("重复了 %d \n", repeatCount)
				continue restart
			}
		}

		//不重复则添加
		if !repeat {
			count++
			repeatCount = 0
			nowRow := maxRow + 1

			f.SetCellValue("Sheet1", "A"+strconv.Itoa(nowRow), qs.Content)
			for qi, qv := range qs.Selects {
				switch qi {
				case 0:
					f.SetCellValue("Sheet1", "B"+strconv.Itoa(nowRow), qv.Content)
				case 1:
					f.SetCellValue("Sheet1", "C"+strconv.Itoa(nowRow), qv.Content)
				case 2:
					f.SetCellValue("Sheet1", "D"+strconv.Itoa(nowRow), qv.Content)
				case 3:
					f.SetCellValue("Sheet1", "E"+strconv.Itoa(nowRow), qv.Content)
				case 4:
					f.SetCellValue("Sheet1", "F"+strconv.Itoa(nowRow), qv.Content)
				}
			}
			f.SetCellValue("Sheet1", "G"+strconv.Itoa(nowRow), qs.CorrectAnswer)
			f.SetCellValue("Sheet1", "H"+strconv.Itoa(nowRow), qs.QuestionID)

			fmt.Printf("保存了 %d \n", count)
		}
	}

	//结束后保存
	if err := f.Save(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("保存文件成功")
}

// Start 爬虫入口
func Start() {

	//初始化通道
	questionCHAN = make(chan Question, 50)
	phoneList = make([]string, 0)

	getPhoneList()

	// go getQues()
	go getQuesWithAns()
	writeExcel()
}
