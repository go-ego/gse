package segment

// Segment type a word with weight.
type Segment struct {
	Text   string
	Weight float64
}

// GetText return the segment's text.
func (s Segment) GetText() string {
	return s.Text
}

// GetWeight return the segment's weight.
func (s Segment) GetWeight() float64 {
	return s.Weight
}

// Segments type a slice of Segment.
type Segments []Segment

func (ss Segments) Len() int {
	return len(ss)
}

func (ss Segments) Less(i, j int) bool {
	if ss[i].Weight == ss[j].Weight {
		return ss[i].Text < ss[j].Text
	}

	return ss[i].Weight < ss[j].Weight
}

func (ss Segments) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}
