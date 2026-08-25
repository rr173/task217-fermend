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

// Locate 返回时刻 t 所在的阶段名；边界时刻归入后一阶段（半开窗 [Start, End)）。
// 窗口按 seq 升序，故首个命中者即为答案：t 落在前一窗口的 End 上时，前一窗口已不覆盖，
// 循环前进到以该点为 Start 的后一窗口并命中之。
func Locate(windows []Window, t int64) string {
	for _, w := range windows {
		if t < w.Start {
			continue
		}
		if w.End == 0 || t < w.End {
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
