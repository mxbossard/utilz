package filez

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

const (
	bufferSize    = 1000 // adapted to read an index of capacity 256 => (256 + 2) * 4 = 1032
	emptyPosition = int32(-1)
)

type BlocOrdering int

const (
	TopToBottom BlocOrdering = iota
	BottomToTop
)

var (
	ErrNotExist = errors.New("bloc does not exist")
)

type BlocUid struct {
	filepath string
	id       int
}

type Bloc struct {
	uid     BlocUid
	data    *[]byte
	updated bool
	cursor  int
}

func (b Bloc) Len() (n int) {
	if b.data == nil {
		return 0
	}
	return len(*b.data)
}

func (b *Bloc) Read(p []byte) (n int, err error) {
	n = copy(p, (*b.data)[b.cursor:])
	b.cursor += n
	if n < len(p) {
		err = io.EOF
	}
	return
}

type BlocCursor struct {
	ordering    BlocOrdering
	file        *BlocFile
	cursor      int
	current     *Bloc
	readPointer int
}

func (c BlocCursor) HasNext() bool {
	return c.cursor < c.file.Len()
}

func (c *BlocCursor) Next() (*Bloc, error) {
	if c.cursor >= c.file.Len() {
		// panic("reached end of cursor")
		return nil, ErrNotExist
	}
	var err error
	c.current, err = c.file.Get(c.cursor)
	c.cursor++
	return c.current, err
}

// TODO: add a WriteTo() method ?

// Read all Blocs sequentially
// Cannot use Next() & Read() concurrently
func (c *BlocCursor) Read(p []byte) (n int, err error) {
	// Loop until p is full
	for n < len(p) {
		// fmt.Printf("looping on blocs; n=%d ; cursor: %d ; pointer: %d\n", n, c.cursor, n)
		if c.current == nil && c.HasNext() {
			c.Next()
		} else {
			return n, io.EOF
		}

		var i int
		i, err = c.current.Read(p[n:])
		c.readPointer += i
		n += i
		if err == io.EOF {
			// End of Bloc
			c.readPointer = 0
			c.current = nil // Clear current block which reached EOF
			if n == len(p) {
				// Limit case: Bloc ended & p full
				return n, nil
			} else {
				// Aggregate next Bloc
				err = nil
			}
		} else if err != io.EOF {
			// Error
			return
		}
	}
	return
}

// Write by blocs.
// Can only write on last bloc. Previous blocks are accessible Read Only.
type BlocFile struct {
	*sync.Mutex
	file      *os.File
	filepath  string
	length    int32
	capacity  int32
	positions map[int32]int32
	lengths   map[int32]int32
	cache     []*Bloc
}

func NewBlocFile(filepath string, cap int32) (*BlocFile, error) {
	f, err := os.OpenFile(filepath, os.O_RDWR+os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("error opening bloc file: %w", err)
	}
	bf := &BlocFile{
		Mutex:     &sync.Mutex{},
		file:      f,
		filepath:  filepath,
		capacity:  cap,
		positions: make(map[int32]int32),
		lengths:   make(map[int32]int32),
	}
	err = bf.initIndex(cap)
	if err != nil {
		return nil, fmt.Errorf("error initing bloc file index: %w", err)
	}

	err = bf.buildIndexCache()
	if err != nil {
		return nil, fmt.Errorf("error building bloc file index cache: %w", err)
	}

	return bf, nil
}

func (f *BlocFile) initIndex(capacity int32) error {
	// f.Lock()
	// defer f.Unlock()

	firstPos := (capacity + 2) * 4
	buf := make([]byte, firstPos)

	// Write capacity
	_, err := binary.Encode(buf[0:4], binary.BigEndian, capacity)
	if err != nil {
		return fmt.Errorf("error encoding bloc file capacity: %w", err)
	}

	// Write first position
	_, err = binary.Encode(buf[4:8], binary.BigEndian, firstPos)
	if err != nil {
		return fmt.Errorf("error encoding bloc file first pos: %w", err)
	}
	f.positions[0] = firstPos
	// fmt.Printf("written Idx of cap: %d, first pos: %d\n", capacity, firstPos)

	// Init next empty positions
	var k int32
	for k = 2; k < capacity+2; k++ {
		offset := k * 4
		_, err := binary.Encode(buf[offset:offset+4], binary.BigEndian, emptyPosition)
		if err != nil {
			return fmt.Errorf("error initializing bloc file positions: %w", err)
		}
	}

	_, err = f.file.Write(buf)
	if err != nil {
		return fmt.Errorf("error writing bloc file index: %w", err)
	}

	f.length = 0
	f.capacity = capacity
	return nil
}

// Update last opened bloc length
func (f *BlocFile) updateLastBlocLength(length int32) error {
	// f.Lock()
	// defer f.Unlock()

	lastBlocId := f.length - 1
	lastPos := f.positions[lastBlocId]
	nextPos := lastPos + length

	// Write next position
	buf := make([]byte, 4)
	_, err := binary.Encode(buf, binary.BigEndian, nextPos)
	if err != nil {
		return fmt.Errorf("error encoding bloc file next position: %w", err)
	}

	nextBlocId := lastBlocId + 1
	// After next bloc Idx pos
	offset := int64((nextBlocId + 1) * 4)
	_, err = f.file.WriteAt(buf, offset)
	if err != nil {
		return fmt.Errorf("error writing bloc file last position: %w", err)
	}
	// fmt.Printf("updated Idx next block #%d pos: %d at: %d\n", nextBlocId, nextPos, offset)
	f.positions[f.length] = nextPos
	f.lengths[f.length-1] = length

	return nil
}

func (f *BlocFile) buildIndexCache() error {
	// FIXME: implement use case of larger capacity than buffer size ?
	// The file begins with the bloc index.
	// Format for n Bloc : [<CAPACITY><POS0><POS1><POS2>...<POSn-1><POSn><-1><-1><-1><-1>]
	// 4 bytes for each integer
	// Bloc len(i) = pos(i+1) - pos(i)
	// f.Lock()
	// defer f.Unlock()

	buf := make([]byte, bufferSize)
	n, err := f.file.ReadAt(buf, 0)
	if err == io.EOF {
		err = nil // Swallow EOF
	} else if err != nil {
		return fmt.Errorf("error reading bloc file index: %w", err)
	}
	if n > 0 {
		var capacity int32
		_, err := binary.Decode(buf[0:4], binary.BigEndian, &capacity)
		if err != nil {
			return fmt.Errorf("error decoding bloc file capcaity: %w", err)
		}
		f.capacity = capacity
		var k int32
		for k = 0; k < capacity+1; k++ {
			offset := (k + 1) * 4
			var pos int32
			_, err := binary.Decode(buf[offset:offset+4], binary.BigEndian, &pos)
			if err != nil {
				return fmt.Errorf("error decoding bloc file position: %w", err)
			}
			if int32(pos) == emptyPosition {
				// Stop parsing index no bloc opened after a negative value
				f.length = k - 1
				break
			}
			f.positions[k] = int32(pos)
			if k > 0 {
				len := int32(f.positions[k] - f.positions[k-1])
				f.lengths[k-1] = len
			}
		}
	}
	// fmt.Printf("cached capacity: %d ; positions cache state: %v\n", f.capacity, f.positions)
	return nil
}

// Return the number of written blocs
func (f BlocFile) Len() int {
	f.Lock()
	defer f.Unlock()
	return int(f.length)
}

// Return the bloc capacity
func (f BlocFile) Cap() int {
	f.Lock()
	defer f.Unlock()
	return int(f.capacity)
}

func (f *BlocFile) Cursor(ordering BlocOrdering) *BlocCursor {
	return &BlocCursor{
		ordering: ordering,
		file:     f,
	}
}

func (f BlocFile) Get(k int) (*Bloc, error) {
	f.Lock()
	defer f.Unlock()
	k32 := int32(k)
	if k32 >= f.length {
		// panic("Get bloc out of bounds")
		return nil, ErrNotExist
	}

	pos := f.positions[k32]
	len := f.lengths[k32]
	buf := make([]byte, len)
	n, err := f.file.ReadAt(buf, int64(pos))
	if err != nil {
		return nil, err
	}
	if int32(n) != len {
		panic("bad count of bytes read")
	}
	// fmt.Printf("read bloc at pos: %d with len: %d\n", pos, len)
	b := &Bloc{
		uid: BlocUid{
			filepath: f.filepath,
			id:       k,
		},
		data: &buf,
	}

	return b, nil
}

func (f *BlocFile) updateLastBloc(data []byte) (*Bloc, error) {
	k := f.length - 1
	pos := f.positions[k]
	n, err := f.file.WriteAt(data, int64(pos))
	if err != nil {
		return nil, err
	}
	if n < len(data) {
		panic("bad count of bytes writen")
	}
	// fmt.Printf("updated last bloc #%d at pos: %d with %d bytes\n", k, pos, n)
	f.updateLastBlocLength(int32(n))
	b := &Bloc{
		uid: BlocUid{
			filepath: f.filepath,
			id:       int(k),
		},
		data: &data,
	}
	return b, nil
}

func (f *BlocFile) UpdateLastBloc(data []byte) (*Bloc, error) {
	f.Lock()
	defer f.Unlock()

	if f.length == 0 {
		return f.WriteNewBloc(data)
	}

	return f.updateLastBloc(data)
}

// Write into a new Bloc.
// Return EOF if max capacity is reached.
func (f *BlocFile) WriteNewBloc(data []byte) (*Bloc, error) {
	f.Lock()
	defer f.Unlock()

	if f.length >= f.capacity {
		return nil, io.EOF
	}

	//  Create new bloc
	f.length++

	return f.updateLastBloc(data)
}

type VirtualBlocFile struct {
	blocks []BlocUid
}

// Return the number of written blocs
func (f VirtualBlocFile) Len() int {
	panic("not implemented yet")
}

func (f VirtualBlocFile) Cursor(limit int) BlocCursor {
	panic("not implemented yet")
}

func (f VirtualBlocFile) Get(n int) (Bloc, error) {
	panic("not implemented yet")
}
