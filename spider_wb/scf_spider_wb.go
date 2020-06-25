package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/parnurzeal/gorequest"
)

// 自定义爬虫参数
const (
	CusUserAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1"
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
	} `json:"selects"`
}

// ExamID 数据库examids
type ExamID struct {
	ExamIDs    []string `json:"exam_ids"`
	UpdateTime int64    `json:"update_time"`
}

// UploadData 上传题目的格式
type UploadData struct {
	Requests []UploadDataBody `json:"requests"`
}

// UploadDataBody 上传题目的格式body
type UploadDataBody struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Body   struct {
		QID      string   `json:"qid"`
		Question string   `json:"question"`
		Answer   string   `json:"answer"`
		Type     string   `json:"type"`
		Selects  []string `json:"selects"`
		Title    string   `json:"title"`
		Website  string   `json:"website"`
	} `json:"body"`
}

// 获取数据库的examID
func getExamID() ([]string, error) {
	fmt.Println("获取数据库exanid")

	// 获取examids
	request := gorequest.New()
	_, body, errs := request.Get("https://lc-api.seast.net/1.1/classes/wb_examid/5ef4745dbaa3480008004933").
		Set("X-LC-Id", "hYVRtO7xCsS9k7ac4o9bfjKn-gzGzoHsz").
		Set("X-LC-Key", "u8XIvYFinbdemgmcSeFrLf87").
		End()
	if errs != nil {
		fmt.Println(errs)
		return []string{}, errors.New("发送请求失败")
	}

	data := ExamID{}
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		fmt.Printf("反序列化examid失败 -> %s", err)
		return []string{}, err
	}
	// fmt.Println(data)

	// 检查更新时间，一个小时内更新了的话就不操作
	if time.Now().Unix()-data.UpdateTime < 60*60 {
		return []string{}, errors.New("最近已更新")
	}

	fmt.Printf("examids = %s\n", data.ExamIDs)

	return data.ExamIDs, nil
}

// 获取详细信息
func getDetail(examID string) (*DetailData, error) {

	fmt.Println("获取试卷详细信息")

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
		return nil, errors.New("发送请求失败")
	}

	data := DetailData{}
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// 获取手机号
func getPhones(examID string) ([]string, error) {
	fmt.Println("获取手机号")

	request := gorequest.New()
	_, body, errs := request.Post("https://weibang.youth.cn/sapi/v1/share_examination/getTotalScoreExaminationRankList").
		Set("User-Agent", CusUserAgent).
		Send(map[string]interface{}{
			"my_uid":             "0",
			"phone":              "phone",
			"share_examation_id": examID,
			"topNumber":          10000,
		}).
		End()
	if errs != nil {
		fmt.Printf("获取手机号错误 -> ")
		fmt.Println(errs)
		return []string{}, errors.New("发送请求失败")
	}

	data := RankData{}
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		fmt.Println(err)
		return []string{}, err
	}

	fmt.Printf("排行榜共%d人\n", len(data.Data.RankList))

	phoneList := []string{}
	for _, item := range data.Data.RankList {
		if checkPhone(item.UserNickName) {
			phoneList = append(phoneList, item.UserNickName)
		}
	}

	fmt.Printf("有效号码 %d 个\n", len(phoneList))
	return phoneList, nil
}

//获取含有答案的题目
func getQuesWithAns(phoneList []string) []Question {
	fmt.Println("获取题目")

	// phoneList = phoneList[0:10]

	qList := []Question{}

	request := gorequest.New()

	for _, phone := range phoneList {
		_, body, errs := request.Post("https://weibang.youth.cn/sapi/v1/share_examination/getAnswerInfo").
			Set("User-Agent", CusUserAgent).
			Send(map[string]string{"phone": phone}).
			End()
		if errs != nil {
			fmt.Println("发送请求失败", errs)
			continue
		}

		data := DataStruct{}
		err := json.Unmarshal([]byte(body), &data)
		if err != nil {
			fmt.Println(err)
			continue
		}

	noRepeat:
		for _, v := range data.Data.QuestionList {
			// 若有重复的则跳过
			for _, q := range qList {
				if q.QuestionID == v.QuestionID {
					continue noRepeat
				}
			}

			qList = append(qList, v)
		}
	}

	fmt.Printf("有效题目%d\n", len(qList))
	return qList

}

// 随机打乱数组
func randArr(strs []string) []string {
	rand.Seed(time.Now().UnixNano())
	for i := len(strs) - 1; i > 0; i-- {
		num := rand.Intn(i + 1)
		strs[i], strs[num] = strs[num], strs[i]
	}
	return strs
}

// 类型编号转文字
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

// 检查号码是否为手机号
func checkPhone(s string) bool {
	r := regexp.MustCompile(`^\d{11}$`)

	isPhone := r.MatchString(s)

	if isPhone {
		// fmt.Printf("手机号：%s\n", s)
		return true
	}

	// fmt.Printf("非手机号：%s\n", s)
	return false
}

// 检查题目，看是否重复
func checkQuestionsRepeat(qList []Question) []Question {
	fmt.Println("检查数据库是否有重复题目")

	fmt.Printf("原题目共有%d\n", len(qList))

	allQuestions := make([]Question, len(qList))
	copy(allQuestions, qList)

	tempQList := []Question{}

	responseAll := []struct {
		Qid string `json:"qid"`
	}{}

	// 分组查询，每100个查询一次
	for true {
		// 大于100则取前100个，小于100则取完
		if len(allQuestions) > 100 {
			tempQList = allQuestions[0:100]

		} else {
			tempQList = allQuestions
		}

		fmt.Printf("本次题目数%d\n", len(tempQList))

		// 构件ids参数，用于查询
		idLiist := []string{}
		for _, qitem := range tempQList {
			idLiist = append(idLiist, qitem.QuestionID)
		}
		idListStr, err := json.Marshal(idLiist)
		if err != nil {
			fmt.Println(err)
			return []Question{}
		}

		// 查询数据库相同的题目
		query := fmt.Sprintf(`?keys=qid&where={"qid":{"$in":%s}}`, string(idListStr))
		request := gorequest.New()
		_, body, errs := request.Get("https://lc-api.seast.net/1.1/classes/wb_questions"+query).
			Set("X-LC-Id", "hYVRtO7xCsS9k7ac4o9bfjKn-gzGzoHsz").
			Set("X-LC-Key", "u8XIvYFinbdemgmcSeFrLf87").
			End()
		if errs != nil {
			fmt.Println(errs)
			return []Question{}
		}

		data := struct {
			Results []struct {
				Qid string `json:"qid"`
			} `json:"results"`
		}{}

		err = json.Unmarshal([]byte(body), &data)
		if err != nil {
			fmt.Println(err)
			return []Question{}
		}

		fmt.Printf("数据库有%d个重复\n", len(data.Results))

		responseAll = append(responseAll, data.Results...)

		if len(allQuestions) > 100 {
			allQuestions = allQuestions[100:]
		} else {
			break
		}
	}

	// 排除重复
	newQList := []Question{}
	for _, v1 := range qList {
		// 查找是否重复
		isRepeat := false
		for _, v2 := range responseAll {
			if v1.QuestionID == v2.Qid {
				isRepeat = true
				break
			}
		}

		// 不重复则添加到新数组
		if !isRepeat {
			newQList = append(newQList, v1)
		}
	}

	// fmt.Println(newQList)
	fmt.Printf("过滤后题目共有%d\n", len(newQList))
	// 过滤相同的题目
	return newQList

}

// 提交题目
func submitQuestions(qList []Question, detail DetailData) {
	fmt.Println("提交题目到数据库")

	tempQList := []Question{}

	for true {
		if len(qList) > 300 {
			tempQList = qList[0:300]

		} else {
			tempQList = qList
		}

		fmt.Printf("本次题目%d个\n", len(tempQList))

		// 构建body
		bodys := []UploadDataBody{}
		for _, qitem := range tempQList {
			// 答案选项转为数组
			ansList := []string{}
			for _, ans := range qitem.Selects {
				ansList = append(ansList, ans.Content)
			}

			dataBody := UploadDataBody{
				Method: "POST",
				Path:   "/1.1/classes/wb_questions",
				Body: struct {
					QID      string   `json:"qid"`
					Question string   `json:"question"`
					Answer   string   `json:"answer"`
					Type     string   `json:"type"`
					Selects  []string `json:"selects"`
					Title    string   `json:"title"`
					Website  string   `json:"website"`
				}{
					QID:      qitem.QuestionID,
					Question: qitem.Content,
					Answer:   qitem.CorrectAnswer,
					Type:     type2str(qitem.QuestionType),
					Selects:  ansList,
					Title:    detail.Data.Detail.Title,
					Website:  "微邦",
				},
			}
			bodys = append(bodys, dataBody)
		}

		data := UploadData{
			Requests: bodys,
		}

		request := gorequest.New()
		_, body, errs := request.Post("https://lc-api.seast.net/1.1/batch").
			Set("X-LC-Id", "hYVRtO7xCsS9k7ac4o9bfjKn-gzGzoHsz").
			Set("X-LC-Key", "u8XIvYFinbdemgmcSeFrLf87").
			Send(data).
			End()
		if errs != nil {
			fmt.Println(errs)
			return
		}

		fmt.Println(body)

		// 退出条件
		if len(qList) > 300 {
			qList = qList[300:]
		} else {
			break
		}
	}

}

func main() {

	// cloudfunction.Start(Scf)

	// 获取数据库中的exam id
	examIDs, err := getExamID()
	if err != nil {
		return
	}

	examID := examIDs[0]

	// 根据examid获取详细信息，检查答题是否结束
	detail, err := getDetail(examID)
	if err != nil {
		return
	}
	nowTime := time.Now().UnixNano() / 1e6
	if nowTime < detail.Data.Detail.StartTime || nowTime > detail.Data.Detail.EndTime {
		fmt.Printf("%s - %s 已超出答题时间，删除题目", detail.Data.Detail.Title, examID)
		return
	}

	// 获取排行榜的手机号
	phoneList, err := getPhones(examID)
	if err != nil {
		return
	}

	if len(phoneList) == 0 {
		fmt.Println("phoneList is 0")
		return
	}

	// 打乱手机号顺序
	phoneList = randArr(phoneList)

	// 根据手机号获取题目
	qList := getQuesWithAns(phoneList)

	// 检查题目重复
	qList = checkQuestionsRepeat(qList)
	// 提交题目
	submitQuestions(qList, *detail)

}
