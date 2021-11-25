package stream_multiplex

import (
	"io"
	"net"
	"os"
	"sync"
	"syscall"
)

type Multiplex struct {
	buf        []byte
	in         io.Reader
	file       *os.File
	fileOffset int64
	err        error

	mu   sync.Mutex
	cond sync.Cond
}

func (m *Multiplex) CacheInFile() error {
	for {
		n, err := m.in.Read(m.buf)
		if err != nil {
			m.mu.Lock()
			m.err = err
			m.mu.Unlock()
			if err != io.EOF {
				return err
			}
		}
		if n > 0 {
			n, err = m.file.Write(m.buf[:n])
			if err != nil {
				m.mu.Lock()
				m.err = err
				m.mu.Unlock()
				return err
			}
			m.mu.Lock()
			m.fileOffset += int64(n)
			m.mu.Unlock()
		}
		m.cond.Broadcast()
		if m.err == io.EOF {
			return nil
		}
	}
}

func (m *Multiplex) addReceiver(conn *net.TCPConn) {
	file, _ := conn.File()
	var (
		outfd      int   = int(file.Fd())
		infd       int   = int(m.file.Fd())
		offset     int64 = 0
		fileOffset int64 = 0
		err        error
		writeSize  int64 = 64 * 1024
	)

	for {
		m.mu.Lock()
		for fileOffset, err = m.fileOffset, m.err; fileOffset == offset && err == nil; m.cond.Wait() {
		}
		m.mu.Unlock()

		for offset < fileOffset {
			if fileOffset-offset < writeSize {
				writeSize = fileOffset - offset
			}
			_, err := syscall.Sendfile(outfd, infd, &offset, int(writeSize))
			if err != nil {
				return
			}
		}

		if err != nil {
			return
		}
	}
}
