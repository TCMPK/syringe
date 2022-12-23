package main

import (
	"container/heap"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	queueCandidateTimes = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "syringe",
			Name:      "queue_candidate_timer",
			Help:      "queue_candidate_timer",
			Buckets:   []float64{1, 5, 10, 30, 60, 300, 600, 1800, 3600},
		})
)

func init() {
	prometheus.Register(queueCandidateTimes)
}

type DomainHeap []*Domain

// Heap Impl
func (pq DomainHeap) Len() int { return len(pq) }

func (pq DomainHeap) Less(i, j int) bool {
	// We want Pop to give us the lowest based on expiration number as the priority
	// The lower the expiry, the higher the priority
	return pq[i].Refresh_at < pq[j].Refresh_at
}

// We just implement the pre-defined function in interface of heap.
func (pq *DomainHeap) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func (pq *DomainHeap) PopDomain() Domain {
	var domain *Domain = heap.Pop(pq).(*Domain)
	return *domain
}

func (pq *DomainHeap) PushDomain(d *Domain) {
	heap.Push(pq, d)
	pq.update(d)
}

func (pq *DomainHeap) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Domain)
	item.index = n
	*pq = append(*pq, item)
}

func (pq DomainHeap) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *DomainHeap) update(item *Domain) {
	heap.Fix(pq, item.index)
}

func (h DomainHeap) Dump() string {
	output := ""
	var i int = 0
	output += fmt.Sprintf("----------- Dump DomainHeap (size=%d) -----------\n", h.Len())
	for _, d := range h {
		output += fmt.Sprintf("[%d]: %s[%d] due in %d seconds \n", i, d.ToString(), d.index, d.SecondsUntilDue())
		i++
	}
	output += fmt.Sprintf("----------- END Dump DomainHeap (size=%d) -----------\n", h.Len())
	return output
}

type heapPopChanMsg struct {
	h      *DomainHeap
	result chan Domain
}

// heapPushChanMsg - the message structure for a push chan
type heapPushChanMsg struct {
	h *DomainHeap
	x Domain
}

func (dh *DomainHeap) AddDomain(d Domain) {
	domainsAdded.Inc()
	HeapPush(dh, d)
}

func HeapPush(h *DomainHeap, x Domain) {
	heapPushChan <- heapPushChanMsg{
		h: h,
		x: x,
	}
}

// HeapPop - safely pop item from a heap interface
func HeapPop(h *DomainHeap) interface{} {
	var result = make(chan Domain)
	heapPopChan <- heapPopChanMsg{
		h:      h,
		result: result,
	}
	return <-result
}

// stopWatchHeapOps - stop watching for heap operations
func (dh *DomainHeap) watchHeapOps() {
	go func() {
		for {
			select {
			case popMsg := <-heapPopChan:
				d := dh.PopDomain()
				log.Debug("heap pop ", d.ToString())
				popMsg.result <- d
			case pushMsg := <-heapPushChan:
				d := &(pushMsg.x)
				log.Debug("heap push ", d.ToString())
				dh.PushDomain(d)
				queueCandidateTimes.Observe(float64(d.SecondsUntilDue()))
			}
		}
	}()
}
