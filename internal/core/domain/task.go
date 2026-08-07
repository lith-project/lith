package domain

// TaskState is the sealed typed view of a raw task marker.
type TaskState interface{ taskState() }
type TaskOpen struct{}

func (TaskOpen) taskState() {}

type TaskDone struct{}

func (TaskDone) taskState() {}

type TaskOther struct{ marker rune }

func (TaskOther) taskState()     {}
func (s TaskOther) Marker() rune { return s.marker }

// Task preserves the raw marker and typed state.
type Task struct {
	marker   rune
	state    TaskState
	rangeVal Range
}

func NewTask(marker rune, rangeVal Range) Task {
	var state TaskState
	switch marker {
	case ' ':
		state = TaskOpen{}
	case 'x', 'X':
		state = TaskDone{}
	default:
		state = TaskOther{marker: marker}
	}
	return Task{marker: marker, state: state, rangeVal: rangeVal}
}
func (t Task) Marker() rune     { return t.marker }
func (t Task) State() TaskState { return t.state }
func (t Task) Range() Range     { return t.rangeVal }
