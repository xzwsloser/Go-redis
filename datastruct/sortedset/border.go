package sortedset

/**
@Author: loser
@Description: 描述边界的接口
*/

const (
	ScoreInfLow  int8 = -1
	ScoreInfHigh int8 = 1
	MemInfLow    int8 = '-'
	MemInfHigh   int8 = '+'
)

type Border interface {
	less(element *Element) bool
	greater(element *Element) bool
	getValue() any
	getExclude() bool              // 闭区间 or 开区间
	IsIntersected(max Border) bool // 注意到如果参数是接口类型,为了获取到属性,
	// 那么可以在接口中在提供一个 get 方法即可
}

// SocreBorder 使用分数表示的边界信息  可以表示普通的浮点数 -inf 和 +inf
type ScoreBorder struct {
	Inf     int8
	Value   float64
	Exclude bool // true 表示开区间,否则表示闭区间
}

func (s *ScoreBorder) less(element *Element) bool {
	if s.Inf == ScoreInfLow {
		return true
	} else if s.Inf == ScoreInfHigh {
		return false
	}

	if s.getExclude() {
		return s.Value < element.Score
	}
	return s.Value <= element.Score
}

func (s *ScoreBorder) greater(element *Element) bool {
	if s.Inf == ScoreInfHigh {
		return true
	} else if s.Inf == ScoreInfLow {
		return false
	}

	if s.getExclude() {
		return s.Value > element.Score
	}
	return s.Value >= element.Score
}

func (s *ScoreBorder) getValue() any {
	if s.Inf == ScoreInfLow {
		return "-inf"
	} else if s.Inf == ScoreInfHigh {
		return "+inf"
	}
	return s.Value
}

func (s *ScoreBorder) getExclude() bool {
	return s.Exclude
}

func (s *ScoreBorder) IsIntersected(max Border) bool {
	minValue := s.Value
	maxValue := max.(*ScoreBorder).Value
	return minValue < maxValue || (minValue == maxValue && !max.getExclude() && !s.getExclude())
}

type MemberBorder struct {
	Inf     int8
	Value   string
	Exclude bool
}

func (m *MemberBorder) less(element *Element) bool {
	if m.Inf == MemInfLow {
		return true
	} else if m.Inf == MemInfHigh {
		return false
	}

	if m.Exclude {
		return m.Value < element.Member
	}
	return m.Value <= element.Member
}

func (m *MemberBorder) greater(element *Element) bool {
	if m.Inf == MemInfHigh {
		return true
	} else if m.Inf == MemInfLow {
		return false
	}

	// 表示不包含边界值
	if m.getExclude() {
		return m.Value > element.Member
	}
	return m.Value >= element.Member
}

func (m *MemberBorder) getValue() any {
	if m.Inf == MemInfLow {
		return "-inf"
	} else if m.Inf == MemInfHigh {
		return "+inf"
	}
	return m.Value
}

func (m *MemberBorder) getExclude() bool {
	return m.Exclude
}

func (m *MemberBorder) IsIntersected(max Border) bool {
	minValue := m.Value
	maxValue := max.(*MemberBorder).Value
	return minValue < maxValue || (minValue == maxValue && !max.getExclude() && !m.getExclude())
}
