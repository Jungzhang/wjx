package main

import (
	"math/rand"
	"time"
	"strconv"
	"os"
	"log"
	"strings"
	"net/http"
	"io/ioutil"
	"bufio"
	"io"
	"net/url"
	"fmt"
	"encoding/json"
	"crypto/tls"
)

type AnswerFileContent struct {
	No              string   // 题号
	ExceptAnswer    []string // 期望答案列表, 保存的为字符串 1、2、3 对应 A、B、C
	NotExceptAnswer []string // 非期望答案列表，保存的为字符串 1、2、3 对应 A、B、C
}

var answerFileContent []AnswerFileContent
var ip string

// 获取代理ip, 要求代理ip接口返回至少包含如下字段
/*
{
    "ip": "192.168.1.109:8080"
}
*/
func getIpProxy(proxyUrl string) string {

	if proxyUrl == "" {
		return ""
	}

	// 获取100次代理ip, 选择可用的ip进行代理添加
	for i := 0; i < 100; i++ {
		resp, err := http.DefaultClient.Get(proxyUrl)
		if err != nil {
			break
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			break
		}

		data := make(map[string]interface{}, 0)
		if err = json.Unmarshal(body, &data); err != nil {
			break
		}

		// 测试代理ip是否可用
		urlI := url.URL{}
		urlProxy, _ := urlI.Parse("http://" + data["ip"].(string))

		transport := http.Transport{}
		transport.Proxy = http.ProxyURL(urlProxy)
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		client := &http.Client{}
		client.Timeout = time.Second * 3
		client.Transport = &transport
		_, err = client.Get("http://example.com")
		if err != nil {
			continue
		}
		return data["ip"].(string)
	}
	log.Println("获取可用代理失败")

	return ""
}

// 初始化答案模板
func initAnswerFileContent(path string) error {

	// 打开答案模板文件
	fi, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fi.Close()
	answerFileContent = make([]AnswerFileContent, 0)

	br := bufio.NewReader(fi)
	for {
		// 读取一行答案模板
		a, _, err := br.ReadLine()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// 解析答案模板
		answers := strings.Split(string(a), " ")

		// 获取答案范围
		answerRange, err := strconv.Atoi(answers[2])
		if err != nil {
			return err
		}

		// 初始化一个map用来保存答案全集
		// 例如:模板中第三项为5,则答案全集为A、B、C、D、E,即该map中key为A、B、C、D、E
		answerMap := make(map[string]bool, 0)
		for i := 65; i < answerRange+65; i++ {
			answerMap[string(i)] = false
		}

		// 解析期望的答案
		exceptAnswer := make([]string, 0)
		// 循环遍历期望答案中每一个答案选项, 并将上边初始化的map相应key中value置为true,用来表示该选项为期望选项
		for _, exceptAnswerItem := range answers[1] {
			answerMap[string(exceptAnswerItem)] = true
			// 保存的为字符串 1、2、3 对应 A、B、C
			exceptAnswer = append(exceptAnswer, strconv.Itoa(int(exceptAnswerItem)-64))
		}
		// 解析不期望的答案
		notExceptAnswer := make([]string, 0)
		for key, used := range answerMap {
			if !used {
				// 保存的为字符串 1、2、3 对应 A、B、C
				notExceptAnswer = append(notExceptAnswer, strconv.Itoa(int(key[0])-64))
			}
		}

		// 合成答案模板
		answerItem := AnswerFileContent{No: answers[0], ExceptAnswer: exceptAnswer, NotExceptAnswer: notExceptAnswer}
		answerFileContent = append(answerFileContent, answerItem)
	}

	return nil
}

// 获取问卷页面
func getWJPage(homeUrl string) ([]byte, error) {

	req, _ := http.NewRequest("GET", homeUrl, strings.NewReader(""))
	req.Header.Add("Host", "https://www.wjx.cn")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.86 Safari/537.36")
	req.Header.Add("DNT", "1")
	transport := http.Transport{}
	if ip != "" {
		urlI := url.URL{}
		urlProxy, _ := urlI.Parse("http://" + ip)
		transport.Proxy = http.ProxyURL(urlProxy)
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := http.Client{}
	client.Timeout = 10 * time.Second
	if ip != "" {
		client.Transport = &transport
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	page, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return page, nil
}

// 获取本次需要提交的答案
func getAnswer(i int) (string, string, error) {

	ret := ""
	printAnswer := ""

	answerMap := make(map[string]string, 0)
	for i := int('1'); i <= int('9'); i++ {
		answerMap[string(i)] = string(int('A') - int('1') + i)
	}

	// 从答案模板中获取答案
	for j, answer := range answerFileContent {
		// 如果已经走了10轮且当前有不期望的答案, 则选择不期望的答案
		if i%10 == 0 && i != 0 && len(answer.NotExceptAnswer) != 0 {
			rs := rand.NewSource(time.Now().UnixNano())
			randObj := rand.New(rs)
			answerIdx := randObj.Intn(len(answer.NotExceptAnswer))
			if j == len(answerFileContent)-1 {
				ret = ret + answer.No + "$" + answer.NotExceptAnswer[answerIdx]
			} else {
				ret = ret + answer.No + "$" + answer.NotExceptAnswer[answerIdx] + "}"
			}
			if printAnswer != "" {
				printAnswer = printAnswer + "; 第" + answer.No + "题:" + answerMap[answer.NotExceptAnswer[answerIdx]]
			} else {
				printAnswer = printAnswer + "第" + answer.No + "题:" + answerMap[answer.NotExceptAnswer[answerIdx]]
			}
		} else {
			// 获取期望的答案
			n := len(answer.ExceptAnswer)
			rs := rand.NewSource(time.Now().UnixNano())
			randObj := rand.New(rs)
			answerIdx := randObj.Intn(n)
			if j == len(answerFileContent)-1 {
				ret = ret + answer.No + "$" + answer.ExceptAnswer[answerIdx]
			} else {
				ret = ret + answer.No + "$" + answer.ExceptAnswer[answerIdx] + "}"
			}
			if printAnswer != "" {
				printAnswer = printAnswer + "; 第" + answer.No + "题:" + answerMap[answer.ExceptAnswer[answerIdx]]
			} else {
				printAnswer = printAnswer + "第" + answer.No + "题:" + answerMap[answer.ExceptAnswer[answerIdx]]
			}
		}
		time.Sleep(time.Microsecond)
	}

	return url.QueryEscape(ret), printAnswer, nil
}

// 获取答案post的地址
func getPostUrl(baseUrl string, id int) (string, error) {

	pageInfo, err := getWJPage(baseUrl + "/jq/" + strconv.Itoa(id) + ".aspx")
	if err != nil {
		return "", err
	}

	return baseUrl + "/joinnew/processjq.ashx?" + getPostUrlParam(pageInfo, id), nil

}

// 获取Post答案时的url参数
func getPostUrlParam(pageInfo []byte, id int) string {

	rs := rand.NewSource(time.Now().UnixNano())
	randObj := rand.New(rs)
	ktimes := randObj.Intn(40)
	nowTime := time.Now().UnixNano() / int64(time.Millisecond)
	starttime, rn, jqnonce := getPageElem(pageInfo)
	jqsign := getJqsign(jqnonce, ktimes)

	return "submittype=1" + "&hlv=1" + "&ktimes=" + strconv.Itoa(ktimes) +
		"&curID=" + strconv.Itoa(id) + "&t=" + strconv.FormatInt(nowTime, 10) +
		"&starttime=" + url.QueryEscape(starttime) + "&rn=" + rn + "&jqnonce=" + jqnonce + "&jqsign=" + url.QueryEscape(jqsign)
}

// 获取页面中需要用到的元素
func getPageElem(page []byte) (starttime, rn, jqnonce string) {
	pageInfo := string(page)
	starttime = getStringFromPage("starttime", &pageInfo)
	rn = getStringFromPage("rndnum", &pageInfo)
	jqnonce = getStringFromPage("jqnonce", &pageInfo)

	return
}

// 根据传入的key获取对应页面元素(纯字符串匹配,无任何外部依赖)
func getStringFromPage(name string, pageInfo *string) string {
	findElem := "var " + name + "="
	firstIdx := strings.Index(*pageInfo, findElem)
	if firstIdx == -1 {
		return ""
	}
	startIdx := firstIdx + len(findElem)
	endIdx := startIdx
	firstMeet := true
	for i := startIdx; i < len(*pageInfo); i++ {
		if string((*pageInfo)[i]) == "\"" {
			if firstMeet {
				firstMeet = false
			} else {
				endIdx = i
				break
			}
		}
	}
	return (*pageInfo)[startIdx+1 : endIdx]
}

// 获取POST答案时需要用到的jqsign字段
func getJqsign(jqnonce string, ktimes int) string {

	result := make([]string, 0)
	salt := ktimes % 10

	if salt == 0 {
		salt = 1
	}

	for _, c := range jqnonce {
		tmp := string(int(c) ^ salt)
		result = append(result, tmp)
	}

	return strings.Join(result, "")
}

// 提交答案
func postAnswer(postUrl, answer string, id int) error {

	req, _ := http.NewRequest("POST", postUrl, strings.NewReader("submitdata="+answer))
	req.Header.Add("Referer", "https://www.wjx.cn/jq/"+strconv.Itoa(id)+".aspx")
	req.Header.Add("Origin", "https://www.wjx.cn")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.86 Safari/537.36")
	req.Header.Add("DNT", "1")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	transport := http.Transport{}
	if ip != "" {
		urlI := url.URL{}
		urlProxy, _ := urlI.Parse("http://" + ip)
		transport.Proxy = http.ProxyURL(urlProxy)                         // set proxy
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //set ssl
	}

	client := http.Client{}
	client.Timeout = 10 * time.Second
	if ip != "" {
		client.Transport = &transport
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	if strings.Index(string(respBody), "jidx=") > -1 {
		return nil
	} else {
		return fmt.Errorf("提交答案失败, body:%s", string(respBody))
	}
}

func main() {

	if len(os.Args) != 4 || len(os.Args) != 5 {
		log.Fatal("参数错误，使用方法：./" + os.Args[0] + "  提交答案数量  " + "  试卷id" + "  答案模板  " + " 代理接口地址(可选)")
	}

	counter, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("参数错误，使用方法：./" + os.Args[0] + "  提交答案数量  " + "  试卷id" + " 答案模板" + " 代理接口地址(可选)")
	}

	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal("参数错误，使用方法：./" + os.Args[0] + "  提交答案数量  " + "  试卷id" + " 答案模板" + " 代理接口地址(可选)")
	}

	err = initAnswerFileContent(os.Args[3])
	if err != nil {
		log.Fatal("初始化答案失败,致命错误:" + err.Error() + "使用方法：./" + os.Args[0] + "  数量  " + "  试卷id" + " 答案模板" + " 代理接口地址(可选)")
	}

	proxyUrl := ""
	if len(os.Args) == 5 {
		proxyUrl = os.Args[4]
	}

	waitTime := time.Duration(0)

	for i := 0; i < counter; i++ {
		time.Sleep(waitTime * time.Second)
		ip = getIpProxy(proxyUrl)
		printIp := ip
		if printIp == "" {
			printIp = "本机IP"
		}
		log.Println("第 " + strconv.Itoa(i) + " 次提交, 使用 " + printIp + " 访问")

		answer, printAnswer, err := getAnswer(i)
		if err != nil {
			log.Println("获取答案失败:", err.Error())
			i--
			continue
		}
		postUrl, err := getPostUrl("https://www.wjx.cn", id)
		if err != nil {
			log.Println("获取提交地址失败:", err.Error())
			i--
			continue
		}
		err = postAnswer(postUrl, answer, id)
		if err != nil {
			log.Println("提交第 "+strconv.Itoa(i)+" 次答案失败: ", err.Error())
			i--
		} else {
			if i%10 == 0 && i != 0 {
				log.Println("第 " + strconv.Itoa(i) + " 次提交成功  本次使用非期望答案  提交答案: " + printAnswer + "   使用ip为: " + printIp)
			} else {
				log.Println("第 " + strconv.Itoa(i) + " 次提交成功  本次使用期望答案  提交答案: " + printAnswer + "   使用ip为: " + printIp)
			}
		}
		// 随机等待0-10秒
		rs := rand.NewSource(time.Now().UnixNano())
		randObj := rand.New(rs)
		waitTime = time.Duration(randObj.Intn(10))
	}
}

