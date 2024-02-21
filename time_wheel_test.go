package toolkit

import (
	"fmt"
	"testing"
	"time"
)

func TestNewTimewheel(t *testing.T) {
	tw := NewTimewheel()
	defer tw.Stop()

	w1 := NewWheel("Train 北京->上海", Single, []func(){
		func() { fmt.Println("北京始发") },
		func() { fmt.Println("济南") },
		func() { fmt.Println("徐州") },
		func() { fmt.Println("南京") },
		func() { fmt.Println("终到上海") },
	}, 1*time.Second)
	w2 := NewWheel("Train 上海->北京", Single, []func(){
		func() { fmt.Println("上海始发") },
		func() { fmt.Println("南京") },
		func() { fmt.Println("徐州") },
		func() { fmt.Println("济南") },
		func() { fmt.Println("终到北京") },
	}, 2*time.Second)
	w3 := NewWheel("定时邮件", Single, []func(){
		func() { fmt.Println("已发送给@Tony") },
		func() { fmt.Println("已发送给@Pony") },
		func() { fmt.Println("已发送给@Alice") },
	}, 3*time.Second)

	go tw.Run()

	go func() {
		for i := 0; i < 2; i++ {
			seg1, _ := tw.Add(w1)
			seg2, _ := tw.Add(w2)
			seg3, _ := tw.Add(w3)

			println("seg:", seg1, seg2, seg3)

			time.Sleep(2 * time.Second)
		}
	}()

	time.Sleep(10 * time.Second)

	fmt.Println("重播")
	tw.Replay(10, 1000)

	time.Sleep(time.Second)

}
