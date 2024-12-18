package toolkit

import (
	"strconv"
	"time"
)

var (
	UTC0 = time.FixedZone("UTC", 0)     // 即：UTC时区、0时区
	UTC8 = time.FixedZone("UTC", 28800) // 即：UTC8时区，东8时区[Asia/Shanghai]
)

func Now() int64 { return time.Now().Unix() }

// Utc8NowSec 获取当前UTC8时区时间戳，单位：秒
func Utc8NowSec() int64 {
	return time.Now().In(UTC8).Unix()
}

// Utc8Now 获取当前UTC8时区时间
func Utc8Now() time.Time {
	return time.Now().In(UTC8)
}

// Utc0Now 获取当前UTC0时区时间
func Utc0Now() time.Time {
	return time.Now().In(UTC0)
}

// Utc8NowMs 获取当前UTC8时区毫秒级时间戳，单位：毫秒
func Utc8NowMs() int64 {
	return time.Now().In(UTC8).UnixNano() / 1e6 // int64(time.Millisecond)
}

// RFC3339Now RFC3339标准日期
func RFC3339Now() string {
	return time.Now().In(UTC8).Format(time.RFC3339)
}

type NumSymbol uint8

const (
	Cycle       NumSymbol = iota // 举例：①、②、③
	Simplified                   // 举例：一、二、三
	Traditional                  // 举例：壹、贰、叁
)

var symbolsPool = map[NumSymbol]map[uint8]string{
	Cycle:       {1: "①", 2: "②", 3: "③", 4: "④", 5: "⑤", 6: "⑥", 7: "⑦", 8: "⑧", 9: "⑨", 10: "⑩"},
	Simplified:  {1: "一", 2: "二", 3: "三", 4: "四", 5: "五", 6: "六", 7: "七", 8: "八", 9: "九", 10: "十"},
	Traditional: {0: "零", 1: "壹", 2: "贰", 3: "叁", 4: "肆", 5: "伍", 6: "陆", 7: "柒", 8: "捌", 9: "玖", 10: "拾"},
}

// NumberWithSymbol 获取数字对应数字符号，支持 0~10
func NumberWithSymbol(num uint8, typ NumSymbol) string {
	v, ok := symbolsPool[typ][num]
	if !ok {
		return strconv.Itoa(int(num))
	}
	return v
}

type NumFormat uint8

func (n NumFormat) Cycle() string       { return NumberWithSymbol(uint8(n), Cycle) }
func (n NumFormat) Simplified() string  { return NumberWithSymbol(uint8(n), Simplified) }
func (n NumFormat) Traditional() string { return NumberWithSymbol(uint8(n), Traditional) }

const (
	Zero NumFormat = iota
	One
	Tow
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
)
