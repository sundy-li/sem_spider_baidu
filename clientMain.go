package main

import (
    "fmt"
    "sem_spider_baidu/client"
)

func main() {
    value, err := client.Query("武易", "1001")
    fmt.Printf("err : %#v\n", err)
    println(value)
}
