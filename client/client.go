package client

import (
    "encoding/json"
    "errors"
    "github.com/vmihailenco/redis"
    "io/ioutil"
    "path"
    "runtime"
    "strconv"
    "time"
)

type Client struct {
    Host     string
    Port     string
    DB       int64
    Password string
    List     string
}

type Conf struct {
    Redis Client
}

var client *redis.Client
var list string

func init() {
    _, filename, _, _ := runtime.Caller(1)
    baseRoot := path.Join(path.Dir(filename), ".")
    file := baseRoot + "/client.json"
    fileStream, err := ioutil.ReadFile(file)
    if err != nil {
        panic(err)
    }
    clientConf := Conf{}
    json.Unmarshal(fileStream, &clientConf)
    client = redis.NewTCPClient(clientConf.Redis.Host+":"+clientConf.Redis.Port, clientConf.Redis.Password, clientConf.Redis.DB)
    list = clientConf.Redis.List
}

func IsConn() bool {
    statusReq := client.Ping()
    if statusReq.Err() != nil {
        return false
    }
    return true
}

func Query(keyWord string, rangionalCode string) (string, error) {
    if !IsConn() {
        return "", errors.New("not conn redis")
    }
    randNumber := strconv.FormatInt(time.Now().Unix(), 10)
    key := keyWord + "@" + rangionalCode + "@" + randNumber
    client.LPush(list, key)
    var value string
    var sleepTime time.Duration
    //斐波那契数列休眠
    var i, j time.Duration
    i = 1
    j = 1
    for {
        if i >= 50 {
            break
        }
        sleepTime = i
        time.Sleep(time.Second * sleepTime)
        i = j + i
        j = i - j
        //一次使用
        stringReq := client.Get(key)
        if stringReq.Err() == nil {
            value = stringReq.Val()
            client.Del(key)
            break
        }
    }
    if value == "" {
        return value, errors.New("time out")
    }
    return value, nil
}
