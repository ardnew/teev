package valve_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/ardnew/valve"
	"github.com/stretchr/testify/require"
)

func TestMeteredReader_Read(t *testing.T) {
	t.Parallel()
	buf := []byte("Hello, World!")
	num := len(buf)
	testReader := bytes.NewBuffer(buf)
	meteredReader := valve.NewReadMeter(testReader)
	testBuffer := make([]byte, num)
	n, err := meteredReader.Read(testBuffer) //nolint: varnamelen
	require.NoError(t, err)
	require.Equal(t, num, n)
	require.Equal(t, int64(num), meteredReader.CountRead())
	require.True(t, bytes.Equal(buf, testBuffer))
	missingReader := valve.Meter{}
	n, err = missingReader.Read(buf)
	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Zero(t, n)
}

func TestMeteredReader_WriteTo(t *testing.T) {
	t.Parallel()
	buf := []byte("Hello, World!")
	num := len(buf)
	testReader := bytes.NewBuffer(buf)
	meteredReader := valve.NewReadMeter(testReader)
	testBuffer := &bytes.Buffer{}
	n, err := meteredReader.WriteTo(testBuffer) //nolint: varnamelen
	require.NoError(t, err)
	require.Equal(t, int64(num), n)
	require.Equal(t, int64(num), meteredReader.CountRead())
	require.True(t, bytes.Equal(buf, testBuffer.Bytes()))
	missingReader := valve.Meter{}
	n, err = missingReader.WriteTo(testBuffer)
	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Zero(t, n)
}

func TestMeteredWriter_Write(t *testing.T) {
	t.Parallel()
	buf := []byte("Hello, World!")
	num := len(buf)
	testWriter := &bytes.Buffer{}
	meteredWriter := valve.NewWriteMeter(testWriter)
	n, err := meteredWriter.Write(buf) //nolint: varnamelen
	require.NoError(t, err)
	require.Equal(t, num, n)
	require.Equal(t, int64(num), meteredWriter.CountWritten())
	require.True(t, bytes.Equal(buf, testWriter.Bytes()))
	missingWriter := valve.Meter{}
	n, err = missingWriter.Write(buf)
	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Zero(t, n)
}

func TestMeteredWriter_ReadFrom(t *testing.T) {
	t.Parallel()
	buf := []byte("Hello, World!")
	num := len(buf)
	testWriter := &bytes.Buffer{}
	meteredWriter := valve.NewWriteMeter(testWriter)
	n, err := meteredWriter.ReadFrom(bytes.NewReader(buf)) //nolint: varnamelen
	require.NoError(t, err)
	require.Equal(t, int64(num), n)
	require.Equal(t, int64(num), meteredWriter.CountWritten())
	require.True(t, bytes.Equal(buf, testWriter.Bytes()))
	missingWriter := valve.Meter{}
	n, err = missingWriter.ReadFrom(bytes.NewReader(buf))
	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Zero(t, n)
}

func TestMeter(t *testing.T) {
	t.Parallel()
	buf := []byte("Hello, World!")
	num := len(buf)
	meter := valve.NewMeter(
		bytes.NewBuffer(buf),
		&bytes.Buffer{},
	)
	n, err := io.Copy(meter, meter)
	require.NoError(t, err)
	require.Equal(t, int64(num), n)
	nr, nw := meter.Count()
	require.Equal(t, int64(num), nr)
	require.Equal(t, int64(num), nw)
	ir, iw := meter.Inc(-int64(num), +int64(num))
	require.Zero(t, ir)
	require.Equal(t, 2*int64(num), iw)
	readWriteMeter := valve.NewReadWriteMeter(mockReadWriteCloser{})
	err = readWriteMeter.Close()
	require.NoError(t, err)
}

type mockReadWriteCloser struct{ rerr, werr, cerr error }

func (m mockReadWriteCloser) Read([]byte) (int, error)  { return 0, m.rerr }
func (m mockReadWriteCloser) Write([]byte) (int, error) { return 0, m.werr }
func (m mockReadWriteCloser) Close() error              { return m.cerr }
