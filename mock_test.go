package valve_test

import (
	"io"
)

type mockError struct{ error }

func (m mockError) Unwrap() error { return m.error }
func (m mockError) Error() string { return m.error.Error() }

// nolint: errname
type mockBuffer struct {
	err error
	buf []byte
}

func (m mockBuffer) Close() error { return m.err }

func (m mockBuffer) Error() string {
	if m.err != nil {
		return m.err.Error()
	}
	return ""
}

// nolint: varnamelen
func (m mockBuffer) Read(p []byte) (n int, err error) {
	defer func() {
		if n > 0 && n == len(p) {
			err = nil
		}
	}()
	if m.err != nil {
		return 0, m.err
	}
	if len(p) == 0 {
		return 0, io.EOF
	}
	if len(p) < len(m.buf) {
		err = io.ErrShortBuffer
	}
	if len(p) > len(m.buf) {
		err = io.ErrUnexpectedEOF
	}
	return copy(p, m.buf), err
}

// nolint: varnamelen
func (m mockBuffer) Write(p []byte) (n int, err error) {
	defer func() {
		if n > 0 && n == len(p) {
			err = nil
		}
	}()
	if m.err != nil {
		return 0, m.err
	}
	if len(p) == 0 {
		return 0, io.EOF
	}
	if len(m.buf) < len(p) {
		err = io.ErrUnexpectedEOF
	}
	if len(m.buf) > len(p) {
		err = io.ErrShortWrite
	}
	return copy(m.buf, p), err
}

func makeMockBuffer() mockBuffer {
	return mockBuffer{mockError{}, []byte{}}
}

// func makeMockReader(p []byte, err error) mockBuffer {
// 	return mockBuffer{err, p}
// }

// func makeMockWriter(p []byte, err error) mockBuffer {
// 	return mockBuffer{err, p}
// }

func makeMockCloser(err error) mockBuffer {
	return mockBuffer{err, nil}
}
