package log

type Entry struct {
	Idx  int64
	Term uint64
	Cmd  any
}
