package lib

import "container/heap"

type IHeapData interface {
	GetRank() int
}

type IMinHeap interface {
	Insert(IHeapData)
	Min() IHeapData
}

func NewMinHeap() IMinHeap {
	minHeap := make(MinHeap, 0)
	return &minHeap
}

type MinHeap []IHeapData

func (mh *MinHeap) Insert(d IHeapData) {
	heap.Push(mh, d)
}
func (mh *MinHeap) Min() IHeapData {
	min := heap.Pop(mh)
	heap.Push(mh, min)
	return min.(IHeapData)
}

func (h MinHeap) Len() int {
	return len(h)
}

func (h MinHeap) Less(i, j int) bool {
	return h[i].GetRank() < h[j].GetRank()
}

func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *MinHeap) Push(x any) {
	*h = append(*h, x.(IHeapData))
}

func (h *MinHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

type MaxHeap []IHeapData

type IMaxHeap interface {
	Insert(IHeapData)
	Max() IHeapData
}

func NewMaxHeap() IMaxHeap {
	max := make(MaxHeap, 0)
	return &max
}

func (mh *MaxHeap) Insert(d IHeapData) {
	heap.Push(mh, d)
}
func (mh *MaxHeap) Max() IHeapData {
	min := heap.Pop(mh)
	heap.Push(mh, min)
	return min.(IHeapData)
}

func (h MaxHeap) Len() int {
	return len(h)
}

func (h MaxHeap) Less(i, j int) bool {
	return h[i].GetRank() > h[j].GetRank()
}

func (h MaxHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *MaxHeap) Push(x any) {
	*h = append(*h, x.(IHeapData))
}

func (h *MaxHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}
