package reachability

import (
	"github.com/sparrc/go-ping"
	"sync"
	"sync/atomic"
	"time"
)

type Checker func() (averageRTT time.Duration, packetLossRate float64)

func NewChecker(count int, interval time.Duration, timeout time.Duration, hosts ...string) Checker {
	var wg sync.WaitGroup
	return func() (time.Duration, float64) {
		successes := int64(0)
		totalRTT := int64(0)
		wg.Add(len(hosts))
		for _, h := range hosts {
			host := h
			go func() {
				defer wg.Done()
				pinger, err := ping.NewPinger(host)
				if err != nil {
					return
				}
				pinger.Count = count
				pinger.Interval = interval
				pinger.Timeout = timeout
				pinger.Run()
				stats := pinger.Statistics()
				atomic.AddInt64(&successes, int64(stats.PacketsRecv))
				atomic.AddInt64(&totalRTT, int64(stats.AvgRtt)*int64(stats.PacketsRecv))
			}()
		}
		wg.Wait()
		s := atomic.LoadInt64(&successes)
		trtt := atomic.LoadInt64(&totalRTT)
		return time.Duration(trtt / s), 1 - float64(s)/float64(len(hosts)*count)
	}
}
