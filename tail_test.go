package tail

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestTail(t *testing.T) {
	f, err := ioutil.TempFile("/tmp", "log")
	assert.NoError(t, err)
	defer os.Remove(f.Name())

	tail, err := NewTail(f.Name(), 0, 100*time.Millisecond)
	assert.NoError(t, err)
	defer tail.Close()

	lines := make(chan string, 10)
	go func() {
		for {
			lines <- tail.ReadLine()
		}
	}()

	f.WriteString("foo\n")
	assert.Equal(t, "foo", <-lines)

	// append
	f.WriteString("bar\nbuz\n")
	assert.Equal(t, "bar", <-lines)
	assert.Equal(t, "buz", <-lines)

	// move rotation
	err = os.Rename(f.Name(), f.Name()+".1")
	assert.NoError(t, err)
	defer os.Remove(f.Name() + ".1")
	f, err = os.Create(f.Name())
	assert.NoError(t, err)
	f.WriteString("foo1\nbar1\n")
	assert.Equal(t, "foo1", <-lines)
	assert.Equal(t, "bar1", <-lines)

	// truncate rotation
	f, err = os.OpenFile(f.Name(), os.O_WRONLY|os.O_TRUNC, 0)
	assert.NoError(t, err)
	f.WriteString("buz1\n")
	assert.Equal(t, "buz1", <-lines)

	// delete
	err = os.Remove(f.Name())
	assert.NoError(t, err)
	time.Sleep(3 * tail.pollInterval)
	f, err = os.Create(f.Name())
	assert.NoError(t, err)
	f.WriteString("foo2\n")
	assert.Equal(t, "foo2", <-lines)

	// no eol
	f.WriteString("bar2\nbu")
	time.Sleep(3 * tail.pollInterval)
	f.WriteString("z2\n")
	assert.Equal(t, "bar2", <-lines)
	assert.Equal(t, "buz2", <-lines)
}
