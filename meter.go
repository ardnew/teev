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
// [Close] will close each underlying interface that implements [io.Closer].
type Meter struct {
	rCount atomic.Int64
	wCount atomic.Int64
	reader io.Reader
	writer io.Writer
}

// NewMeter returns a new [Meter]
// that counts the total bytes read from r and written to w.
func NewMeter(r io.Reader, w io.Writer) *Meter {
	return &Meter{reader: r, writer: w}
}

// NewReadMeter returns a new [Meter]
// that counts the total bytes read from r.
func NewReadMeter(r io.Reader) *Meter {
	return &Meter{reader: r}
}

// NewWriteMeter returns a new [Meter]
// that counts the total bytes written to w.
func NewWriteMeter(w io.Writer) *Meter {
	return &Meter{writer: w}
}

// NewReadWriteMeter returns a new [Meter]
// that counts the total bytes read from and written to rw.
func NewReadWriteMeter(rw io.ReadWriter) *Meter {
	return &Meter{reader: rw, writer: rw}
}

// Read writes the underlying [io.Reader] to p
// and increments the total bytes read by n.
//
// See [io.Reader] for details.
func (m *Meter) Read(p []byte) (n int, err error) {
	if m.reader == nil {
		return 0, io.ErrClosedPipe
	}
	n, err = m.reader.Read(p)
	_ = m.IncRead(int64(n))
	return
}

// ReadFrom copies from r to the underlying [io.Writer]
// and increments the total bytes written by n.
//
// See [io.ReaderFrom] for details.
func (m *Meter) ReadFrom(r io.Reader) (n int64, err error) {
	if m.writer == nil {
		return 0, io.ErrClosedPipe
	}
	n, err = io.Copy(m.writer, r)
	_ = m.IncWritten(n)
	return
}

// Write writes p to the underlying [io.Writer]
// and increments the total bytes written by n.
//
// See [io.Writer] for details.
func (m *Meter) Write(p []byte) (n int, err error) {
	if m.writer == nil {
		return 0, io.ErrClosedPipe
	}
	n, err = m.writer.Write(p)
	_ = m.IncWritten(int64(n))
	return
}

// WriteTo copies from the underlying [io.Reader] to w
// and increments the total bytes read by n.
//
// See [io.WriterTo] for details.
func (m *Meter) WriteTo(w io.Writer) (n int64, err error) {
	if m.reader == nil {
		return 0, io.ErrClosedPipe
	}
	n, err = io.Copy(w, m.reader)
	_ = m.IncRead(n)
	return
}

// Count returns the total bytes read and written.
func (m *Meter) Count() (r, w int64) {
	return m.CountRead(), m.CountWritten()
}

// CountRead returns the total bytes read.
func (m *Meter) CountRead() int64 {
	return m.rCount.Load()
}

// CountWritten returns the total bytes written.
func (m *Meter) CountWritten() int64 {
	return m.wCount.Load()
}

// Inc increments the total bytes read and written by nr and nw, respectively,
// and returns the new byte counts.
func (m *Meter) Inc(nr, nw int64) (r, w int64) {
	return m.IncRead(nr), m.IncWritten(nw)
}

// IncRead increments the total bytes read by n
// and returns the new byte count.
func (m *Meter) IncRead(n int64) int64 {
	return m.rCount.Add(n)
}

// IncWritten increments the total bytes written by n
// and returns the new byte count.
func (m *Meter) IncWritten(n int64) int64 {
	return m.wCount.Add(n)
}

// Close closes each underlying interface that implements [io.Closer].
//
// See [io.Closer] for details.
func (m *Meter) Close() error {
	return m.close(m.reader, m.writer)
}

func (m *Meter) close(v ...interface{}) (err error) {
	for _, v := range v {
		if c, ok := v.(io.Closer); ok {
			err = errors.Join(err, c.Close())
		}
	}
	return
}
