package sendfile

import (
	"os"
	"syscall"
)

func SendFile(infd int, f os.File) (int, error) {
	fd := int(f.Fd())
	return syscall.Sendfile(infd, fd, nil, 1024*1024)
}
