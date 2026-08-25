package calibration

import (
	"math"

	"task217-fermend/internal/model"
)

// FindMinTime 返回曲线值最小点对应的时刻（如溶氧谷 = 回升起点）。
func FindMinTime(points []*model.SamplePoint) (int64, float64, bool) {
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

// FindMaxSlopeTime 返回一阶差分绝对斜率最大点的时刻（如 pH 拐点）。
func FindMaxSlopeTime(points []*model.SamplePoint) (int64, bool) {
	if len(points) < 2 {
		return 0, false
	}
	bestT := points[1].TUnix
	bestSlope := math.Abs(points[1].Value - points[0].Value)
	for i := 2; i < len(points); i++ {
		slope := math.Abs(points[i].Value - points[i-1].Value)
		if slope > bestSlope {
			bestSlope = slope
			bestT = points[i].TUnix
		}
	}
	return bestT, true
}

// FindFeedKneeTime 返回累积补料曲线最大二阶变化（速率转折）点的时刻。
func FindFeedKneeTime(points []*model.SamplePoint) (int64, bool) {
	if len(points) < 3 {
		return 0, false
	}
	bestT := points[1].TUnix
	bestCurv := 0.0
	for i := 2; i < len(points); i++ {
		v1 := points[i-1].Value - points[i-2].Value
		v2 := points[i].Value - points[i-1].Value
		curv := math.Abs(v2 - v1)
		if curv > bestCurv {
			bestCurv = curv
			bestT = points[i].TUnix
		}
	}
	return bestT, true
}

// EventTime 按通道物理量类型选取事件检测器，返回该通道特征事件时刻。
// 溶氧取谷点（回升起点）、pH 取最大斜率（拐点）、补料取速率转折点，
// 与诊断阶段 InferEndpoints 的拐点识别口径一致，保证跨通道滞后估计可比。
func EventTime(kind model.ChannelKind, points []*model.SamplePoint) (int64, bool) {
	switch kind {
	case model.KindDissolvedOxygen:
		t, _, ok := FindMinTime(points)
		return t, ok
	case model.KindPH:
		return FindMaxSlopeTime(points)
	case model.KindFeed:
		return FindFeedKneeTime(points)
	default:
		// 温度/搅拌等非拐点通道无可用事件时刻。
		return 0, false
	}
}

// EstimateLag 用同一物理事件（各自拐点/谷点）的时间差估计滞后秒数。
// 返回 channel 相对 reference 的滞后（正值表示 channel 落后）。
func EstimateLag(refEventT, channelEventT int64) float64 {
	return float64(channelEventT - refEventT)
}

// GapDetect 检测相邻采样点间隔超过阈值的时间窗缺口。
func GapDetect(points []*model.SamplePoint, maxGapSecs int64) []model.SensorSegment {
	var gaps []model.SensorSegment
	for i := 1; i < len(points); i++ {
		dt := points[i].TUnix - points[i-1].TUnix
		if dt > maxGapSecs {
			gaps = append(gaps, model.SensorSegment{
				StartUnix: points[i-1].TUnix,
				EndUnix:   points[i].TUnix,
				Status:    model.SegmentGap,
			})
		}
	}
	return gaps
}
