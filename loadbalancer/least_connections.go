package loadbalancer

import "container/heap"

type targetInfo struct {
	target        string
	index         int
	count         int
	weight        int
	insertionTime uint64
}

type targetsHeap struct {
	time     uint64
	heap     []*targetInfo
	weighted bool
}

func (th targetsHeap) Less(i, j int) bool {
	if th.heap[i].count == th.heap[j].count {
		if th.weighted {
			if th.heap[i].weight == th.heap[j].weight {
				return th.heap[i].insertionTime < th.heap[j].insertionTime
			}

			return th.heap[i].weight > th.heap[j].weight
		}

		return th.heap[i].insertionTime < th.heap[j].insertionTime
	}

	return th.heap[i].count < th.heap[j].count
}
func (th targetsHeap) Len() int { return len(th.heap) }

func (th targetsHeap) Swap(i, j int) {
	th.heap[i], th.heap[j] = th.heap[j], th.heap[i]
	th.heap[i].index, th.heap[j].index = th.heap[j].index, th.heap[i].index
}

func (th *targetsHeap) Push(x any) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	th.time++
	item := x.(*targetInfo)
	item.index = len(th.heap)
	item.insertionTime = th.time

	th.heap = append(th.heap, item)
}

func (th *targetsHeap) Pop() any {
	old := th.heap
	n := len(old)
	x := old[n-1]
	th.heap = old[0 : n-1]
	return x
}

type LeastConnPolicy struct {
	basePolicy
	minHeap    heap.Interface
	targetsMap map[string]*targetInfo
}

func NewLeastConnLoadBalancer(targets []string, weights []int) *LeastConnPolicy {
	tm := make(map[string]*targetInfo, len(targets))
	h := targetsHeap{
		time: 0,
		heap: make([]*targetInfo, len(targets)),
	}

	for i, t := range targets {
		ti := targetInfo{
			target:        t,
			count:         0,
			index:         i,
			insertionTime: uint64(i),
		}

		if weights != nil {
			ti.weight = weights[i]
		}

		tm[t] = &ti
		h.heap[i] = &ti
		h.weighted = weights != nil
	}

	currTime := h.heap[len(h.heap)-1].insertionTime
	h.time = currTime

	heap.Init(&h)

	lb := LeastConnPolicy{
		minHeap:    &h,
		targetsMap: tm,
	}

	return &lb
}

func (p *LeastConnPolicy) ChooseTarget(_ []string) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	min := heap.Pop(p.minHeap).(*targetInfo)
	target := min.target
	min.count++

	heap.Push(p.minHeap, min)

	return target
}

func (p *LeastConnPolicy) Done(target string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ti := p.targetsMap[target]
	ti.count--

	heap.Fix(p.minHeap, ti.index)
}
