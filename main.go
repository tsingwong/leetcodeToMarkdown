/*
 * @Description:
 * @Author: Tsingwong
 * @Date: 2021-11-09 11:11:17


 */
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	USER_AGENT   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36"
	URL          = "https://leetcode-cn.com/graphql"
	CONNECTION   = "keep-alive"
	CONTENT_TYPE = "application/json"
	REFERER      = "https://leetcode-cn.com/problemset/all/"
	METHOD       = "POST"
)

type IProblem interface {
	getQuestionInfo() (Question, error)
}

type Question struct {
	Difficulty string `json:"difficulty"`
	Title      string `json:"titleCn"`
	TitleSlug  string `json:"titleSlug"`
	ID         string `json:"frontendQuestionId"`
}

type TodayRecord struct {
	Question Question `json:"question"`
}

type TodayData struct {
	TodayRecord []TodayRecord `json:"todayRecord"`
}
type TodayRes struct {
	Data TodayData `json:"data"`
}

type NumberData struct {
	ProblemsetQuestionList ProblemsetQuestionList `json:"problemsetQuestionList"`
}
type ProblemsetQuestionList struct {
	Questions []Question `json:"questions"`
}

type NumberRes struct {
	Data NumberData `json:"data"`
}

type QuestionData struct {
	Title   string   `json:"translatedTitle"`
	Content string   `json:"translatedContent"`
	Id      string   `json:"questionId"`
	Hints   []string `json:"hints"`
}
type QuestionContent struct {
	Question QuestionData `json:"question"`
}

type QuestionRes struct {
	Data QuestionContent `json:"data"`
}

type Today struct{}

func (t Today) getQuestionInfo() (Question, error) {
	query := bytes.NewBuffer([]byte(`{"query":"\n    query questionOfToday {\n  todayRecord {\n    date\n    userStatus\n    question {\n      questionId\n      frontendQuestionId: questionFrontendId\n      difficulty\n      title\n      titleCn: translatedTitle\n      titleSlug\n      paidOnly: isPaidOnly\n      freqBar\n      isFavor\n      acRate\n      status\n      solutionNum\n      hasVideoSolution\n      topicTags {\n        name\n        nameTranslated: translatedName\n        id\n      }\n      extra {\n        topCompanyTags {\n          imgUrl\n          slug\n          numSubscribed\n        }\n      }\n    }\n    lastSubmission {\n      id\n    }\n  }\n}\n    ","variables":{},"operationName":"questionOfToday"}`))
	body, err := generateHTTP(query)
	if err != nil {
		return Question{}, err
	}
	var data TodayRes
	err = json.Unmarshal(body, &data)
	if err != nil {
		return Question{}, err
	}
	return data.Data.TodayRecord[0].Question, nil
}

type Number struct {
	id int
}

func (n Number) getQuestionInfo() (Question, error) {
	query := strings.NewReader(fmt.Sprintf(`{
    "query": "query problemsetQuestionList($limit:Int,$skip:Int){problemsetQuestionList(limit:$limit skip:$skip){questions{acRate}}}",
    "variables": {
      "skip": "%v",
      "limit": "1"
    }
  }`, n.id-1))
	body, err := generateHTTP(query)
	if err != nil {
		return Question{}, err
	}
	var data NumberRes
	err = json.Unmarshal(body, &data)
	if err != nil {
		return Question{}, err
	}
	return data.Data.ProblemsetQuestionList.Questions[0], nil
}

type Text struct {
	searchText string
}

func (t Text) getQuestionInfo() (Question, error) {
	query := bytes.NewBuffer([]byte(fmt.Sprintf(`
	{
    "query": "\n    query problemsetQuestionList($categorySlug: String, $limit: Int, $skip: Int, $filters: QuestionListFilterInput) {\n  problemsetQuestionList(\n    categorySlug: $categorySlug\n    limit: $limit\n    skip: $skip\n    filters: $filters\n  ) {\n    hasMore\n    total\n    questions {\n      acRate\n      difficulty\n      freqBar\n      frontendQuestionId\n      isFavor\n      paidOnly\n      solutionNum\n      status\n      title\n      titleCn\n      titleSlug\n      topicTags {\n        name\n        nameTranslated\n        id\n        slug\n      }\n      extra {\n        hasVideoSolution\n        topCompanyTags {\n          imgUrl\n          slug\n          numSubscribed\n        }\n      }\n    }\n  }\n}\n    ",
    "variables": {
        "categorySlug": "",
        "skip": 0,
        "limit": 50,
        "filters": {
            "searchKeywords": "%v"
        }
    }
	}`, t.searchText)))
	body, err := generateHTTP(query)
	if err != nil {
		return Question{}, err
	}
	var data NumberRes
	err = json.Unmarshal(body, &data)
	if err != nil {
		return Question{}, err
	}
	for _, question := range data.Data.ProblemsetQuestionList.Questions {
		if question.ID == t.searchText || strings.Contains(question.Title, t.searchText) {
			return question, nil
		}
	}
	return Question{}, nil
}

func generateHTTP(query io.Reader) ([]byte, error) {
	req, _ := http.NewRequest("POST", URL, query)
	req.Header.Set("User-Agent", USER_AGENT)
	req.Header.Set("Connection", CONNECTION)
	req.Header.Set("Content-Type", CONTENT_TYPE)
	req.Header.Set("Referer", REFERER)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func main() {
	inputSlice := os.Args[1:]
	if len(inputSlice) == 0 {
		today := Today{}
		outputProblem(today)
	} else {
		// getByNumber OR getByString
		input := strings.Join(inputSlice, " ")
		num, err := strconv.Atoi(input)
		if err != nil {
			text := Text{searchText: input}
			outputProblem(text)
			return
		}
		number := Number{id: num}
		outputProblem(number)
	}
}

// func getByNumber(num int) (Question, error) {

// }

func getProblemContent(slug string) (string, error) {
	query := bytes.NewBuffer([]byte(fmt.Sprintf(`{"operationName":"questionData","variables":{"titleSlug":"%s"},"query":"query questionData($titleSlug: String!) {\n  question(titleSlug: $titleSlug) {\n    questionId\n    questionFrontendId\n    categoryTitle\n    boundTopicId\n    title\n    titleSlug\n    content\n    translatedTitle\n    translatedContent\n    isPaidOnly\n    difficulty\n    likes\n    dislikes\n    isLiked\n    similarQuestions\n    contributors {\n      username\n      profileUrl\n      avatarUrl\n      __typename\n    }\n    langToValidPlayground\n    topicTags {\n      name\n      slug\n      translatedName\n      __typename\n    }\n    companyTagStats\n    codeSnippets {\n      lang\n      langSlug\n      code\n      __typename\n    }\n    stats\n    hints\n    solution {\n      id\n      canSeeDetail\n      __typename\n    }\n    status\n    sampleTestCase\n    metaData\n    judgerAvailable\n    judgeType\n    mysqlSchemas\n    enableRunCode\n    envInfo\n    book {\n      id\n      bookName\n      pressName\n      source\n      shortDescription\n      fullDescription\n      bookImgUrl\n      pressImgUrl\n      productUrl\n      __typename\n    }\n    isSubscribed\n    isDailyQuestion\n    dailyRecordStatus\n    editorType\n    ugcQuestionId\n    style\n    exampleTestcases\n    __typename\n  }\n}\n"}
  `, slug)))
	body, err := generateHTTP(query)
	if err != nil {
		return "", err
	}
	var data QuestionRes
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}
	return data.Data.Question.Content, nil
}

func outputProblem(obj IProblem) {
	data, err := obj.getQuestionInfo()
	if err != nil {
		log.Fatal(fmt.Sprintln("发生了一些错误，具体信息是: ", err))
	}
	log.Printf("当前题目为：%s", data.Title)
	log.Print("当前题目 ID 为：", data.ID)
	log.Printf("难度为：%s", data.Difficulty)
	log.Printf("地址为：%s", "https://leetcode-cn.com/problems/"+data.TitleSlug)
	fmt.Println("=====================")
	content, err := getProblemContent(data.TitleSlug)
	if err != nil {
		log.Fatal(fmt.Sprintln("发生了一些错误，具体信息是: ", err))
	}

	preReg := regexp.MustCompile(`<pre>[\s\S]*?</pre>`)
	reg := regexp.MustCompile(`<[^>p]*>`)
	content = preReg.ReplaceAllStringFunc(content, func(src string) string {
		return reg.ReplaceAllString(src, ``)
	})
	removePattern := []string{"</?p>", "</?ul>", "</?ol>", "</li>", "</sup>"}
	for _, pattern := range removePattern {
		reg := regexp.MustCompile(pattern)
		content = reg.ReplaceAllString(content, "")
	}
	replacePattern := [][]string{
		{"&nbsp;", " "},
		{"&quot;", `"`},
		{"&lt;", "<"},
		{"&gt;", ">"},
		{"&le;", "≤"},
		{"&ge;", "≥"},
		{"<sup>", "^"},
		{"&#39", "'"},
		{"&times;", "x"},
		{"&ldquo;", "“"},
		{"&rdquo;", "”"},
		{" *<strong> *", " **"},
		{" *</strong> *", "** "},
		{" *<code> *", " `"},
		{" *</code> *", "` "},
		{"<pre>", "```\n"},
		{"</pre>", "\n```\n"},
		{"<em> *</em>", ""},
		{" *<em> *", " *"},
		{" *</em> *", "* "},
		{"</?div.*?>", ""},
		{"	*</?li>", "- "},
		{`(\s\n\s*)+`, "\n"},
	}
	for _, pattern := range replacePattern {
		reg := regexp.MustCompile(pattern[0])
		content = reg.ReplaceAllString(content, pattern[1])
	}
	content = fmt.Sprintf(`## %v.%s
  `, data.ID, data.Title) + content
	fmt.Println(content)
	cmd := exec.Command("pbcopy")
	str, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer str.Close()
		io.WriteString(str, content)
	}()
	cmd.CombinedOutput()
	fmt.Println("当前题目已经被保存到粘贴板中...")
}
