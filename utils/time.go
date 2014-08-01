package utils

import (
    "fmt"
    "time"
)

type TimeLog struct {
    start int64
}

func NewTime() *TimeLog {
    return &TimeLog{
        start: time.Now().Unix(),
    }
}

func (t *TimeLog) Init() {
    t.start = time.Now().Unix()
}
func (t *TimeLog) Cost() int {
    return int(time.Now().Unix() - t.start)
}

func (t *TimeLog) String() string {
    return fmt.Sprintf("cost ----- %d seconds----", t.Cost())
}
