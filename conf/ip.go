package conf

import (
    "encoding/json"
    "io/ioutil"
    "path"
    "runtime"
)

var baseRoot string

func init() {
    _, filename, _, _ := runtime.Caller(1)
    baseRoot = path.Join(path.Dir(filename), "..")
}

type IPs struct {
    Regional []string
    IP       map[string][]string
}

func Parse() (map[string][]string, error) {
    file := baseRoot + "/conf/ip.json"
    fileStream, err := ioutil.ReadFile(file)
    if err != nil {
        panic(err)
    }
    var ips = IPs{}
    err = json.Unmarshal(fileStream, &ips)
    return ips.IP, err
}
