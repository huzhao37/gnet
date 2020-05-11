package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/huzhao37/gnet"
)

type EpmsServer struct {
	*gnet.EventServer
	tick             time.Duration
	connectedSockets sync.Map
}

////on  init complete when serve ini
func (ps *EpmsServer) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	log.Printf("Push server is listening on %s (multi-cores: %t, loops: %d), "+
		"pushing data every %s ...\n", srv.Addr.String(), srv.Multicore, srv.NumEventLoop, ps.tick.String())
	return
}

//on opened
func (ps *EpmsServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Printf("Socket with addr: %s has been opened...\n", c.RemoteAddr().String())
	ps.connectedSockets.Store(c.RemoteAddr().String(), c)
	return
}

//on closed
func (ps *EpmsServer) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	log.Printf("Socket with addr: %s is closing...\n", c.RemoteAddr().String())
	ps.connectedSockets.Delete(c.RemoteAddr().String())
	return
}

//ticks push
func (ps *EpmsServer) Tick() (delay time.Duration, action gnet.Action) {
	log.Println("It's time to push data to clients!!!")
	ps.connectedSockets.Range(func(key, value interface{}) bool {
		addr := key.(string)
		c := value.(gnet.Conn)
		c.AsyncWrite([]byte(fmt.Sprintf("heart beating to %s\n", addr)))
		return true
	})
	delay = ps.tick
	return
}

//on message
func (ps *EpmsServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
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

	log.Fatal(gnet.Serve(push, fmt.Sprintf("tcp://:%d", port), gnet.WithMulticore(multicore), gnet.WithTicker(ticker), gnet.WithCodec(gnet.NewLengthFieldBasedFrameCodec(encoderConfig, decoderConfig))))
}
