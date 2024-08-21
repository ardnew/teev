package valve

import (
	"errors"
	"io"
	"sync/atomic"
)

// Meter records the total bytes read and written,
// through the underlying [io.Reader] and [io.Writer] given at construction,
// using any of the following interfaces:
//
//   - [io.Reader] (read)
//   - [io.ReaderFrom] (write)
//   - [io.Writer] (write)
//   - [io.WriterTo] (read)
//
// Constructors also exist for read-only, write-only, and read-write Meters.
// Methods without an underlying interface return [io.ErrClosedPipe].
//
// Meter also implements the [io.Closer] interface.
// Closing a Meter closes each underlying interface that implements [io.Closer].
type Meter struct {
	io.Reader
	io.Writer
	rCount atomic.Int64
	wCount atomic.Int64
}

// NewMeter returns a new [Meter]
// that counts the total bytes read from r and written to w.
func NewMeter(r io.Reader, w io.Writer) *Meter {
	return &Meter{Reader: r, Writer: w}
}

// NewReadMeter returns a new [Meter]
// that counts the total bytes read from r.
func NewReadMeter(r io.Reader) *Meter {
	return &Meter{Reader: r}
}

// NewWriteMeter returns a new [Meter]
// that counts the total bytes written to w.
func NewWriteMeter(w io.Writer) *Meter {
	return &Meter{Writer: w}
}

// NewReadWriteMeter returns a new [Meter]
// that counts the total bytes read from and written to rw.
func NewReadWriteMeter(rw io.ReadWriter) *Meter {
	return &Meter{Reader: rw, Writer: rw}
}

// CanRead returns true if the Meter is capable of reading bytes.
func (m *Meter) CanRead() bool {
	return m.Reader != nil
}

// CanWrite returns true if the Meter is capable of writing bytes.
func (m *Meter) CanWrite() bool {
	return m.Writer != nil
}

// Read reads bytes from the underlying [io.Reader] to p
// and increments the total bytes read by n.
//
// See [io.Reader] for details.
func (m *Meter) Read(p []byte) (n int, err error) {
	if !m.CanRead() {
		return 0, io.ErrClosedPipe
	}
	n, err = m.Reader.Read(p)
	_ = m.AddCountRead(int64(n))
	return
}

// ReadFrom copies bytes from r to the underlying [io.Writer]
// and increments the total bytes written by n.
//
// See [io.ReaderFrom] for details.
func (m *Meter) ReadFrom(r io.Reader) (n int64, err error) {
	if !m.CanWrite() {
		return 0, io.ErrClosedPipe
	}
	n, err = io.Copy(m.Writer, r)
	_ = m.AddCountWrite(n)
	return
}

// Write writes bytes from p to the underlying [io.Writer]
// and increments the total bytes written by n.
//
// See [io.Writer] for details.
func (m *Meter) Write(p []byte) (n int, err error) {
	if !m.CanWrite() {
		return 0, io.ErrClosedPipe
	}
	n, err = m.Writer.Write(p)
	_ = m.AddCountWrite(int64(n))
	return
}

// WriteTo copies bytes from the underlying [io.Reader] to w
// and increments the total bytes read by n.
//
// See [io.WriterTo] for details.
func (m *Meter) WriteTo(w io.Writer) (n int64, err error) {
	if !m.CanRead() {
		return 0, io.ErrClosedPipe
	}
	n, err = io.Copy(w, m.Reader)
	_ = m.AddCountRead(n)
	return
}

// Close closes each underlying interface that implements [io.Closer].
//
// See [io.Closer] for details.
func (m *Meter) Close() error {
	return m.close(m.Reader, m.Writer)
}

func (m *Meter) close(v ...interface{}) (err error) {
	for _, v := range v {
		if c, ok := v.(io.Closer); ok {
			err = errors.Join(err, c.Close())
		}
	}
	return
}

// Count returns the total bytes read and written.
func (m *Meter) Count() (r, w int64) {
	return m.CountRead(), m.CountWrite()
}

// CountRead returns the total bytes read.
func (m *Meter) CountRead() int64 {
	return m.rCount.Load()
}

// CountWrite returns the total bytes written.
func (m *Meter) CountWrite() int64 {
	return m.wCount.Load()
}

// AddCount increments the total bytes read by r and written by w
// and returns the new byte counts.
func (m *Meter) AddCount(r, w int64) (nr, nw int64) {
	return m.AddCountRead(r), m.AddCountWrite(w)
}

// AddCountRead increments the total bytes read by r
// and returns the new byte count.
func (m *Meter) AddCountRead(r int64) int64 {
	return m.rCount.Add(r)
}

// AddCountWrite increments the total bytes written by w
// and returns the new byte count.
func (m *Meter) AddCountWrite(w int64) int64 {
	return m.wCount.Add(w)
}

// SetCount sets the total bytes read to r and written to w.
func (m *Meter) SetCount(r, w int64) {
	m.SetCountRead(r)
	m.SetCountWrite(w)
}

// SetCountRead sets the total bytes read to r.
func (m *Meter) SetCountRead(r int64) {
	m.rCount.Store(r)
}

// SetCountWrite sets the total bytes written to w.
func (m *Meter) SetCountWrite(w int64) {
	m.wCount.Store(w)
}

// ResetCount sets the total bytes read and written to zero.
func (m *Meter) ResetCount() {
	m.ResetCountRead()
	m.ResetCountWrite()
}

// ResetCountRead sets the total bytes read to zero.
func (m *Meter) ResetCountRead() {
	m.SetCountRead(0)
}

// ResetCountWrite sets the total bytes written to zero.
func (m *Meter) ResetCountWrite() {
	m.SetCountWrite(0)
}
