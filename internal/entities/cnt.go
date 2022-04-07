package entities

import "sync/atomic"

type CntAtomic32 uint32

func (c *CntAtomic32) Inc(delta uint32) uint32 {
	return atomic.AddUint32((*uint32)(c), delta)
}

func (c *CntAtomic32) Total() uint32 {
	return atomic.LoadUint32((*uint32)(c))
}
