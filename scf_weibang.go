package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sync"

	"github.com/parnurzeal/gorequest"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
)

/*
微邦答题助手
腾讯云SCF 后台 v2
*/

// DefineEvent 入参参数
type DefineEvent struct {
	Headers struct {
		Host string `json:"host"`
	} `json:"headers"`
	Path string `json:"path"`
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
// var questionsList = []Question{}
// var answerList = []Answer{}

// 过滤特殊字符
// func filterSymbol(old string) string {
// 	newS := strings.ReplaceAll(old, " ", "")
// 	newS = strings.ReplaceAll(newS, "(", "")
// 	newS = strings.ReplaceAll(newS, ")", "")
// 	newS = strings.ReplaceAll(newS, "（", "")
// 	newS = strings.ReplaceAll(newS, "）", "")
// 	newS = strings.ReplaceAll(newS, "_", "")
// 	newS = strings.ReplaceAll(newS, "", "")
// 	newS = strings.ReplaceAll(newS, "\n", "")
// 	newS = strings.ReplaceAll(newS, "\r", "")

// 	newS = strings.TrimSpace(newS)

// 	return newS
// }

// 添加答案至全局列表
// func appendAns(ques Question, row []string) {
// 	//临时答案
// 	tempAns := []string{}

// 	//切割答案，以匹配多选
// 	ansOptList := strings.Split(row[7], "")
// 	for _, ansOpt := range ansOptList {
// 		switch ansOpt {
// 		case "A":
// 			tempAns = append(tempAns, strings.TrimSpace(row[2]))
// 		case "B":
// 			tempAns = append(tempAns, strings.TrimSpace(row[3]))
// 		case "C":
// 			tempAns = append(tempAns, strings.TrimSpace(row[4]))
// 		case "D":
// 			tempAns = append(tempAns, strings.TrimSpace(row[5]))
// 		case "E":
// 			tempAns = append(tempAns, strings.TrimSpace(row[6]))
// 		default:

// 		}
// 	}

// 	//构建答案
// 	answerList = append(answerList, Answer{
// 		Q:    ques.Q, //题目原样返回至前端
// 		A:    tempAns,
// 		Opt:  row[7],
// 		Type: row[8],
// 	})

// }

// 查找数据库中的题目
func searchquestions(qList []Question) []Answer {
	fmt.Println("开始获取题目")

	fmt.Printf("原题目共有%d题\n", len(qList))

	answerList := []Answer{}

	count := make(chan int, 3)
	// allQuestions := make([]Question, len(qList))
	// copy(allQuestions, qList)

	// tempQList := []Question{}

	// responseAll := []struct {
	// 	Qid string `json:"qid"`
	// }{}

	// lock := sync.Mutex
	var lock sync.Mutex
	var wg sync.WaitGroup

	// 构件ids参数，用于查询
	for _, qitem := range qList {
		//构建答案
		// idLiist := []string{}
		// for _, qitem := range qList {
		// 	idLiist = append(idLiist, qitem.QuestionID)
		// }
		// idListStr, err := json.Marshal(idLiist)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return []Question{}
		// }

		wg.Add(1)
		go func(qitem Question) {
			count <- 0
			defer func() {
				<-count
				wg.Done()
			}()

			// 查询数据库相同的题目
			query := fmt.Sprintf(`?where={"question":"%s","selects":"%s"}`, qitem.Q, qitem.OptA)
			request := gorequest.New()
			_, body, errs := request.Get("https://lc-api.seast.net/1.1/classes/wb_questions"+query).
				Set("X-LC-Id", "hYVRtO7xCsS9k7ac4o9bfjKn-gzGzoHsz").
				Set("X-LC-Key", "u8XIvYFinbdemgmcSeFrLf87").
				End()
			if errs != nil {
				fmt.Println(errs)
				return
			}

			fmt.Println(body)

			data := struct {
				Results []struct {
					QID      string   `json:"qid"`
					Question string   `json:"question"`
					Answer   string   `json:"answer"`
					Type     string   `json:"type"`
					Selects  []string `json:"selects"`
					Title    string   `json:"title"`
					Website  string   `json:"website"`
				} `json:"results"`
			}{}

			err := json.Unmarshal([]byte(body), &data)
			if err != nil {
				fmt.Println(err)
				return
			}

			// fmt.Println(data)

			if len(data.Results) > 0 {
				lock.Lock()
				answerList = append(answerList, Answer{
					Q:    data.Results[0].Question,
					A:    data.Results[0].Selects,
					Opt:  data.Results[0].Answer,
					Type: data.Results[0].Type,
				})
				lock.Unlock()
			}
		}(qitem)

	}

	wg.Wait()

	return answerList
}

// 添加数据库中的某个examid
func addExamID(examID string) {
	if len(examID) == 0 {
		return
	}

	data := struct {
		ExamIDs struct {
			Op      string   `json:"__op"`
			Objects []string `json:"objects"`
		} `json:"exam_ids"`
	}{
		ExamIDs: struct {
			Op      string   `json:"__op"`
			Objects []string `json:"objects"`
		}{
			Op: "AddUnique",
			Objects: []string{
				examID,
			},
		},
	}

	request := gorequest.New()
	_, body, errs := request.Put("https://lc-api.seast.net/1.1/classes/wb_examid/5ef4745dbaa3480008004933").
		Set("X-LC-Id", "hYVRtO7xCsS9k7ac4o9bfjKn-gzGzoHsz").
		Set("X-LC-Key", "u8XIvYFinbdemgmcSeFrLf87").
		Send(data).
		End()
	if errs != nil {
		fmt.Printf("getPhone 发送请求错误")
		fmt.Println(errs)
		return
	}

	fmt.Println(body)
}

// Scf 云函数入口
func Scf(event DefineEvent) ([]Answer, error) {
	fmt.Println(event)

	// 初始化题目和答案列表
	questionsList := []Question{}
	// answerList := []Answer{}

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
	// MatchAns()
	answerList := searchquestions(questionsList)

	// 有查找不到的题目则加入examid
	if len(questionsList) != len(answerList) {

		if event.Headers.Host == "weibang.youth.cn" {
			// 匹配examid
			r := regexp.MustCompile(`detail/(.*?)/showDetail/`)
			res := r.FindStringSubmatch(event.Path)

			if len(res) == 2 {
				addExamID(res[1])
			}
		}

	}

	return answerList, nil
}

func main() {
	// test()
	cloudfunction.Start(Scf)

}

func test() {
	path := "https://weibang.youth.cn/webpage_sapi/examination/detail/2iWIg5Xb3FIinROD/showDetail/0/phone/platform/45c01a/orgId/httphost/httpport/token"

	r := regexp.MustCompile(`detail/(.*?)/showDetail/`)
	res := r.FindStringSubmatch(path)
	fmt.Println(res)
}
