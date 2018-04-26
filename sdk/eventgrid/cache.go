package eventgrid

import (
	"sync"
	"time"
)

// CacheDefaultMaxDepth is the maximum number of Events that will
// be stored here, before they begin automatically removed.
const CacheDefaultMaxDepth uint = 100000

// CacheDefaultTTL is the default length of time that each event will live
// in the cache before it is aut
const CacheDefaultTTL = time.Hour * 48

// Cache will hold a set number of events for a short amount of time.
type Cache struct {
	sync.RWMutex
	maxDepth uint
	ttl      time.Duration
	root     *cacheNode
}

// MaxDepth gets the largest number of `Event` instances that this `Cache`
// will hold before automatically deleting the least recently arriving ones.
func (c *Cache) MaxDepth() uint {
	c.RLock()
	defer c.RUnlock()

	return c._MaxDepth()
}

func (c *Cache) _MaxDepth() uint {
	if c.maxDepth == 0 {
		return CacheDefaultMaxDepth
	}
	return c.maxDepth
}

// SetMaxDepth changes the largest number of `Event` instances that this `Cache`.
// will hold.
func (c *Cache) SetMaxDepth(depth uint) {
	c.Lock()
	defer c.Unlock()

	c.maxDepth = depth
}

// TTL get the amount of time each event will last before being cleared from the `Cache`.
func (c *Cache) TTL() time.Duration {
	c.RLock()
	defer c.RUnlock()

	return c._TTL()
}

func (c *Cache) _TTL() time.Duration {
	if c.ttl <= 0 {
		return CacheDefaultTTL
	}
	return c.ttl
}

// SetTTL sets the amount of time each event will last before being cleared from the `Cache`.
func (c *Cache) SetTTL(d time.Duration) {
	c.Lock()
	defer c.Unlock()

	c.ttl = d
}

// Add creates an entry in the `Cache`.
func (c *Cache) Add(e Event) {
	c.Lock()
	defer c.Unlock()

	created := &cacheNode{
		Event:      e,
		next:       c.root,
		expiration: time.Now().Add(c._TTL()),
	}

	c.root = created
}

// Clear removes all entries from the Event Cache.
func (c *Cache) Clear() {
	c.Lock()
	defer c.Unlock()

	c.root = nil
}

// List reads all of the Events in the cache at a particular moment.
func (c *Cache) List() (results []Event) {
	c.RLock()
	defer c.RUnlock()

	called := time.Now()

	prev := c.root
	i := uint(0)

	for current := c.root; current != nil; current = current.next {
		if i >= c._MaxDepth() {
			current.next = nil
			break
		}

		if current.expiration.After(called) {
			results = append(results, current.Event)
			i++
		} else {
			prev.next = current.next
		}
	}
	return
}

type cacheNode struct {
	Event
	next       *cacheNode
	expiration time.Time
}
