package couchbase

import (
	"errors"
	"time"

	"github.com/jkassemi/gomemcached/client"
)

type connectionPool struct {
	host, name  string
	mkConn      func(host, name string) (*memcached.Client, error)
	connections chan *memcached.Client
}

func newConnectionPool(host, name string, poolSize int) *connectionPool {
	return &connectionPool{
		host:        host,
		name:        name,
		connections: make(chan *memcached.Client, poolSize),
		mkConn:      defaultMkConn,
	}
}

func defaultMkConn(host, name string) (*memcached.Client, error) {
	conn, err := memcached.Connect("tcp", host)
	if err != nil {
		return nil, err
	}
	if name != "default" {
		conn.Auth(name, "") // error checking?
	}
	return conn, nil
}

func (cp *connectionPool) Close() (err error) {
	defer func() { err, _ = recover().(error) }()
	close(cp.connections)
	for c := range cp.connections {
		c.Close()
	}
	return
}

func (cp *connectionPool) Get() (*memcached.Client, error) {
	if cp == nil {
		return nil, errors.New("no pool")
	}

	select {
	case rv := <-cp.connections:
		return rv, nil
	case <-time.After(time.Millisecond):
		// Build a connection if we can't get a real one.
		// This can potentially be an overflow connection, or
		// a pooled connection.
		return cp.mkConn(cp.host, cp.name)
	}
}

func (cp *connectionPool) Return(c *memcached.Client) {
	if cp == nil {
		return
	}

	if c != nil {
		if c.IsHealthy() {
			select {
			case cp.connections <- c:
			default:
				// Overflow connection.
				c.Close()
			}
		} else {
			c.Close()
		}
	}
}

func (cp *connectionPool) StartTapFeed(args *memcached.TapArguments) (*memcached.TapFeed, error) {
	if cp == nil {
		return nil, errors.New("no pool")
	}
	mc, err := cp.Get()
	if err != nil {
		return nil, err
	}
	return mc.StartTapFeed(*args)
}
