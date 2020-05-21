package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/huzhao37/gnet/examples/epms/protocols"
	"github.com/huzhao37/gnet/examples/epms/queue"
	"log"
	"sync"
	"time"

	"github.com/huzhao37/gnet"
)

type EpmsServer struct {
	*gnet.EventServer
	tick               time.Duration
	connectedSockets   sync.Map
	SystemReadQueue    queue.Queue
	SystemWriteQueue   queue.Queue
	BusinessReadQueue  queue.Queue
	BusinessWriteQueue queue.Queue
}

////on  init complete when serve ini
func (es *EpmsServer) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	log.Printf("Push server is listening on %s (multi-cores: %t, loops: %d), "+
		"pushing data every %s ...\n", srv.Addr.String(), srv.Multicore, srv.NumEventLoop, es.tick.String())
	return
}

//on opened
func (es *EpmsServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Printf("Socket with addr: %s has been opened...\n", c.RemoteAddr().String())
	es.connectedSockets.Store(c.RemoteAddr().String(), c)
	return
}

//on closed
func (es *EpmsServer) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	log.Printf("Socket with addr: %s is closing...\n", c.RemoteAddr().String())
	es.connectedSockets.Delete(c.RemoteAddr().String())
	return
}

//ticks push
func (es *EpmsServer) Tick() (delay time.Duration, action gnet.Action) {
	log.Println("It's time to push data to clients!!!")
	es.connectedSockets.Range(func(key, value interface{}) bool {
		addr := key.(string)
		c := value.(gnet.Conn)
		c.AsyncWrite([]byte(fmt.Sprintf("heart beating to %s\n", addr)))
		return true
	})
	delay = es.tick
	return
}

//on message
func (es *EpmsServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	out = frame
	return
}

func main() {
	var port int
	var multicore bool
	var interval time.Duration
	var ticker bool

	// Example command: go run epmsserver.go --port 9000 --tick 1s --multicore=true
	flag.IntVar(&port, "port", 9000, "server port")
	flag.BoolVar(&multicore, "multicore", true, "multicore")
	flag.DurationVar(&interval, "tick", 0, "pushing tick")
	flag.Parse()
	if interval > 0 {
		ticker = true
	}
	push := &EpmsServer{tick: interval}
	//自定义编解码
	encoderConfig := gnet.EncoderConfig{
		ByteOrder:                       binary.BigEndian,
		LengthFieldLength:               4,
		LengthAdjustment:                0,
		LengthIncludesLengthFieldLength: false,
	}
	decoderConfig := gnet.DecoderConfig{
		ByteOrder:           binary.BigEndian,
		LengthFieldOffset:   0,
		LengthFieldLength:   4,
		LengthAdjustment:    0,
		InitialBytesToStrip: 4,
	}
	//内部队列初始化
	SystemReadQueue := new(queue.Queue)
	SystemReadQueue.Init()

	SystemWriteQueue := new(queue.Queue)
	SystemWriteQueue.Init()

	BusinessReadQueue := new(queue.Queue)
	BusinessReadQueue.Init()

	BusinessWriteQueue := new(queue.Queue)
	BusinessWriteQueue.Init()
	log.Fatal(gnet.Serve(push, fmt.Sprintf("tcp://:%d", port), gnet.WithMulticore(multicore), gnet.WithTicker(ticker), gnet.WithCodec(gnet.NewLengthFieldBasedFrameCodec(encoderConfig, decoderConfig))))
}

func (es *EpmsServer) RegisterService(serviceName string, svc interface{}, metaData string) error {
	//todo
	return InprocessClient.Register(serviceName, svc, metaData)
}

func (es *EpmsServer) UnRegisterService(serviceName string) error {
	//todo
	return InprocessClient.Unregister(serviceName)
}

func (es *EpmsServer) SystemHandler(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) *Call {
	return InprocessClient.Go(ctx, servicePath, serviceMethod, args, reply, nil)
}

func (es *EpmsServer) BusinessHandler(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) *Call {
	return InprocessClient.Go(ctx, servicePath, serviceMethod, args, reply, nil)
}

//***add async queue for solve msg order***
//according to msg type ,write data to bussiness—read-queue  or system-read-queue
func (es *EpmsServer) HandlerRead(c gnet.Conn, ctx interface{}, data []byte) {
	//todo
	//1.unpack(协议中处理后的bytes，再次转换成thrift struct)
	epmsBody := protocols.BytesToEpmsBody(data)
	//2.get msg type
	//3.wirte msg data to queue
	//系统消息写入系统队列，业务消息写入业务队列
	if epmsBody.MsgType == protocols.NC_EPMS_HEARTBEAT {
		es.SystemReadQueue.Enqueue(data)
	} else {

	}
}

func (es *EpmsServer) HandlerWrite() {
	//todo
}

func (es *EpmsServer) SystemHandlerWrite() {
	//todo
}
