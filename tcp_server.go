package toolkit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrServerOverload = errors.New("server overloaded")
	ErrNotFoundConn   = errors.New("cannot find connection")
)

type Message struct {
	Data []byte
}

type Switcher interface {
	Read(ctx context.Context)
	Write(ctx context.Context, msg Message)
}

type Processor struct {
	net.Conn
	receive chan Message // 接收的数据
	send    chan Message // 发送的数据
	Frame   Frame
}

func NewProcessor(conn net.Conn, chanCap int, frame Frame) *Processor {
	return &Processor{
		Conn:    conn,
		receive: make(chan Message, chanCap),
		send:    make(chan Message, chanCap),
		Frame:   frame,
	}
}

func (p *Processor) Read(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			_, _ = fmt.Fprint(os.Stderr, "read context exit")
			return
		default:
			header := make([]byte, p.Frame.header)
			_, err := io.ReadFull(p.Conn, header)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "read conn err:%s", err.Error())
				return
			}
		}
	}
}

// Frame 数据帧约束条件
type Frame struct {
	bigEndian bool  // 是否为大端模式
	header    int32 // 字节流头部占用字节数
	seqSize   int32 // 消息序号占用字节数
}

func NewFrame(bigEndian bool, header, seqSize int32) *Frame {
	return &Frame{bigEndian: bigEndian, header: header, seqSize: seqSize}
}

type Server struct {
	wg        *sync.WaitGroup
	mu        sync.RWMutex
	conns     map[int32]net.Conn
	heartbeat time.Duration
	connNum   int32
	maxNum    int32
}

func NewServer(maxConnNum int32, heartbeat time.Duration) *Server {
	return &Server{
		wg:        &sync.WaitGroup{},
		conns:     make(map[int32]net.Conn),
		connNum:   maxConnNum,
		heartbeat: heartbeat,
	}
}

// RunGoroutine 运行协程
func (s *Server) RunGoroutine(fn func()) {
	s.wg.Add(1)
	defer s.wg.Done()
	fn()
}

// WaitGoroutine 等待协程退出
func (s *Server) WaitGoroutine() {
	s.wg.Wait()
}

// SetConn 设置新的连接
func (s *Server) SetConn(conn net.Conn) (int32, error) {
	if s.Overload() {
		return -1, ErrServerOverload
	}
	newNum := atomic.AddInt32(&s.connNum, 1)
	s.mu.Lock()
	s.conns[newNum] = conn
	s.mu.Unlock()
	return newNum, nil
}

// GetConn 获取连接
func (s *Server) GetConn(connId int32) (net.Conn, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.conns[connId]; !ok {
		return nil, ErrNotFoundConn
	}
	return s.conns[connId], nil
}

func (s *Server) CloseConn(connId int32) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.conns[connId]; !ok {
		return nil
	}
	return s.conns[connId].Close()
}

// Overload 连接数超过负载
func (s *Server) Overload() bool {
	return atomic.LoadInt32(&s.connNum) >= s.maxNum
}
