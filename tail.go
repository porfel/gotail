package tail

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type Tail struct {
	fileName     string
	pollInterval time.Duration
	file         *os.File
	stat         os.FileInfo
	reader       *bufio.Reader
	rel          bool
}

func NewTail(fileName string, offset int64, pollInterval time.Duration) (tail Tail, err error) {
	tail.fileName = fileName
	tail.pollInterval = pollInterval
	tail.file, err = os.Open(fileName)
	tail.rel = false
	if err != nil {
		return tail, fmt.Errorf("failed to open file %s: %s", fileName, err)
	}

	tail.stat, err = tail.file.Stat()
	if err != nil {
		return tail, fmt.Errorf("failed to stat file %s: %s", fileName, err)
	}

	if offset != 0 {
		_, err = tail.file.Seek(offset, os.SEEK_SET)
		if err != nil {
			return tail, fmt.Errorf("failed to seek file %s: %s", fileName, err)
		}
	}

	tail.reader = bufio.NewReader(tail.file)

	return tail, nil
}

func (tail *Tail) ReadLine() string {
	var linePart string
	for {
		if tail.rel {
			log.Println("tailer: Return empty line")
			tail.rel = false
			return ""
		}
		line, err := tail.reader.ReadString('\n')
		if err == nil {
			if linePart != "" {
				line = linePart + line
				linePart = ""
			}
			return strings.TrimRight(line, "\n")
		}
		linePart = line
		tail.waitForChanges()
	}
}

// Request for next or current call to ReadLine returns the empty line
func (tail *Tail) RequestEmptyLine() {
	tail.rel = true
}

func (tail *Tail) waitForChanges() {
	var stat os.FileInfo
	var err error
	for {
		time.Sleep(tail.pollInterval)
		stat, err = os.Stat(tail.fileName)
		if tail.rel {
			log.Println("tailer: stop cycle required")
			break
		}
		if err != nil {
			log.Printf("failed to stat file %s: %s", tail.fileName, err)
			continue
		}
		if !os.SameFile(tail.stat, stat) {
			log.Printf("file was moved %s", tail.fileName)
			tail.file.Close()
			tail.file, err = os.Open(tail.fileName)
			if err != nil {
				log.Printf("failed to open file %s: %s", tail.fileName, err)
				continue
			}
			tail.reader = bufio.NewReader(tail.file)
			break
		}
		if stat.Size() < tail.stat.Size() {
			log.Printf("file was truncated %s", tail.fileName)
			_, err = tail.file.Seek(0, os.SEEK_SET)
			if err != nil {
				log.Printf("failed to seek file %s: %s", tail.fileName, err)
				continue
			}
			break
		}
		if stat.Size() > tail.stat.Size() {
			break
		}
		if tail.rel {
			break
		}
	}
	if stat != nil {
		tail.stat = stat
	}
}

func (tail *Tail) Close() {
	if tail.file != nil {
		tail.file.Close()
	}
}

func (tail *Tail) Offset() (int64, error) {
	offset, err := tail.file.Seek(0, os.SEEK_CUR)
	if err != nil {
		return 0, err
	}
	return offset - int64(tail.reader.Buffered()), nil
}
