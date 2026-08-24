// Package diagnosis 负责代谢终点推断：拐点识别、多通道综合与采样比对。
package diagnosis

import (
	"math"
	"sort"

	"task217-fermend/internal/model"
)

// Knee 表示一个代谢拐点候选。
type Knee struct {
	Source  string  // knee_do / knee_ph / knee_feed
	TUnix   int64
	Value   float64
	Channel int64
}

// DoReboundTime 溶氧回升点：溶氧谷（最小值）对应时刻。
func DoReboundTime(points []*model.SamplePoint) (int64, float64, bool) {
	if len(points) == 0 {
		return 0, 0, false
	}
	best := points[0]
	for _, p := range points[1:] {
		if p.Value < best.Value {
			best = p
		}
	}
	return best.TUnix, best.Value, true
}

// PhKneeTime pH 拐点：一阶差分绝对斜率最大点。
func PhKneeTime(points []*model.SamplePoint) (int64, float64, bool) {
	if len(points) < 2 {
		return 0, 0, false
	}
	bestT := points[1].TUnix
	bestV := points[1].Value
	bestSlope := math.Abs(points[1].Value - points[0].Value)
	for i := 2; i < len(points); i++ {
		slope := math.Abs(points[i].Value - points[i-1].Value)
		if slope > bestSlope {
			bestSlope = slope
			bestT = points[i].TUnix
			bestV = points[i].Value
		}
	}
	return bestT, bestV, true
}

// FeedKneeTime 补料拐点：累积补料曲线的最大二阶变化（速率转折）。
func FeedKneeTime(points []*model.SamplePoint) (int64, float64, bool) {
	if len(points) < 3 {
		return 0, 0, false
	}
	bestT := points[1].TUnix
	bestV := points[1].Value
	bestCurv := 0.0
	for i := 2; i < len(points); i++ {
		v1 := points[i-1].Value - points[i-2].Value
		v2 := points[i].Value - points[i-1].Value
		curv := math.Abs(v2 - v1)
		if curv > bestCurv {
			bestCurv = curv
			bestT = points[i].TUnix
			bestV = points[i].Value
		}
	}
	return bestT, bestV, true
}

// MedianTime 返回时间序列的中位数（稳健综合）。
func MedianTime(ts []int64) int64 {
	if len(ts) == 0 {
		return 0
	}
	sorted := append([]int64(nil), ts...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	return sorted[len(sorted)/2]
}
