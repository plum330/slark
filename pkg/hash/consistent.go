package hash

import (
	utils "github.com/go-slark/slark/pkg"
	"github.com/spaolacci/murmur3"
	"sort"
	"strconv"
	"sync"
)

type Consistent struct {
	f     func([]byte) uint64
	vn    int                 // 虚拟节点数量
	vnl   []uint64            // sorted虚拟节点hash构成环上的节点
	ring  map[uint64][]string // 虚拟节点 - 真实节点(节点冲突)
	nodes map[string]struct{} // 真实节点
	l     sync.RWMutex
}

type Option func(*Consistent)

func Func(f func([]byte) uint64) Option {
	return func(c *Consistent) {
		c.f = f

	}
}

func VirtualNodes(vn int) Option {
	return func(c *Consistent) {
		c.vn = vn
	}
}

func New(opts ...Option) *Consistent {
	c := &Consistent{
		f:     murmur3.Sum64,
		vn:    32,
		vnl:   make([]uint64, 0),
		ring:  make(map[uint64][]string),
		nodes: make(map[string]struct{}),
		l:     sync.RWMutex{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Add 添加真实节点(如ip) & 虚拟节点
func (c *Consistent) Add(node string) {
	// 可重复添加
	c.Delete(node)
	c.l.Lock()
	defer c.l.Unlock()
	// 添加到真实节点
	c.nodes[node] = struct{}{}
	// 计算虚拟节点映射关系 & 排序
	for i := 0; i < c.vn; i++ {
		h := c.f([]byte(node + strconv.Itoa(i)))
		c.vnl = append(c.vnl, h)
		c.ring[h] = append(c.ring[h], node)
	}
	sort.Slice(c.vnl, func(i, j int) bool {
		return c.vnl[i] < c.vnl[j]
	})
}

// Delete 删除真实节点 & 虚拟节点
func (c *Consistent) Delete(node string) {
	c.l.Lock()
	defer c.l.Unlock()
	_, ok := c.nodes[node]
	if !ok {
		return
	}
	for i := 0; i < c.vn; i++ {
		// 真实节点追加i字节作为虚拟节点计算hash
		h := c.f([]byte(node + strconv.Itoa(i)))
		// 二分查找首个大于hash值的节点
		index := sort.Search(len(c.vnl), func(i int) bool {
			return c.vnl[i] >= h
		})
		// 删除虚拟节点
		if index <= len(c.vnl)-1 && c.vnl[index] == h {
			c.vnl = append(c.vnl[:index], c.vnl[index+1:]...)
		}
		// 删除虚拟节点映射关系
		nodes, o := c.ring[h]
		if !o {
			continue
		}
		c.ring[h] = utils.Delete(nodes, node)
		if len(c.ring[h]) == 0 {
			delete(c.ring, h)
		}
	}
	// 删除真实节点
	delete(c.nodes, node)
}

func (c *Consistent) Fetch(node string) string {
	c.l.RLock()
	defer c.l.RUnlock()
	if len(c.ring) == 0 {
		return ""
	}
	// 计算真实节点hash
	h := c.f([]byte(node))
	// 顺时针在环上查找第一个大于当前hash的虚拟节点
	size := len(c.vnl)
	index := sort.Search(size, func(i int) bool {
		return c.vnl[i] >= h
	}) % size
	nodes := c.ring[c.vnl[index]]
	size = len(nodes)
	if size == 0 {
		return ""
	} else if size == 1 {
		return nodes[0]
	} else {
		// 冲突:一个虚拟节点对应多个真实节点，再hash并对真实节点取模
		// 32 bit FNV_prime取值 = 2^24 + 2^8 + 0x93 = 16777619 / FNV保持较小冲突概率
		h = c.f([]byte(node + "-" + strconv.Itoa(16777619)))
		return nodes[h%uint64(len(nodes))]
	}
}
