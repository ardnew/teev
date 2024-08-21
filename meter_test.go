package valve_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/ardnew/valve"
	"github.com/stretchr/testify/require"
)

//nolint: gochecknoglobals
var (
	meterSrcBuf = []byte("Hello, World!")
	meterSrcLen = len(meterSrcBuf)
)

func TestMeter_Read(t *testing.T) {
	t.Parallel()

	reader := valve.NewReadMeter(bytes.NewReader(meterSrcBuf))
	buffer := make([]byte, meterSrcLen)
	n, err := reader.Read(buffer)

	require.NoError(t, err)
	require.Equal(t, meterSrcLen, n)
	require.Equal(t, int64(meterSrcLen), reader.CountRead())
	require.True(t, bytes.Equal(meterSrcBuf, buffer))
}

func TestMeter_ReadWithoutConstructor(t *testing.T) {
	t.Parallel()

	reader := valve.Meter{Reader: bytes.NewReader(meterSrcBuf)}
	buffer := make([]byte, meterSrcLen)
	n, err := reader.Read(buffer)

	require.NoError(t, err)
	require.Equal(t, meterSrcLen, n)
	require.Equal(t, int64(meterSrcLen), reader.CountRead())
	require.True(t, bytes.Equal(meterSrcBuf, buffer))
}

func TestMeter_ReadWithoutReader(t *testing.T) {
	t.Parallel()

	reader := valve.Meter{}
	buffer := make([]byte, meterSrcLen)
	n, err := reader.Read(buffer)

	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Zero(t, n)
}

func TestMeter_ReadFrom(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	writer := valve.NewWriteMeter(buffer)
	n, err := writer.ReadFrom(bytes.NewReader(meterSrcBuf))

	require.NoError(t, err)
	require.Equal(t, int64(meterSrcLen), n)
	require.Equal(t, int64(meterSrcLen), writer.CountWrite())
	require.True(t, bytes.Equal(meterSrcBuf, buffer.Bytes()))
}

func TestMeter_ReadFromWithoutConstructor(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	writer := valve.Meter{Writer: buffer}
	n, err := writer.ReadFrom(bytes.NewReader(meterSrcBuf))

	require.NoError(t, err)
	require.Equal(t, int64(meterSrcLen), n)
	require.Equal(t, int64(meterSrcLen), writer.CountWrite())
	require.True(t, bytes.Equal(meterSrcBuf, buffer.Bytes()))
}

func TestMeter_ReadFromWithoutWriter(t *testing.T) {
	t.Parallel()

	writer := valve.Meter{}
	n, err := writer.ReadFrom(bytes.NewReader(meterSrcBuf))

	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Zero(t, n)
}

func TestMeter_Write(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	writer := valve.NewWriteMeter(buffer)
	n, err := writer.Write(meterSrcBuf)

	require.NoError(t, err)
	require.Equal(t, meterSrcLen, n)
	require.Equal(t, int64(meterSrcLen), writer.CountWrite())
	require.True(t, bytes.Equal(meterSrcBuf, buffer.Bytes()))
}

func TestMeter_WriteWithoutConstructor(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	writer := valve.Meter{Writer: buffer}
	n, err := writer.Write(meterSrcBuf)

	require.NoError(t, err)
	require.Equal(t, meterSrcLen, n)
	require.Equal(t, int64(meterSrcLen), writer.CountWrite())
	require.True(t, bytes.Equal(meterSrcBuf, buffer.Bytes()))
}

func TestMeter_WriteWithoutWriter(t *testing.T) {
	t.Parallel()

	writer := valve.Meter{}
	n, err := writer.Write(meterSrcBuf)

	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Zero(t, n)
}

func TestMeter_WriteTo(t *testing.T) {
	t.Parallel()

	reader := valve.NewReadMeter(bytes.NewReader(meterSrcBuf))
	buffer := &bytes.Buffer{}
	n, err := reader.WriteTo(buffer)

	require.NoError(t, err)
	require.Equal(t, int64(meterSrcLen), n)
	require.Equal(t, int64(meterSrcLen), reader.CountRead())
	require.True(t, bytes.Equal(meterSrcBuf, buffer.Bytes()))
}

func TestMeter_WriteToWithoutConstructor(t *testing.T) {
	t.Parallel()

	reader := valve.Meter{Reader: bytes.NewReader(meterSrcBuf)}
	buffer := &bytes.Buffer{}
	n, err := reader.WriteTo(buffer)

	require.NoError(t, err)
	require.Equal(t, int64(meterSrcLen), n)
	require.Equal(t, int64(meterSrcLen), reader.CountRead())
	require.True(t, bytes.Equal(meterSrcBuf, buffer.Bytes()))
}

func TestMeter_WriteToWithoutReader(t *testing.T) {
	t.Parallel()

	reader := valve.Meter{}
	buffer := &bytes.Buffer{}
	n, err := reader.WriteTo(buffer)

	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Zero(t, n)
}

func TestMeter_Close(t *testing.T) {
	t.Parallel()

	cerr := fmt.Errorf("close error: %w", io.EOF)
	pass := valve.Meter{}
	fail := valve.NewReadWriteMeter(makeMockCloser(cerr))

	require.NoError(t, pass.Close())
	require.ErrorIs(t, fail.Close(), cerr)
}

func TestMeter_Count(t *testing.T) {
	t.Parallel()

	count := makeMockBuffer()
	meter := valve.NewMeter(count, count)
	r, w := meter.Count()

	require.Zero(t, r)
	require.Zero(t, w)
}

func TestMeter_CountRead(t *testing.T) {
	t.Parallel()

	meter := valve.Meter{}
	count := meter.CountRead()

	require.Zero(t, count)
}

func TestMeter_CountWrite(t *testing.T) {
	t.Parallel()

	meter := valve.Meter{}
	count := meter.CountWrite()

	require.Zero(t, count)
}

func TestMeter_Inc(t *testing.T) {
	t.Parallel()

	meter := valve.Meter{}
	r, w := meter.AddCount(10, 20)

	require.Equal(t, int64(10), r)
	require.Equal(t, int64(20), w)
}

func TestMeter_IncRead(t *testing.T) {
	t.Parallel()

	meter := valve.Meter{}
	count := meter.AddCountRead(10)

	require.Equal(t, int64(10), count)
}

func TestMeter_IncWrite(t *testing.T) {
	t.Parallel()

	meter := valve.Meter{}
	count := meter.AddCountWrite(20)

	require.Equal(t, int64(20), count)
}

func TestMeter_Set(t *testing.T) {
	t.Parallel()

	meter := valve.Meter{}
	meter.SetCount(10, 20)
	r, w := meter.Count()

	require.Equal(t, int64(10), r)
	require.Equal(t, int64(20), w)
}

func TestMeter_SetRead(t *testing.T) {
	t.Parallel()

	meter := valve.Meter{}
	meter.SetCountRead(10)
	count := meter.CountRead()

	require.Equal(t, int64(10), count)
}

func TestMeter_SetWrite(t *testing.T) {
	t.Parallel()

	meter := valve.Meter{}
	meter.SetCountWrite(20)
	count := meter.CountWrite()

	require.Equal(t, int64(20), count)
}

func TestMeter_Reset(t *testing.T) {
	t.Parallel()

	meter := valve.Meter{}
	meter.ResetCount()
	r, w := meter.Count()

	require.Zero(t, r)
	require.Zero(t, w)
}

func TestMeter_ResetRead(t *testing.T) {
	t.Parallel()

	meter := valve.Meter{}
	meter.ResetCountRead()
	count := meter.CountRead()

	require.Zero(t, count)
}

func TestMeter_ResetWrite(t *testing.T) {
	t.Parallel()

	meter := valve.Meter{}
	meter.ResetCountWrite()
	count := meter.CountWrite()

	require.Zero(t, count)
}
