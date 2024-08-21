package valve

import (
	"fmt"
	"io"
	"sync/atomic"

	"github.com/ardnew/valve/internal"
)

// Limit restricts the total bytes read and written,
// through the underlying [io.Reader] and [io.Writer] interfaces,
// by governing I/O requests forwarded to an embedded [Meter].
type Limit struct {
	*Meter
	rMax atomic.Int64
	wMax atomic.Int64
}

const Unlimited = -1

// NewLimit returns a new [Limit]
// that restricts the total bytes read from r and written to w
// to a maximum of rMax and wMax bytes, respectively.
func NewLimit(r io.Reader, rMax int64, w io.Writer, wMax int64) *Limit {
	l := &Limit{Meter: NewMeter(r, w)}
	l.SetMaxCount(rMax, wMax)
	return l
}

// NewReadLimit returns a new [Limit]
// that restricts the total bytes read from r to a maximum of rMax bytes.
func NewReadLimit(r io.Reader, rMax int64) *Limit {
	l := &Limit{Meter: NewReadMeter(r)}
	l.SetMaxCountRead(rMax)
	return l
}

// NewWriteLimit returns a new [Limit]
// that restricts the total bytes written to w to a maximum of wMax bytes.
func NewWriteLimit(w io.Writer, wMax int64) *Limit {
	l := &Limit{Meter: NewWriteMeter(w)}
	l.SetMaxCountWrite(wMax)
	return l
}

// NewReadWriteLimit returns a new [Limit]
// that restricts the total bytes read from and written to rw
// to a maximum of rMax and wMax bytes, respectively.
func NewReadWriteLimit(rw io.ReadWriter, rMax, wMax int64) *Limit {
	l := &Limit{Meter: NewReadWriteMeter(rw)}
	l.SetMaxCount(rMax, wMax)
	return l
}

// CanRead returns true if the Limit is capable of reading bytes.
func (l *Limit) CanRead() bool {
	return l.Meter != nil && l.Meter.CanRead()
}

// CanWrite returns true if the Limit is capable of writing bytes.
func (l *Limit) CanWrite() bool {
	return l.Meter != nil && l.Meter.CanWrite()
}

// Read reads bytes from the underlying [io.Reader] to p
// and increments the total bytes read by n
// until the total bytes read reaches the maximum limit.
//
// See [Meter] for additional details.
func (l *Limit) Read(p []byte) (n int, err error) { //nolint: varnamelen
	if !l.CanRead() {
		return 0, io.ErrClosedPipe
	}
	var e error //nolint: varnamelen
	switch req, rem := int64(len(p)), l.RemainingCountRead(); {
	case l.MaxCountRead() == Unlimited:
		return l.Meter.Read(p)
	case l.CountRead() >= l.MaxCountRead():
		return 0, l.MakeReadLimitError(req, 0)
	case req > rem:
		p, e = p[:rem], l.MakeReadLimitError(req, rem)
	}
	if n, err = l.Reader.Read(p); err == nil {
		err = e
	}
	_ = l.AddCountRead(int64(n))
	return
}

// ReadFrom copies bytes from r to the underlying [io.Writer]
// and increments the total bytes written by n
// until the total bytes written reaches the maximum limit.
//
// See [Meter] for additional details.
func (l *Limit) ReadFrom(r io.Reader) (n int64, err error) { //nolint: varnamelen
	if !l.CanWrite() {
		return 0, io.ErrClosedPipe
	}
	switch rem := l.RemainingCountWrite(); {
	case l.MaxCountWrite() == Unlimited:
		return l.Meter.ReadFrom(r)
	case rem <= 0:
		return 0, l.MakeWriteLimitError(rem, 0)
	default:
		n, err = io.CopyN(l.Writer, r, rem)
		// if err != nil && n == rem {
		// 	err = nil
		// }
		_ = l.AddCountWrite(n)
		return
	}
}

// Write writes bytes from p to the underlying [io.Writer]
// and increments the total bytes written by n
// until the total bytes written reaches the maximum limit.
//
// See [Meter] for additional details.
func (l *Limit) Write(p []byte) (n int, err error) { //nolint: varnamelen
	if !l.CanWrite() {
		return 0, io.ErrClosedPipe
	}
	var e error //nolint: varnamelen
	switch req, rem := int64(len(p)), l.RemainingCountWrite(); {
	case l.MaxCountWrite() == Unlimited:
		return l.Meter.Write(p)
	case l.CountWrite() >= l.MaxCountWrite():
		return 0, l.MakeWriteLimitError(req, 0)
	case req > rem:
		p, e = p[:rem], l.MakeWriteLimitError(req, rem)
	}
	if n, err = l.Writer.Write(p); err == nil {
		err = e
	}
	_ = l.AddCountWrite(int64(n))
	return
}

// WriteTo writes bytes from the underlying [io.Reader]
// to w and increments the total bytes read by n
// until the total bytes read reaches the maximum limit.
//
// See [Meter] for additional details.
func (l *Limit) WriteTo(w io.Writer) (n int64, err error) { //nolint: varnamelen
	if !l.CanRead() {
		return 0, io.ErrClosedPipe
	}
	switch rem := l.RemainingCountRead(); {
	case l.MaxCountRead() == Unlimited:
		return l.Meter.WriteTo(w)
	case rem <= 0:
		return 0, l.MakeReadLimitError(rem, 0)
	default:
		n, err = io.CopyN(w, l.Reader, rem)
		// if err != nil && n == rem {
		// 	err = nil
		// }
		_ = l.AddCountRead(n)
		return
	}
}

// Close closes the embedded [Meter].
func (l *Limit) Close() error {
	if l.Meter != nil {
		return l.Meter.Close()
	}
	return nil
}

// MaxCount returns the maximum bytes that may be read and written.
func (l *Limit) MaxCount() (r, w int64) {
	return l.rMax.Load(), l.wMax.Load()
}

// MaxCountRead returns the maximum bytes that may be read.
func (l *Limit) MaxCountRead() int64 {
	return l.rMax.Load()
}

// MaxCountWrite returns the maximum bytes that may be written.
func (l *Limit) MaxCountWrite() int64 {
	return l.wMax.Load()
}

// RemainingCount returns the total bytes that may be read and written
// before exceeding their respective limits.
func (l *Limit) RemainingCount() (r, w int64) {
	return l.MaxCountRead() - l.CountRead(), l.MaxCountWrite() - l.CountWrite()
}

// RemainingCountRead returns the total bytes that may be read
// before exceeding the read limit.
func (l *Limit) RemainingCountRead() int64 {
	return l.MaxCountRead() - l.CountRead()
}

// RemainingCountWrite returns the total bytes that may be written
// before exceeding the write limit.
func (l *Limit) RemainingCountWrite() int64 {
	return l.MaxCountWrite() - l.CountWrite()
}

// SetMaxCount restricts the total bytes read and written
// to a maximum of r and w bytes, respectively.
func (l *Limit) SetMaxCount(r, w int64) {
	l.rMax.Store(r)
	l.wMax.Store(w)
}

// SetMaxCountRead restricts the total bytes read to a maximum of r bytes.
func (l *Limit) SetMaxCountRead(r int64) {
	l.rMax.Store(r)
}

// SetMaxCountWrite restricts the total bytes written to a maximum of w bytes.
func (l *Limit) SetMaxCountWrite(w int64) {
	l.wMax.Store(w)
}

// MakeReadLimitError returns a [LimitError] describing a short read of n bytes
// after attempting to read req bytes.
func (l *Limit) MakeReadLimitError(req, n int64) error {
	return internal.MakeError(LimitError{Limit: l, op: Read, Requested: req, Accepted: n})
}

// MakeWriteLimitError returns a [LimitError] describing a short write of n
// bytes after attempting to write req bytes.
func (l *Limit) MakeWriteLimitError(req, n int64) error {
	return internal.MakeError(LimitError{Limit: l, op: Write, Requested: req, Accepted: n})
}

// LimitError is returned when a short read/write occurs due to a byte limit.
type LimitError struct {
	// Limit is the object that imposed the I/O limit.
	*Limit
	// op is a bitmask identifying the requested I/O operation.
	op IO
	// Requested is the number of bytes requested for read/write.
	Requested int64
	// Accepted is the number of bytes successfully read/written.
	Accepted int64
}

// String returns a string representation of the [LimitError].
func (e LimitError) Error() string {
	var eMax int64
	switch {
	case e.op&Read != 0:
		eMax = e.MaxCountRead()
	case e.op&Write != 0:
		eMax = e.MaxCountWrite()
	default:
		return internal.MakeInvalidOperationError().Error()
	}
	return fmt.Sprintf(
		"short %s: %d of %d bytes (cumulative %s limit = %d bytes)",
		e.op, e.Accepted, e.Requested, e.op, eMax,
	)
}
