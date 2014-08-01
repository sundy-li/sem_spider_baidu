package service

import (
    "code.google.com/p/go.net/html"
    "encoding/json"
    "fmt"
    "github.com/PuerkitoBio/goquery"
    "github.com/robertkrimen/otto"
    "github.com/vmihailenco/redis"
    "io/ioutil"
    "net/http"
    "net/url"
    "regexp"
    "runtime"
    "sem_spider_baidu/conf"
    "sem_spider_baidu/utils"
    "strings"
    "sunteng/commons/log"
    "sync"
    "time"
)

var logger = log.Get("spider")

var client *redis.Client
var ips map[string][]string

var list string

func init() {
    serviceConf := conf.GetServiceConf()
    list = serviceConf.List
    host := serviceConf.Host
    port := serviceConf.Port
    password := serviceConf.Password
    DB := serviceConf.DB
    client = redis.NewTCPClient(host+":"+port, password, DB)

    if client.Ping().Err() != nil {
        logger.Errorf("%#v\n", client.Ping().Err())
    } else {
        fmt.Println("ping ok")
    }
    ips, _ = conf.Parse()
}

//queueName 队列名称
func Querys(queueName string) {
    //默认使用配置中的队列
    queueName = list
    var i int = 0
    runtime.GOMAXPROCS(runtime.NumCPU())
    for {
        strReq := client.RPop(queueName)
        if strReq.Err() != nil {
            continue
        }
        fmt.Printf("value %#v\n", strReq.Val())

        wordRangional := strReq.Val()
        strs := strings.Split(wordRangional, "@")
        word := strs[0]
        rangional := strs[1]

        listIP, ok := ips[rangional]
        if !ok {
            client.Set(wordRangional, "rangional code no exist")
            continue
        }
        randNumber := int(time.Now().Unix())
        ProxyIP := listIP[randNumber%len(listIP)]
        url := getUrl(word)
        go query(url, ProxyIP, wordRangional, i)

        time.Sleep(time.Second * 1)
        fmt.Printf("num goroutine -->%d \n", runtime.NumGoroutine())
    }

}

func query(url string, ProxyIP string, wordRangional string, i int) {
    rank, err := Search(url, ProxyIP)
    if err != nil {
        // fmt.Printf("search err : %#v\n", err)
        return
    }
    rankByte, _ := json.Marshal(rank)
    rankStr := string(rankByte)
    fmt.Printf("key:%s\tvalue:%s\n", wordRangional, rankStr)
    client.Set(wordRangional, rankStr)
}

func getUrl(word string) string {
    search := "http://www.baidu.com/s?tn=baidu&ie=utf-8&wd="
    words := EnscapeWords(word)
    return search + words
}

func getPage(url string) {
    start := getTime()
    client := http.Client{}
    resp, _ := client.Get(url)
    body, _ := ioutil.ReadAll(resp.Body)
    fmt.Println(string(body))
    defer resp.Body.Close()
    end := getTime()
    println("use time", end-start)
}

func Search(url string, ProxyIP string) (map[string][]string, error) {
    rank := map[string][]string{}

    start := getTime()
    resp, err := GetByProxy(url, ProxyIP)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    t := time.Now().Unix()
    document, err := goquery.NewDocumentFromResponse(resp)
    e := time.Now().Unix()
    println("go query ", e-t)
    if err != nil {
        return nil, err
    }
    startParse := time.Now().Unix()

    var waitGroup sync.WaitGroup

    waitGroup.Add(1)
    waitGroup.Add(1)
    go func() {
        left := document.Find("div#content_left")
        leftPage := getLeft(left)
        rank["left"] = leftPage
        waitGroup.Done()
    }()

    go func() {
        right := document.Find("div#content_right")
        rightPage := getRight(right)
        rank["right"] = rightPage
        waitGroup.Done()
    }()
    waitGroup.Wait()
    // fmt.Printf("%v\n", leftPage)
    end := getTime()
    println("parse Time ", end-startParse)
    println("use time", end-start)
    return rank, nil
}

func getLeft(left *goquery.Selection) []string {
    left_page := []string{}
    left.Find("[href^='http://www.baidu.com/baidu.php']").Each(func(i int, this *goquery.Selection) {
        if this.Parent().Is("h3") || this.Children().First().Is("[style='display:inline-block;text-decoration:underline;']") {
            href, _ := this.Attr("href")
            left_page = append(left_page, href)
        }
    })
    if len(left_page) < 1 {
        return left_page
    }

    var flag = make(map[string]bool)
    var index = 0
    for _, url := range left_page {
        if !flag[url] {
            left_page[index] = url
            index = index + 1
        }
        flag[url] = true
    }

    left_page = left_page[:index]

    var t1 = utils.NewTime()
    runtime.GOMAXPROCS(runtime.NumCPU())
    var waitGroup sync.WaitGroup
    for i, link := range left_page {
        waitGroup.Add(1)
        fmt.Printf("left add \n", link)
        go func(i int, link string) {
            left_page[i] = getRealUrl(i, link)
            waitGroup.Done()
        }(i, link)
    }
    waitGroup.Wait()

    fmt.Printf("%d -> left cost\n", t1.Cost())
    return left_page
}

func getRight(right *goquery.Selection) []string {
    ec_im := right.Find("div#ec_im_container")
    script, _ := ec_im.Find("script").First().Html()
    script = html.UnescapeString(script)
    reg := regexp.MustCompile(`var[\s]*?g[\s]*?=[\s]*?function\(jDomain\)[\s\S]*?return[\s]*?i[\s]*?};`)
    scriptfung := reg.FindString(script)
    right_page := make([]string, 0, 10)
    vm := otto.New()

    vm.Run(scriptfung)
    logger.Debugf("%s\n", scriptfung)
    ec_im.Find("[id^='bdfs']").Each(func(i int, this *goquery.Selection) {
        href, _ := this.Children().First().Attr("href")
        vm.Set("url", href)
        vm.Run(`url = a(url);`)
        value, _ := vm.Get("url")
        href, err := value.ToString()
        if err != nil {
            logger.Debugf("%s\n", err.Error())
        }
        right_page = append(right_page, href)
    })

    if len(right_page) < 1 {
        return right_page
    }

    var t1 = utils.NewTime()

    logger.Debugf("%#v\n", right_page)
    var waitGroup sync.WaitGroup
    for i, link := range right_page {
        waitGroup.Add(1)
        go func(i int, link string) {
            right_page[i] = getRealUrl(i, link)
            waitGroup.Done()
        }(i, link)
    }
    waitGroup.Wait()
    logger.Debug(t1.String())
    return right_page
}

/**
 * get realUrl
 */
func getRealUrl(i int, link string) string {
    logger.Debugf(" link%d- > %s \n", i, link)
    var str = link

    var client = http.DefaultClient
    resp, err := client.Get(link)

    // println(s[0], link)

    if err != nil {
        logger.Errorf("url - %s  : %s", link, err.Error())
        return link
    }
    defer resp.Body.Close()
    if resp.StatusCode == 200 {
        str = resp.Request.URL.String()
        var index = strings.Index(str, `?`)
        if index > 0 {
            str = str[:index]
        }
    } else {
        str = link
    }
    return str
}

func getTime() int64 {
    return time.Now().Unix()
}

/**
* 连接并编码查询字符
**/
func EnscapeWords(words ...string) string {
    str := ""
    for _, word := range words {
        str = str + " " + word
    }
    str = str[1:]
    str = url.QueryEscape(str)
    return str
}
