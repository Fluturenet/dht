// Pings DHT nodes with the given network addresses.
package main

import (
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"time"

	"github.com/anacrolix/dht"
	"github.com/anacrolix/dht/krpc"
	"github.com/anacrolix/tagflag"
	"github.com/bradfitz/iter"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var args = struct {
		Timeout time.Duration
		Nodes   []string `type:"pos" arity:"+" help:"nodes to ping e.g. router.bittorrent.com:6881"`
	}{
		Timeout: math.MaxInt64,
	}
	tagflag.Parse(&args)
	s, err := dht.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("dht server on %s", s.Addr())
	timeout := time.After(args.Timeout)
	pongChan := make(chan pong)
	startPings(s, pongChan, args.Nodes)
	numResp := receivePongs(pongChan, timeout, len(args.Nodes))
	fmt.Printf("%d/%d responses (%f%%)\n", numResp, len(args.Nodes), 100*float64(numResp)/float64(len(args.Nodes)))
}

func receivePongs(pongChan chan pong, timeout <-chan time.Time, maxPongs int) (numResp int) {
	for range iter.N(maxPongs) {
		select {
		case pong := <-pongChan:
			if pong.err != nil {
				fmt.Fprintf(os.Stderr, "pong error: %s", pong.err)
				break
			}
			numResp++
			fmt.Printf("%-65s %s\n", fmt.Sprintf("%x (%s):", pong.krpc.SenderID(), pong.addr), pong.rtt)
		case <-timeout:
			fmt.Fprintf(os.Stderr, "timed out\n")
			return
		}
	}
	return
}

func startPings(s *dht.Server, pongChan chan pong, nodes []string) {
	for i, addr := range nodes {
		if i != 0 {
			// Put a small sleep between pings to avoid network issues.
			time.Sleep(1 * time.Millisecond)
		}
		ping(addr, pongChan, s)
	}
}

type pong struct {
	addr string
	krpc krpc.Msg
	err  error
	rtt  time.Duration
}

func ping(netloc string, pongChan chan pong, s *dht.Server) {
	addr, err := net.ResolveUDPAddr("udp4", netloc)
	if err != nil {
		log.Fatal(err)
	}
	start := time.Now()
	if err := s.Ping(addr, func(resp krpc.Msg, err error) {
		pongChan <- pong{
			addr: netloc,
			krpc: resp,
			rtt:  time.Now().Sub(start),
			err:  err,
		}
	}); err != nil {
		log.Fatal(err)
	}
}