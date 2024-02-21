package toolkit

import (
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// 时间轮最小化实现

const Unit = 100 * time.Millisecond // 单位

type Segment int64 // 时间片类型

type Type uint8 // 类型

const (
	_      Type = iota
	Single      // 单次任务
	Limit       // 需要执行多少次
	Loop        // 循环
)

type Wheeler interface {
	Called() string
	Before()
	Trigger()
	After()
	Time() time.Duration // 设置时间，多久后执行
}

type Wheel struct {
	Name  string
	Type  Type
	Pool  []func()
	Delay time.Duration
}

func NewWheel(name string, typ Type, fn []func(), delay time.Duration) *Wheel {
	if len(fn) == 0 {
		return nil
	}
	return &Wheel{
		Name:  name,
		Type:  typ,
		Pool:  fn,
		Delay: delay,
	}
}

// Called 轮子名称
func (w *Wheel) Called() string {
	return w.Name
}

func (w *Wheel) Before() {}
func (w *Wheel) After()  {}

// Trigger 触发任务
func (w *Wheel) Trigger() {
	for _, fn := range w.Pool {
		fn()
	}
}

func (w *Wheel) Time() time.Duration {
	return w.Delay
}

const (
	Replay int32 = iota // 回放
	Normal              // 正常
	Pause               // 暂停
)

type Timewheel struct {
	mu           sync.RWMutex
	state        atomic.Int32          // 状态
	Offset       int64                 // 偏移量
	buckets      map[Segment][]Wheeler // 轮子集合
	nameSeg      map[string][]Segment  // 名称映射时间段
	BaseTimeline time.Duration         // 时间基线
	Ticker       *time.Ticker          // 定时器
}

func NewTimewheel() *Timewheel {
	tw := &Timewheel{
		buckets:      make(map[Segment][]Wheeler, 1<<10),
		nameSeg:      make(map[string][]Segment, 1<<10),
		BaseTimeline: UnixMilli() / Unit * Unit,
		Ticker:       time.NewTicker(Unit),
	}
	tw.state.Store(Normal)
	return tw
}

// Add 添加转轮任务
func (tw *Timewheel) Add(w Wheeler) (Segment, error) {
	seg := RelativeSeg(UnixMilli()+w.Time(), tw.BaseTimeline)
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if _, ok := tw.buckets[seg]; !ok {
		tw.buckets[seg] = make([]Wheeler, 0, 1<<8)
		tw.nameSeg[w.Called()] = make([]Segment, 0, 1<<8)
	}
	if len(tw.buckets[seg]) == cap(tw.buckets[seg]) {
		return seg, errors.New("current segment list in buckets is full")
	}
	tw.buckets[seg] = append(tw.buckets[seg], w)
	tw.nameSeg[w.Called()] = append(tw.nameSeg[w.Called()], seg)
	return seg, nil
}

// Remove 移除转轮任务
func (tw *Timewheel) Remove(name string) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	segs, ok := tw.nameSeg[name]
	if !ok {
		return
	}
	for _, seg := range segs {
		for i, wheel := range tw.buckets[seg] {
			if wheel.Called() == name {
				buf := make([]Wheeler, len(tw.buckets[seg])-1)
				copy(buf[:i], tw.buckets[seg][:i])
				copy(buf[i:], tw.buckets[seg][i+1:])
				tw.buckets[seg] = buf
			}
		}
	}
	delete(tw.nameSeg, name)
}

// Replay 重播
func (tw *Timewheel) Replay(from, to Segment) {
	if from >= to {
		return // nothing to do
	}

	tw.state.Swap(Replay)
	defer tw.Play()

	tw.mu.Lock()
	defer tw.mu.Unlock()

	replay := make([]Segment, 0, 1<<8)
	for segment := range tw.buckets {
		if segment >= from && segment <= to {
			replay = append(replay, segment)
		}
	}
	sort.SliceStable(replay, func(i, j int) bool { return replay[i] < replay[j] })

	for _, segment := range replay {
		for _, wheeler := range tw.buckets[segment] {
			wheeler.Before()
			wheeler.Trigger()
			wheeler.After()
		}
	}
}

// Play 播放
func (tw *Timewheel) Play() {
	tw.state.Swap(Normal)
}

// Pause 暂停
func (tw *Timewheel) Pause() {
	tw.state.Swap(Pause)
}

func (tw *Timewheel) Run() {
	for {
		if !tw.IsNormal() {
			time.Sleep(Unit)
			continue
		}
		select {
		case <-tw.Ticker.C:
			ws := tw.Wheelers(RelativeSeg(UnixMilli(), tw.BaseTimeline))
			for _, w := range ws {
				w.Before()
				w.Trigger()
				w.After()
			}
		}
	}
}

func (tw *Timewheel) IsNormal() bool {
	return tw.state.Load() == Normal
}

func (tw *Timewheel) Wheelers(seg Segment) []Wheeler {
	if !tw.IsNormal() {
		return nil
	}
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if ws, ok := tw.buckets[seg]; ok {
		return ws
	}
	return nil
}

func (tw *Timewheel) Stop() {
	tw.Pause()
	tw.Ticker.Stop()
	tw.buckets, tw.nameSeg = nil, nil
}

// UnixMilli 当前毫秒时间戳
func UnixMilli() time.Duration {
	return time.Duration(time.Now().UnixMilli()) * time.Millisecond
}

// RelativeSeg 时间片相对基线时间位置
func RelativeSeg(target, base time.Duration) Segment {
	duration := target - base
	seg := Segment(duration / Unit)
	if duration%Unit > 0 {
		seg += 1
	}
	return seg
}
