package argfile

import (
	"os"
	"io"
	"bufio"
	"fmt"
)

var (
	MaxFD   = 1024
	ticket = make(chan struct{}, MaxFD)
)

type File struct {
	io.ReadCloser
	Name           string
	closefn        func() error
}

func (fd *File) Close() {
	<- ticket
	fd.closefn()
}

func emit(to chan *File, args ...string) {
	if len(args) == 0 {
		ticket <- struct{}{}
		to <- &File{os.Stdin, "/dev/stdin", os.Stdin.Close}
		close(to)
		<- ticket
		return
	}

	emitfd := func(n string) {
		ticket <- struct{}{}
		fd, err := os.Open(n)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			fd.Close()
			<-ticket
		} else {
			to <- &File{Name: n, ReadCloser: fd, closefn: fd.Close}
		}
	}

	for _, v := range args {
		if v != "-" {
			emitfd(v)
		} else {
			in := bufio.NewScanner(os.Stdin)
			for in.Scan() {
				emitfd(in.Text())
			}
		}
	}
	close(to)
}

func Next(args ...string) (to chan *File) {
	to = make(chan *File)
	go emit(to, args...)
	return to
}
