package valve_test

type mockReadWriteCloser struct{ rerr, werr, cerr error }

func (m mockReadWriteCloser) Read([]byte) (int, error)  { return 0, m.rerr }
func (m mockReadWriteCloser) Write([]byte) (int, error) { return 0, m.werr }
func (m mockReadWriteCloser) Close() error              { return m.cerr }
