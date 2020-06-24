package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/parnurzeal/gorequest"
)

// 自定义爬虫参数
const (
	CusUserAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1"
	TargetURL    = "https://weibang.youth.cn/sapi/v1/share_examination/getShareExaminationDetail"
)

// DetailData detail数据结构
type DetailData struct {
	Data struct {
		Detail struct {
			StartTime int64  `json:"startTime"`
			EndTime   int64  `json:"endTime"`
			Title     string `json:"title"`
		} `json:"detail"`
	} `json:"data"`
}

// RankData 排行榜数据结构
type RankData struct {
	Data struct {
		RankList []struct {
			UserNickName string `json:"userNickName"`
			UserID       string `json:"userId"`
		} `json:"rankList"`
	} `json:"data"`
}

// UserInfo 用于请求用户题目的信息
type UserInfo struct {
	MyUID            string `json:"my_uid"`
	Phone            string `json:"phone"`
	ShareExamationID string `json:"share_examation_id"`
}

// *************************

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
	QuestionType  int    `json:"questionType"`
	Selects       []struct {
		Content string `json:"content"`
	}
}

var questionCHAN chan Question
var userList []UserInfo
var detailData DetailData

// 获取配置
func getConfig() string {
	f, err := ioutil.ReadFile("./exam_id.txt")
	if err != nil {
		fmt.Println(err)
		panic(err)

	}

	return string(f)
}

//获取含有答案的题目
func getQuesWithAns() {
	request := gorequest.New()
	for _, user := range userList {
		_, body, errs := request.Post("https://weibang.youth.cn/sapi/v1/share_examination/getAnswerInfo").
			Set("User-Agent", CusUserAgent).
			Send(user).
			End()

		if errs != nil {
			fmt.Println("发送请求失败", errs)
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
	close(questionCHAN)

}

//写入xlsx
func writeExcel() {
	count := 0
	repeatCount := 0
	fileName := "./target.xlsx"

	var f *excelize.File

	_, err := os.Stat(fileName)
	if err != nil {
		fmt.Printf("target.xlsx 文件不存在 -> ")
		fmt.Println(err)

		f = excelize.NewFile()
	} else {
		f, err = excelize.OpenFile(fileName)
		if err != nil {
			fmt.Println(err)
			return
		}
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
			f.SaveAs(fileName)
		}

		repeat := false

		//查看当前题目是否重复

		rows := f.GetRows("Sheet1")
		maxRow := len(rows)

		for i := 1; i <= maxRow; i++ {
			if f.GetCellValue("Sheet1", "A"+strconv.Itoa(i)) == qs.QuestionID {
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

			f.SetCellValue("Sheet1", "B"+strconv.Itoa(nowRow), qs.Content)
			for qi, qv := range qs.Selects {
				switch qi {
				case 0:
					f.SetCellValue("Sheet1", "C"+strconv.Itoa(nowRow), qv.Content)
				case 1:
					f.SetCellValue("Sheet1", "D"+strconv.Itoa(nowRow), qv.Content)
				case 2:
					f.SetCellValue("Sheet1", "E"+strconv.Itoa(nowRow), qv.Content)
				case 3:
					f.SetCellValue("Sheet1", "F"+strconv.Itoa(nowRow), qv.Content)
				case 4:
					f.SetCellValue("Sheet1", "G"+strconv.Itoa(nowRow), qv.Content)
				}
			}
			f.SetCellValue("Sheet1", "H"+strconv.Itoa(nowRow), qs.CorrectAnswer)
			f.SetCellValue("Sheet1", "A"+strconv.Itoa(nowRow), qs.QuestionID)
			f.SetCellValue("Sheet1", "I"+strconv.Itoa(nowRow), type2str(qs.QuestionType))
			f.SetCellValue("Sheet1", "K"+strconv.Itoa(nowRow), detailData.Data.Detail.Title)
			f.SetCellValue("Sheet1", "L"+strconv.Itoa(nowRow), "微邦")

			fmt.Printf("保存了 %d \n", count)
		}
	}

	//结束后保存
	if err := f.SaveAs(fileName); err != nil {
		fmt.Println(err)
	}
	fmt.Println("保存文件成功")
}

func type2str(t int) string {
	switch t {
	case 1:
		return "单选"
	case 2:
		return "多选"
	case 3:
		return "判断"
	default:
		return "默认"
	}
}

// 获取排行榜信息
func getTop(examID string) {
	request := gorequest.New()
	_, body, errs := request.Post("https://weibang.youth.cn/sapi/v1/share_examination/getShareExaminationDetail").
		Set("User-Agent", CusUserAgent).
		Send(map[string]string{
			"my_uid":               "0",
			"share_examination_id": examID,
		}).
		End()
	if errs != nil {
		fmt.Println(errs)
		return
	}

	detailData = DetailData{}
	json.Unmarshal([]byte(body), &detailData)

	return
}

// 获取手机号
func getPhones(examID string) {
	request := gorequest.New()
	_, body, errs := request.Post("https://weibang.youth.cn/sapi/v1/share_examination/getTotalScoreExaminationRankList").
		Set("User-Agent", CusUserAgent).
		Send(map[string]interface{}{
			"my_uid":             "0",
			"phone":              "phone",
			"share_examation_id": examID,
			"topNumber":          500,
		}).
		End()
	if errs != nil {
		fmt.Printf("获取手机号错误 -> ")
		fmt.Println(errs)
		return
	}

	data := RankData{}
	json.Unmarshal([]byte(body), &data)

	for _, item := range data.Data.RankList {
		userList = append(userList, UserInfo{
			MyUID:            item.UserID,
			Phone:            item.UserNickName,
			ShareExamationID: examID,
		})
	}
}

// Start 爬虫入口
func main() {
	//初始化通道

	questionCHAN = make(chan Question, 50)
	userList = make([]UserInfo, 0)

	// 读取配置
	eid := getConfig()

	// 获取排行榜信息
	getTop(eid)
	// 不在答题时间段则不理他
	nowTime := time.Now().UnixNano() / 1e6
	if nowTime < detailData.Data.Detail.StartTime && nowTime > detailData.Data.Detail.EndTime {
		fmt.Println("当前时间不在答题时间段")
		return
	}

	// 获取phone
	getPhones(eid)

	// 获取题目
	go getQuesWithAns()

	// 写入文件
	writeExcel()
}
