package pool

import (
	"net"
)

type connPool struct {
	Conn   net.Conn
	Use    bool
}

type ConnPool struct {
	pools []*connPool
}

func InitPool(num int) *ConnPool {
	return &ConnPool{
		pools: make([]*connPool, num),
	}
}

func (this *ConnPool) AddConn(conn net.Conn) {
	for i, c := range this.pools {
		if c == nil {
			this.pools[i] = &connPool{
				Conn:   conn,
				Use:    false,
			}
			break
		}
	}
}

func (this *ConnPool) Cap() int {
	i := 0

	for _, c := range this.pools {
		if c != nil {
			i++
		}
	}

	return i
}

func (this *ConnPool) NextConn() (int, *net.Conn) {
	for i, c := range this.pools {
		if c != nil && !c.Use {
			c.Use = true
			return i, &c.Conn
		}
	}

	return 0, nil
}

func (this *ConnPool) Destory(i int) {
	this.pools[i] = nil
}
