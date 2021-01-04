package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
)

// 自定义爬虫参数
const (
	CusUserAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1"
)

// DetailData detail页面数据结构
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

// InfoData 考试记录数据结构体
type InfoData struct {
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

// UploadData 数据库上传题目的格式
type UploadData struct {
	Requests []UploadDataBody `json:"requests"`
}

// UploadDataBody 数据库上传题目的格式body
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

// GetExamID 获取数据库的examID
func GetExamID() ([]string, error) {
	fmt.Println("<!!> 开始获取数据库 exanid")

	// 获取examids
	request := gorequest.New()
	_, body, errs := request.Get("https://lc-api.seast.net/1.1/classes/wb_examid/5ef4745dbaa3480008004933").
		Set("X-LC-Id", "hYVRtO7xCsS9k7ac4o9bfjKn-gzGzoHsz").
		Set("X-LC-Key", "u8XIvYFinbdemgmcSeFrLf87").
		End()
	if errs != nil {
		fmt.Println(errs)
		return []string{}, errors.New("GetExamID 发送请求失败")
	}

	data := ExamID{}
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		fmt.Printf("GetExamID 反序列化examid失败 -> %s", err)
		return []string{}, err
	}

	// 检查更新时间，一个小时内更新了的话就不操作
	if time.Now().Unix()-data.UpdateTime < 60*60 {
		return []string{}, errors.New("1小时内已更新，不再更新")
	}

	fmt.Printf("examids = %s\n", data.ExamIDs)

	return randArr(data.ExamIDs), nil
}

// GetDetail 获取详细页面信息
func GetDetail(examID string) (*DetailData, error) {

	fmt.Println("<!!> 开始获取试卷详细信息")

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

	// 校验开始&结束时间是否错误
	if data.Data.Detail.EndTime == 0 {
		fmt.Printf("%s 获取不到正确detail数据\n", examID)
		return nil, errors.New("获取详细信息失败")

	}

	fmt.Printf("试卷标题 《%s》 \n", data.Data.Detail.Title)

	return &data, nil
}

// 删除数据库中的某个examid
func deleteExamID(examID string) {
	data := fmt.Sprintf(`{"exam_ids":{"__op":"Remove","objects":["%s"]}}`, examID)

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

	fmt.Printf("[tools] 删除 examid -> %s\n", body)
}

// 添加 examid 到数据库
func addExamID(examID string) {

	data := fmt.Sprintf(`{"exam_ids":{"__op":"AddUnique","objects":["%s"]}}`, examID)

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

	fmt.Printf("[tools] 添加 examid 到数据库 -> %s\n", body)
}

// GetPhones 获取手机号
func GetPhones(examID string) ([]string, error) {
	fmt.Println("<!!> 开始获取手机号")

	// 从排行榜获取个人信息
	request := gorequest.New()
	_, body, errs := request.Post("https://weibang.youth.cn/sapi/v1/share_examination/getTotalScoreExaminationRankList").
		Set("User-Agent", CusUserAgent).
		Send(map[string]interface{}{
			"my_uid":             "0",
			"phone":              "phone",
			"share_examation_id": examID,
			"topNumber":          5000,
		}).
		End()
	if errs != nil {
		fmt.Printf("getPhone 发送请求错误")
		fmt.Println(errs)
		return []string{}, errors.New("发送请求失败")
	}

	data := RankData{}
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		fmt.Println(err)
		return []string{}, err
	}

	fmt.Printf("排行榜共 %d 人\n", len(data.Data.RankList))

	phoneList := []string{}
	for _, item := range data.Data.RankList {
		if isPhone(item.UserNickName) {
			phoneList = append(phoneList, item.UserNickName)
		}
	}

	fmt.Printf("有效手机号码 %d 个\n", len(phoneList))

	if len(phoneList) == 0 {
		return []string{}, errors.New("手机号数目为0")
	}

	return phoneList, nil
}

// GetQuesWithAns 获取含有答案的题目
func GetQuesWithAns(phoneList []string) []Question {
	fmt.Println("<!!> 开始获取题目")

	qList := []Question{}

	request := gorequest.New()

	for _, phone := range phoneList {
		_, body, errs := request.Post("https://weibang.youth.cn/sapi/v1/share_examination/getAnswerInfo").
			Set("User-Agent", CusUserAgent).
			Send(map[string]string{"phone": phone}).
			End()
		if errs != nil {
			fmt.Println("getQuesWithAns发送请求失败", errs)
			continue
		}

		data := InfoData{}
		err := json.Unmarshal([]byte(body), &data)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// 保存不重复的题目
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

	fmt.Printf("有效题目%d题\n", len(qList))
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
func isPhone(s string) bool {
	r := regexp.MustCompile(`^\d{11}$`)
	isPhone := r.MatchString(s)
	if isPhone {
		return true
	}
	return false
}

// CheckQuestionsRepeat 检查题目，看是否重复
func CheckQuestionsRepeat(qList []Question) []Question {
	/** 通过 qid 查询数据库获取有 qid 的题目，再将所有题目与数据库返回的数据比较，获得不重复的题目
	 */

	fmt.Println("<!!> 开始检查数据库是否有重复题目")
	fmt.Printf("原题目共有 %d 题\n", len(qList))

	// 题数为0
	if len(qList) == 0 {
		return []Question{}
	}

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

		fmt.Printf("[分次查询] 本次查询题目 %d 个\n", len(tempQList))

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

		// 查询数据库相同的题目（使用qid查询）
		query := fmt.Sprintf(`?keys=qid&where={"qid":{"$in":%s}}`, string(idListStr))
		request := gorequest.New()
		_, body, errs := request.Get("https://lc-api.seast.net/1.1/classes/wb_questions").
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

		fmt.Printf("[分次查询] 数据库有 %d 个重复\n", len(data.Results))

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

	fmt.Printf("去除重复后题目共有 %d 个\n", len(newQList))

	return newQList

}

// SubmitQuestions 提交题目
func SubmitQuestions(qList []Question, detail DetailData) {
	fmt.Println("<!!> 开始提交题目到数据库")
	fmt.Printf("原题目共 %d 个 \n", len(qList))

	// 题数为0
	if len(qList) == 0 {
		return
	}

	tempQList := []Question{}

	for true {
		if len(qList) > 300 {
			tempQList = qList[0:300]
		} else {
			tempQList = qList
		}

		fmt.Printf("[分次上传] 上传题目 %d 个\n", len(tempQList))

		// 构建body
		bodys := []UploadDataBody{}
		for _, qitem := range tempQList {
			// 答案选项转为数组
			ansList := []string{}
			for _, ans := range qitem.Selects {
				ansList = append(ansList, filterSymbol(ans.Content))
			}

			// 将question、selects 进行去除特殊字符
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
					Question: filterSymbol(qitem.Content),
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
		_, _, errs := request.Post("https://lc-api.seast.net/1.1/batch").
			Set("X-LC-Id", "hYVRtO7xCsS9k7ac4o9bfjKn-gzGzoHsz").
			Set("X-LC-Key", "u8XIvYFinbdemgmcSeFrLf87").
			Send(data).
			End()
		if errs != nil {
			fmt.Println(errs)
			return
		}

		fmt.Println("[分次上传] 上传题目成功")

		// 退出条件
		if len(qList) > 300 {
			qList = qList[300:]
		} else {
			break
		}
	}

	fmt.Println("上传题目完毕")
}

// 过滤特殊字符
func filterSymbol(old string) string {
	newS := strings.ReplaceAll(old, " ", "")
	newS = strings.ReplaceAll(newS, "", "")
	newS = strings.ReplaceAll(newS, "\n", "")
	newS = strings.ReplaceAll(newS, "\r", "")
	newS = strings.TrimSpace(newS)
	return newS
}

// Scf 入口函数
func Scf() (string, error) {
	// 获取数据库中的exam id
	examIDs, err := GetExamID()
	if err != nil {
		fmt.Println(err)
		return "连接数据库失败", err
	}

	// 按examid依次爬取题目
	for _, examID := range examIDs {
		fmt.Println("***********************")
		fmt.Printf("当前执行 %s\n", examID)

		// 根据examid获取试卷详细信息
		detail, err := GetDetail(examID)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// 检查试卷答题时间是否结束
		nowTime := time.Now().UnixNano() / 1e6
		if nowTime < detail.Data.Detail.StartTime || nowTime > detail.Data.Detail.EndTime {
			fmt.Printf("%s - %s 已超出答题时间范围，将删除此记录", detail.Data.Detail.Title, examID)
			deleteExamID(examID)
			continue
		}

		// 获取排行榜的手机号
		phoneList, err := GetPhones(examID)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// 打乱手机号顺序，只取前100个足够
		phoneList = randArr(phoneList)
		if len(phoneList) > 100 {
			phoneList = phoneList[:100]
		}

		// 根据手机号获取题目
		qList := GetQuesWithAns(phoneList)

		// 检查题目重复
		qList = CheckQuestionsRepeat(qList)
		if len(qList) == 0 {
			fmt.Println("获得的新题目题数为 0 ")
			continue
		}

		// 提交题目
		SubmitQuestions(qList, *detail)
	}

	fmt.Println("********** 程序结束 **********")
	return "ok", nil
}

func main() {
	cloudfunction.Start(Scf)
	// Scf()
	// deleteExamID("aa")
	// addExamID("ssxxx")
	// Scf()
}
