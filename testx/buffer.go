package testx

import (
	"bytes"
	"sync"
	"testing"
)

type ConcurrentBuffer struct {
	b *bytes.Buffer
	m sync.RWMutex
	t *testing.T
}

func NewConcurrentBuffer(t *testing.T) *ConcurrentBuffer {
	return &ConcurrentBuffer{
		b: new(bytes.Buffer),
		t: t,
	}
}

func (c *ConcurrentBuffer) Write(p []byte) (n int, err error) {
	c.m.Lock()
	defer c.m.Unlock()
	return c.b.Write(p)
}

func (c *ConcurrentBuffer) String() string {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.b.String()
}
