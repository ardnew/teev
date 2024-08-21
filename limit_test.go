package valve_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/ardnew/valve"
	"github.com/ardnew/valve/internal"
	"github.com/stretchr/testify/require"
)

//nolint: gochecknoglobals
var (
	limitSrcBuf = []byte("Hello, World!")
	limitSrcLen = len(limitSrcBuf)
	limitExpBuf = limitSrcBuf[:limitExpLen]
	limitExpLen = limitSrcLen - 8
)

func TestLimit(t *testing.T) {
	t.Parallel()

	limit := valve.NewLimit(
		bytes.NewBuffer(limitSrcBuf), int64(limitExpLen),
		bytes.NewBuffer(limitSrcBuf), int64(limitExpLen+1),
	)

	n, err := io.Copy(limit, limit)

	require.NoErrorf(t, err, "[%+v]", err)
	require.Equal(t, int64(limitExpLen), n)
	require.NotNil(t, limit)
}

//nolint: varnamelen
func TestLimit_Read(t *testing.T) {
	t.Parallel()

	reader := valve.NewReadLimit(bytes.NewReader(limitSrcBuf), int64(limitExpLen))
	buffer := make([]byte, limitSrcLen)
	n, err := reader.Read(buffer)
	expErr := reader.MakeReadLimitError(int64(limitSrcLen), int64(limitExpLen))

	require.ErrorIsf(t, err, expErr, "[%+v] != [%+v]", err, expErr)
	require.Equal(t, err.Error(), expErr.Error())
	require.Equal(t, limitExpLen, n)
	require.Equal(t, int64(limitExpLen), reader.CountRead())
	require.Truef(t, bytes.Equal(limitExpBuf, buffer[:n]), "[% x] != [% x]", limitExpBuf, buffer[:n])
}

//nolint: varnamelen
func TestLimit_ReadUnlimited(t *testing.T) {
	t.Parallel()

	reader := valve.NewReadLimit(bytes.NewReader(limitExpBuf), valve.Unlimited)
	buffer := make([]byte, limitSrcLen)
	n1, err1 := reader.Read(buffer)
	n2, err2 := reader.Read(buffer)
	expBuf := make([]byte, limitSrcLen)
	copy(expBuf, limitExpBuf)

	require.NoError(t, err1)
	require.ErrorIsf(t, err2, io.EOF, "[%+v] != [%+v]", err2, io.EOF)
	require.Equal(t, limitExpLen, n1)
	require.Zero(t, n2)
	require.Equal(t, int64(limitExpLen), reader.CountRead())
	require.True(t, bytes.Equal(expBuf, buffer), "[% x] != [% x]", expBuf, buffer)
}

//nolint: varnamelen
func TestLimit_ReadLimited(t *testing.T) {
	t.Parallel()

	reader := valve.NewReadLimit(bytes.NewReader(limitSrcBuf), int64(limitExpLen))
	buffer := make([]byte, limitSrcLen)
	n1, err1 := reader.Read(buffer)
	n2, err2 := reader.Read(buffer)
	expErr1 := reader.MakeReadLimitError(int64(limitSrcLen), int64(limitExpLen))
	expErr2 := reader.MakeReadLimitError(int64(limitSrcLen), 0)
	expBuf1 := make([]byte, limitSrcLen)
	copy(expBuf1, limitExpBuf)

	require.ErrorIsf(t, err1, expErr1, "[%+v] != [%+v]", err1, expErr1)
	require.ErrorIsf(t, err2, expErr2, "[%+v] != [%+v]", err2, expErr2)
	require.Equal(t, limitExpLen, n1)
	require.Zero(t, n2)
	require.Equal(t, int64(limitExpLen), reader.CountRead())
	require.True(t, bytes.Equal(expBuf1, buffer), "[% x] != [% x]", expBuf1, buffer)
}

func TestLimit_ReadShort(t *testing.T) {
	t.Parallel()

	reader := valve.NewReadLimit(bytes.NewReader(limitExpBuf), int64(limitSrcLen))
	buffer := make([]byte, limitSrcLen)
	n, err := reader.Read(buffer)
	expBuf := make([]byte, limitSrcLen)
	copy(expBuf, limitExpBuf)

	require.NoError(t, err)
	require.Equal(t, limitExpLen, n)
	require.Equal(t, int64(limitExpLen), reader.CountRead())
	require.True(t, bytes.Equal(expBuf, buffer), "[% x] != [% x]", expBuf, buffer)
}

func TestLimit_ReadWithoutReader(t *testing.T) {
	t.Parallel()

	reader := valve.Limit{}
	buffer := make([]byte, limitSrcLen)
	n, err := reader.Read(buffer)

	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Zero(t, n)
}

func TestLimit_ReadFrom(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	writer := valve.NewWriteLimit(buffer, int64(limitExpLen))
	n, err := writer.ReadFrom(bytes.NewReader(limitSrcBuf))

	require.NoError(t, err)
	require.Equal(t, int64(limitExpLen), n)
	require.Equal(t, int64(limitExpLen), writer.CountWrite())
	require.True(t, bytes.Equal(limitExpBuf, buffer.Bytes()), "[% x] != [% x]", limitExpBuf, buffer.Bytes())
}

//nolint: varnamelen
func TestLimit_ReadFromUnlimited(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	writer := valve.NewWriteLimit(buffer, valve.Unlimited)
	n1, err1 := writer.ReadFrom(bytes.NewReader(limitSrcBuf))
	n2, err2 := writer.ReadFrom(bytes.NewReader(limitSrcBuf))
	expBuf := limitSrcBuf
	expBuf = append(expBuf, limitSrcBuf...)

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.Equal(t, int64(limitSrcLen), n1)
	require.Equal(t, int64(limitSrcLen), n2)
	require.Equal(t, int64(limitSrcLen*2), writer.CountWrite())
	require.True(t, bytes.Equal(expBuf, buffer.Bytes()), "[% x] != [% x]", expBuf, buffer.Bytes())
}

func TestLimit_ReadFromLimited(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	writer := valve.NewWriteLimit(buffer, int64(limitSrcLen))
	nPass, errPass := writer.ReadFrom(bytes.NewReader(limitSrcBuf))
	nFail, errFail := writer.ReadFrom(bytes.NewReader(limitSrcBuf))
	// Requested 0 bytes, because a Reader does not reveal its size.
	expErr := writer.MakeWriteLimitError(0, 0)

	require.NoError(t, errPass)
	require.ErrorIsf(t, errFail, expErr, "[%+v] != [%+v]", errFail, expErr)
	require.Equal(t, int64(limitSrcLen), nPass)
	require.Zero(t, nFail)
	require.Equal(t, int64(limitSrcLen), writer.CountWrite())
	require.True(t, bytes.Equal(limitSrcBuf, buffer.Bytes()), "[% x] != [% x]", limitSrcBuf, buffer.Bytes())
}

func TestLimit_ReadFromWithoutWriter(t *testing.T) {
	t.Parallel()

	writer := valve.Limit{}
	n, err := writer.ReadFrom(bytes.NewReader(limitSrcBuf))

	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Zero(t, n)
}

func TestLimit_Write(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	writer := valve.NewWriteLimit(buffer, int64(limitExpLen))
	n, err := writer.Write(limitSrcBuf)
	expErr := writer.MakeWriteLimitError(int64(limitSrcLen), int64(limitExpLen))

	require.ErrorIsf(t, err, expErr, "[%+v] != [%+v]", err, expErr)
	require.Equal(t, err.Error(), expErr.Error())
	require.Equal(t, limitExpLen, n)
	require.Equal(t, int64(limitExpLen), writer.CountWrite())
	require.True(t, bytes.Equal(limitExpBuf, buffer.Bytes()), "[% x] != [% x]", limitExpBuf, buffer.Bytes())
}

//nolint: varnamelen
func TestLimit_WriteUnlimited(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	writer := valve.NewWriteLimit(buffer, valve.Unlimited)
	n1, err1 := writer.Write(limitSrcBuf)
	n2, err2 := writer.Write(limitSrcBuf)
	expBuf := limitSrcBuf
	expBuf = append(expBuf, limitSrcBuf...)

	require.NoErrorf(t, err1, "[%+v]", err1)
	require.NoErrorf(t, err2, "[%+v]", err2)
	require.Equal(t, limitSrcLen, n1)
	require.Equal(t, limitSrcLen, n2)
	require.Equal(t, int64(limitSrcLen*2), writer.CountWrite())
	require.True(t, bytes.Equal(expBuf, buffer.Bytes()), "[% x] != [% x]", expBuf, buffer.Bytes())
}

func TestLimit_WriteLimited(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	writer := valve.NewWriteLimit(buffer, int64(limitSrcLen))
	nPass, errPass := writer.Write(limitSrcBuf)
	nFail, errFail := writer.Write(limitSrcBuf)
	expErr := writer.MakeWriteLimitError(int64(limitSrcLen), 0)

	require.NoError(t, errPass)
	require.ErrorIsf(t, errFail, expErr, "[%+v] != [%+v]", errFail, expErr)
	require.Equal(t, limitSrcLen, nPass)
	require.Zero(t, nFail)
	require.Equal(t, int64(limitSrcLen), writer.CountWrite())
	require.True(t, bytes.Equal(limitSrcBuf, buffer.Bytes()), "[% x] != [% x]", limitSrcBuf, buffer.Bytes())
}

func TestLimit_WriteWithoutWriter(t *testing.T) {
	t.Parallel()

	writer := valve.Limit{}
	n, err := writer.Write(limitSrcBuf)

	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Zero(t, n)
}

func TestLimit_WriteTo(t *testing.T) {
	t.Parallel()

	reader := valve.NewReadLimit(bytes.NewReader(limitSrcBuf), int64(limitExpLen))
	buffer := &bytes.Buffer{}
	n, err := reader.WriteTo(buffer)

	require.NoError(t, err)
	require.Equal(t, int64(limitExpLen), n)
	require.Equal(t, int64(limitExpLen), reader.CountRead())
	require.True(t, bytes.Equal(limitExpBuf, buffer.Bytes()), "[% x] != [% x]", limitExpBuf, buffer.Bytes())
}

// nolint: varnamelen
func TestLimit_WriteToUnlimited(t *testing.T) {
	t.Parallel()

	reader := valve.NewReadLimit(bytes.NewReader(limitExpBuf), valve.Unlimited)
	buffer := &bytes.Buffer{}
	n1, err1 := reader.WriteTo(buffer)
	n2, err2 := reader.WriteTo(buffer)

	require.NoError(t, err1, "[%+v]", err1)
	require.NoError(t, err2, "[%+v]", err2)
	require.Equal(t, int64(limitExpLen), n1)
	require.Zero(t, n2)
	require.Equal(t, int64(limitExpLen), reader.CountRead())
	require.True(t, bytes.Equal(limitExpBuf, buffer.Bytes()), "[% x] != [% x]", limitExpBuf, buffer.Bytes())
}

//nolint: varnamelen
func TestLimit_WriteToLimited(t *testing.T) {
	t.Parallel()

	reader := valve.NewReadLimit(bytes.NewReader(limitSrcBuf), int64(limitExpLen-1))
	buffer := &bytes.Buffer{}
	n1, err1 := reader.WriteTo(buffer)
	n2, err2 := reader.WriteTo(buffer)
	expErr := reader.MakeReadLimitError(0, 0)
	expBuf1 := make([]byte, limitExpLen-1)
	copy(expBuf1, limitExpBuf)

	require.NoError(t, err1, "[%+v]", err1)
	require.ErrorIsf(t, err2, expErr, "[%+v] != [%+v]", err2, expErr)
	require.Equal(t, int64(limitExpLen-1), n1)
	require.Zero(t, n2)
	require.Equal(t, int64(limitExpLen-1), reader.CountRead())
	require.True(t, bytes.Equal(expBuf1, buffer.Bytes()), "[% x] != [% x]", expBuf1, buffer.Bytes())
}

func TestLimit_WriteToShort(t *testing.T) {
	t.Parallel()

	reader := valve.NewReadLimit(bytes.NewReader(limitExpBuf), int64(limitSrcLen))
	buffer := &bytes.Buffer{}
	n, err := reader.WriteTo(buffer)

	require.ErrorIsf(t, err, io.EOF, "[%+v] != [%+v]", err, io.EOF)
	require.Equal(t, int64(limitExpLen), n)
	require.Equal(t, int64(limitExpLen), reader.CountRead())
	require.True(t, bytes.Equal(limitExpBuf, buffer.Bytes()), "[% x] != [% x]", limitExpBuf, buffer.Bytes())
}

func TestLimit_WriteToWithoutReader(t *testing.T) {
	t.Parallel()

	reader := valve.Limit{}
	buffer := &bytes.Buffer{}
	n, err := reader.WriteTo(buffer)

	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Zero(t, n)
}

func TestLimit_Close(t *testing.T) {
	t.Parallel()

	zero := valve.Limit{}
	base := valve.Limit{Meter: &valve.Meter{}}

	require.NoError(t, zero.Close())
	require.NoError(t, base.Close())
}

func TestLimit_MaxCount(t *testing.T) {
	t.Parallel()

	limit := valve.NewReadWriteLimit(bytes.NewBuffer(limitSrcBuf), int64(limitExpLen), int64(limitExpLen+1))
	rMax, wMax := limit.MaxCount()

	require.Equal(t, int64(limitExpLen), rMax)
	require.Equal(t, int64(limitExpLen+1), wMax)
}

//nolint: varnamelen
func TestLimit_RemainingCount(t *testing.T) {
	t.Parallel()

	limit := valve.NewReadWriteLimit(bytes.NewBuffer(limitSrcBuf), int64(limitExpLen), int64(limitExpLen+1))
	buffer := make([]byte, limitExpLen-1)
	rn, rerr := limit.Read(buffer)
	wn, werr := limit.Write(buffer)
	rRem, wRem := limit.RemainingCount()

	require.NoError(t, rerr)
	require.NoError(t, werr)
	require.Equal(t, limitExpLen-1, rn)
	require.Equal(t, limitExpLen-1, wn)
	require.Equal(t, int64(1), rRem)
	require.Equal(t, int64(2), wRem)
}

func TestLimit_SetMaxCount(t *testing.T) {
	t.Parallel()

	limit := valve.NewReadWriteLimit(bytes.NewBuffer(limitSrcBuf), int64(limitExpLen), int64(limitExpLen))
	limit.SetMaxCount(int64(limitSrcLen), int64(limitSrcLen-1))
	rMax, wMax := limit.MaxCount()

	require.Equal(t, int64(limitSrcLen), rMax)
	require.Equal(t, int64(limitSrcLen-1), wMax)
}

func TestLimitError_Error(t *testing.T) {
	t.Parallel()

	// Parse YAML-formatted [valve.LimitError.Error] strings to [internal.Error]
	// so that they can be compared for equivalence using [errors.Is],
	// which will ignore the datetimes and stacktraces.
	err := internal.UnformatYAML(valve.LimitError{}.Error())
	exp := internal.UnformatYAML(internal.MakeInvalidOperationError().Error())

	require.ErrorIsf(t, err, exp, "[%+v] != [%+v]", err, exp)
}
