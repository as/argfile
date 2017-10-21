package argfile

import (
	"os"
	"io"
	"bufio"
	"fmt"
	"sync"
)

var (
	MaxFD   = 1024
	ticket = make(chan struct{}, MaxFD)
)

type File struct {
	io.ReadCloser
	Name           string
	closefn        func() error
	wg *sync.WaitGroup
}

func (fd *File) Close() {
	fd.closefn()
	fd.wg.Done()
	<- ticket
}

func emit(to chan *File, args ...string) {
	var wg = new(sync.WaitGroup)
	if len(args) == 0 {
		ticket <- struct{}{}
		wg.Add(1)
		to <- &File{os.Stdin, "/dev/stdin", os.Stdin.Close, wg}
		close(to)
		return
	}

	emitfd := func(n string) {
		ticket <- struct{}{}
		wg.Add(1)
		fd, err := os.Open(n)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			fd.Close()
		} else {
			to <- &File{Name: n, ReadCloser: fd, closefn: fd.Close, wg: wg}
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
	wg.Wait()
	close(ticket)
	close(to)
}

func Next(args ...string) (to chan *File) {
	to = make(chan *File)
	go emit(to, args...)
	return to
}
