package conf

import (
    "encoding/json"
    "io/ioutil"
)

type Service struct {
    Host     string
    Port     string
    DB       int64
    Password string
    List     string
}

type Conf struct {
    Redis Service
}

func GetServiceConf() *Service {
    file := baseRoot + "/conf/service.json"
    fileStream, err := ioutil.ReadFile(file)
    if err != nil {
        panic(err)
    }
    serviceConf := Conf{}
    json.Unmarshal(fileStream, &serviceConf)
    return &serviceConf.Redis
}
