package iterz

import (
	"iter"
	"sort"
	"sync"

	"github.com/mxbossard/utilz/collectionz"
)

func BuildIntSeq(ints ...int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for _, i := range ints {
			if !yield(i) {
				return
			}
		}
	}
}

func BuildStringSeq(strings ...string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, s := range strings {
			if !yield(s) {
				return
			}
		}
	}
}

func Flatten[K any](it iter.Seq[K]) (r []K) {
	for k := range it {
		r = append(r, k)
	}
	return
}

func Flatten2[K, V any](it iter.Seq2[K, V]) (r1 []K, r2 []V) {
	for k, v := range it {
		r1 = append(r1, k)
		r2 = append(r2, v)
	}
	return
}

func Filter[K any](filter func(k K) bool, it iter.Seq[K]) iter.Seq[K] {
	return func(yield func(K) bool) {
		for i := range it {
			if filter(i) {
				if !yield(i) {
					break
				}
			}
		}
	}
}

func Filter2[K, V any](filter func(k K, v V) bool, it iter.Seq2[K, V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range it {
			if filter(k, v) {
				if !yield(k, v) {
					break
				}
			}
		}
	}
}

func Link[K, V any](itk iter.Seq[K], itv iter.Seq[V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		nextK, stop := iter.Pull(itk)
		nextV, stop := iter.Pull(itv)
		defer stop()
		for {
			k, ok := nextK()
			if !ok {
				return
			}
			v, ok := nextV()
			if !ok {
				return
			}
			if !yield(k, v) {
				return
			}
		}
	}
}

// Merge multiple iterators ordering items following compare func.
// If item A need to be returned before B, compare(A, B) MUST return a negative number
// If items MAY be returned at indifferent order compare(A, B) must return 0
// On compare returning 0 keep same iterator in priority.
// Ideas:
// - Could preload iterators with a buffer size to enhance oredering
func Merge[T any](compare func(a, b T) int, its ...iter.Seq[T]) iter.Seq[T] {
	stop := false
	chansVal := make(map[int]chan T)
	chans := &chansVal
	nextItemsVal := make(map[int]T)
	nextItems := &nextItemsVal
	mutex := &sync.Mutex{}

	init := func() {
		stop = false
		// m := &sync.Mutex{}

		// Launch goroutines to stream produced datas
		for pos, it := range its {
			// Make a buffered chan of size 1
			c := make(chan T, 1)
			mutex.Lock()
			(*chans)[pos] = c
			mutex.Unlock()
			go func() {
				// Continuously push next data
				defer close(c)
				for item := range it {
					c <- item
					if stop {
						break
					}
				}
			}()
		}

		// Init consume first item of each chan
		for pos, c := range *chans {
			e, ok := <-c
			if ok {
				(*nextItems)[pos] = e
			} else {
				mutex.Lock()
				delete(*chans, pos)
				mutex.Unlock()
			}
		}

		// fmt.Printf("cat paginer initialized chans: %d\n", len(*chans))
	}

	init()

	cat := func(yield func(T) bool) {
		defer func() {
			stop = true
		}()
		var positions []int
		// var lastItem T
		lastPos := -1
		looped := false
	End:
		for !looped || len(positions) > 0 {
			looped = true
			positions = collectionz.Keys(*nextItems)
			sort.Ints(positions)
			selectedPos := -1
			var selectedItem T
			// fmt.Printf("nextItems: %v ; selectedPos: %d ; lastPos: %d\n", *nextItems, selectedPos, lastPos)
			for _, pos := range positions {
				item, ok := (*nextItems)[pos]
				if !ok {
					// no more entries in chan
					delete(*nextItems, pos)
					selectedPos = -1
					// fmt.Printf("chan %d closed\n", pos)
					continue
				}

				if selectedPos == -1 {
					// Select first by default
					// c := compare(lastItem, item)
					selectedPos = pos
					selectedItem = item
					continue
				}

				// Select next position to push items
				c := compare(item, selectedItem)
				// fmt.Printf("compared %v to %v => score: %d\n", selectedItem, item, c)
				if c < 0 {
					// Prioritaire
					selectedPos = pos
					selectedItem = item
					// fmt.Printf("selected pos: #%d\n", pos)
				} else if c == 0 {
					// Same score keep last position in priority
					if pos == lastPos {
						// Keep last position
						selectedPos = pos
						selectedItem = item
						// fmt.Printf("selected equal pos: #%d\n", pos)
					}
				} else {
					// v > 0
					// // Need to do something ?
				}
			}

			if selectedPos == -1 {
				// Nothing selected
				continue
			}

			// fmt.Printf("selectedItem: %v\n", selectedItem)
			if !yield(selectedItem) {
				break End
			}
			delete(*nextItems, selectedPos)
			lastPos = selectedPos

			mutex.Lock()
			c, ok := (*chans)[selectedPos]
			mutex.Unlock()
			if ok {
				// Consume next item of selected chan
				e, ok := <-c
				if ok {
					(*nextItems)[selectedPos] = e
				} else {
					mutex.Lock()
					delete(*chans, selectedPos)
					mutex.Unlock()
				}
			}
		}
	}

	return cat
}

type Entry[K, V any] struct {
	K K
	V V
}

func Merge2[K, V any](compare func(ka, kb K, va, vb V) int, its ...iter.Seq2[K, V]) iter.Seq2[K, V] {
	stop := false
	chansVal := make(map[int]chan Entry[K, V])
	chans := &chansVal
	nextItemsVal := make(map[int]Entry[K, V])
	nextItems := &nextItemsVal
	mutex := &sync.Mutex{}

	init := func() {
		stop = false
		// m := &sync.Mutex{}

		// Launch goroutines to stream produced datas
		for pos, it := range its {
			// Make a buffered chan of size 1
			c := make(chan Entry[K, V], 1)
			mutex.Lock()
			(*chans)[pos] = c
			mutex.Unlock()
			go func() {
				// Continuously push next data
				defer close(c)
				for k, v := range it {
					c <- Entry[K, V]{k, v}
					if stop {
						break
					}
				}
			}()
		}

		// Init consume first item of each chan
		for pos, c := range *chans {
			e, ok := <-c
			if ok {
				(*nextItems)[pos] = e
			} else {
				mutex.Lock()
				delete(*chans, pos)
				mutex.Unlock()
			}
		}

		// fmt.Printf("cat paginer initialized chans: %d\n", len(*chans))
	}

	init()

	cat := func(yield func(K, V) bool) {
		defer func() {
			stop = true
		}()
		var positions []int
		// var lastItem T
		lastPos := -1
		looped := false
	End:
		for !looped || len(positions) > 0 {
			looped = true
			positions = collectionz.Keys(*nextItems)
			sort.Ints(positions)
			selectedPos := -1
			var selectedItem Entry[K, V]
			// fmt.Printf("nextItems: %v ; selectedPos: %d ; lastPos: %d\n", *nextItems, selectedPos, lastPos)
			for _, pos := range positions {
				item, ok := (*nextItems)[pos]
				if !ok {
					// no more entries in chan
					delete(*nextItems, pos)
					selectedPos = -1
					// fmt.Printf("chan %d closed\n", pos)
					continue
				}

				if selectedPos == -1 {
					// Select first by default
					// c := compare(lastItem, item)
					selectedPos = pos
					selectedItem = item
					continue
				}

				// Select next position to push items
				c := compare(item.K, selectedItem.K, item.V, selectedItem.V)
				// fmt.Printf("compared %v to %v => score: %d\n", selectedItem, item, c)
				if c < 0 {
					// Prioritaire
					selectedPos = pos
					selectedItem = item
					// fmt.Printf("selected pos: #%d\n", pos)
				} else if c == 0 {
					// Same score keep last position in priority
					if pos == lastPos {
						// Keep last position
						selectedPos = pos
						selectedItem = item
						// fmt.Printf("selected equal pos: #%d\n", pos)
					}
				} else {
					// v > 0
					// // Need to do something ?
				}
			}

			if selectedPos == -1 {
				// Nothing selected
				continue
			}

			// fmt.Printf("selectedItem: %v\n", selectedItem)
			if !yield(selectedItem.K, selectedItem.V) {
				break End
			}
			delete(*nextItems, selectedPos)
			lastPos = selectedPos

			mutex.Lock()
			c, ok := (*chans)[selectedPos]
			mutex.Unlock()
			if ok {
				// Consume next item of selected chan
				e, ok := <-c
				if ok {
					(*nextItems)[selectedPos] = e
				} else {
					mutex.Lock()
					delete(*chans, selectedPos)
					mutex.Unlock()
				}
			}
		}
	}

	return cat
}
