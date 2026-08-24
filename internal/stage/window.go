package stage

import "task217-fermend/internal/model"

// Window 表示一个半开时间窗 [Start, End)，End 为 0 表示开放到尾。
type Window struct {
	Name  string
	Start int64
	End   int64
}

// BuildWindows 把阶段列表归一化为时间窗切片，并按序号排序（调用方已按 seq 排序）。
func BuildWindows(stages []*model.Stage) []Window {
	out := make([]Window, 0, len(stages))
	for _, s := range stages {
		out = append(out, Window{Name: s.Name, Start: s.StartUnix, End: s.EndUnix})
	}
	return out
}

// Locate 返回时刻 t 所在的阶段名；不在任何阶段返回空串。
func Locate(windows []Window, t int64) string {
	for _, w := range windows {
		if t < w.Start {
			continue
		}
		if w.End == 0 || t <= w.End {
			return w.Name
		}
	}
	return ""
}

// Covering 判断窗口是否覆盖时刻 t。
func (w Window) Covering(t int64) bool {
	if t < w.Start {
		return false
	}
	return w.End == 0 || t < w.End
}
