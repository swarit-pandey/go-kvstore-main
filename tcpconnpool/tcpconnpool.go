package tcpconnpool

import (
	"errors"
	"net"
	"sync"
)

type ConnPool struct {
	mu        sync.Mutex
	conns     chan net.Conn
	addr      string
	maxConns  int
	connected int
}

func NewConnPool(addr string, maxConns int) *ConnPool {
	return &ConnPool{
		conns:    make(chan net.Conn, maxConns),
		addr:     addr,
		maxConns: maxConns,
	}
}

func (p *ConnPool) Get() (net.Conn, error) {
	select {
	case conn := <-p.conns:
		return conn, nil
	default:
		p.mu.Lock()
		defer p.mu.Unlock()

		if p.connected > p.maxConns {
			return nil, errors.New("max connection reached")
		}

		conn, err := net.Dial("tcp", p.addr)
		if err != nil {
			return nil, err
		}

		p.connected++
		return conn, nil
	}
}

func (p *ConnPool) Put(conn net.Conn) {
	select {
	case p.conns <- conn:
	default:
		p.mu.Lock()
		p.connected--
		p.mu.Unlock()
		conn.Close()
	}
}

func (p *ConnPool) Close() {
	for {
		select {
		case conn := <-p.conns:
			conn.Close()
		default:
			return
		}
	}
}
