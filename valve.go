package valve

// IO is a bitmask identifying types of I/O operations.
type IO int

const (
	Read IO = 1 << iota
	Write
	Close

	// Commonly used combinations.
	ReadWrite = Read | Write

	// Sentinel values.
	NOP      IO = 0
	DEADBEEF IO = ^NOP
)

func (o IO) String() string {
	switch o {
	case Read:
		return "read"
	case Write:
		return "write"
	case Close:
		return "close"
	case ReadWrite:
		return "read/write"
	case NOP:
		return "nop"
	case DEADBEEF:
		return "invalid"
	default:
		return "unknown"
	}
}
