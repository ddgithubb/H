package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"

	"go_tests/MutexMap"

	"github.com/dustinxie/lockfree"
	ants "github.com/panjf2000/ants/v2"
	cmap "github.com/streamrail/concurrent-map"
)

const NUM_WORKERS = 10000
const CORES = 0

func BenchmarkLoadSyncMap(b *testing.B) {
	runtime.GOMAXPROCS(CORES)
	var sm sync.Map

	var wg sync.WaitGroup
	wg.Add(NUM_WORKERS)

	for user := 0; user < NUM_WORKERS; user++ {
		go func() {
			for i := 0; i < b.N; i++ {
				//go func() {
				testChannel := make(chan string)
				sm.Store("TEST", testChannel)
				//}()
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkLoadConcurrentMap(b *testing.B) {
	runtime.GOMAXPROCS(CORES)
	sm := cmap.New()

	var wg sync.WaitGroup
	wg.Add(NUM_WORKERS)

	for user := 0; user < NUM_WORKERS; user++ {
		go func() {
			for i := 0; i < b.N; i++ {
				//go func() {
				testChannel := make(chan string)
				sm.Set("TEST", testChannel)
				//}()
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkLoadCustomMap(b *testing.B) {
	runtime.GOMAXPROCS(CORES)
	sm := MutexMap.NewMutexStringChannelMap()

	var wg sync.WaitGroup
	wg.Add(NUM_WORKERS)

	for user := 0; user < NUM_WORKERS; user++ {
		go func() {
			for i := 0; i < b.N; i++ {
				//go func() {
				testChannel := make(chan string)
				sm.Set("TEST", testChannel)
				//}()
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkLoadLockFreeMap(b *testing.B) {
	runtime.GOMAXPROCS(CORES)
	sm := lockfree.NewHashMap()

	var wg sync.WaitGroup
	wg.Add(NUM_WORKERS)

	for user := 0; user < NUM_WORKERS; user++ {
		go func() {
			for i := 0; i < b.N; i++ {
				//go func() {
				testChannel := make(chan string)
				sm.Set("TEST", testChannel)
				//}()
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkReadSyncMap(b *testing.B) {
	runtime.GOMAXPROCS(CORES)
	var sm sync.Map

	testChannel := make(chan string)
	sm.Store(69, testChannel)

	go func() {
		for {
			<-testChannel
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		//go func() {
		loadedChan, ok := sm.Load(69)
		if !ok {
			return
		}

		loadedChan.(chan string) <- "OK"
		//}()
	}
}

func BenchmarkReadConcurrentMap(b *testing.B) {
	runtime.GOMAXPROCS(CORES)
	sm := cmap.New()

	testChannel := make(chan string)
	sm.Set("TESTCHAN", testChannel)

	go func() {
		for {
			<-testChannel
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		//go func() {
		loadedChan, ok := sm.Get("TESTCHAN")
		if !ok {
			return
		}

		loadedChan.(chan string) <- "OK"
		//}()
	}
}

func BenchmarkReadCustomSyncMap(b *testing.B) {
	runtime.GOMAXPROCS(CORES)
	sm := MutexMap.NewMutexStringChannelMap()

	testChannel := make(chan string)
	sm.Set("TESTCHAN", testChannel)

	go func() {
		for {
			<-testChannel
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		//go func() {
		loadedChan, ok := sm.Get("TESTCHAN")
		if !ok {
			return
		}

		loadedChan <- "OK"
		//}()
	}
}

func BenchmarkReadLockFreeMap(b *testing.B) {
	runtime.GOMAXPROCS(CORES)
	sm := lockfree.NewHashMap()

	testChannel := make(chan string)
	sm.Set("TESTCHAN", testChannel)

	go func() {
		for {
			<-testChannel
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		//go func() {
		loadedChan, ok := sm.Get("TESTCHAN")
		if !ok {
			return
		}

		loadedChan.(chan string) <- "OK"
		//}()
	}
}

func BenchmarkMixedSyncMap(b *testing.B) {
	runtime.GOMAXPROCS(CORES)
	var sm sync.Map

	go func() {
		for i := 0; i < b.N; i++ {
			//go func() {
			testChannel := make(chan string)
			sm.Store(fmt.Sprint(i), testChannel)
			//}()
		}
	}()

	for i := 0; i < b.N; i++ {
		//go func() {
		_, ok := sm.Load(fmt.Sprint(i))
		if ok {

		}
		//}()
	}
}

func BenchmarkMixedConcurrentMap(b *testing.B) {
	runtime.GOMAXPROCS(CORES)
	sm := cmap.New()

	testChannel := make(chan string)
	sm.Set("TESTCHAN", testChannel)

	go func() {
		for {
			<-testChannel
		}
	}()

	go func() {
		for i := 0; i < b.N; i++ {
			//go func() {
			tc := make(chan string)
			sm.Set(fmt.Sprint(i), tc)
			//}()
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		//go func() {
		_, _ = sm.Get(fmt.Sprint(i))
		//testChannel<-"OK"
		//}()
	}
}

//To test with BenchmarkMixedConcurrentMap
//SHOWS that unless your oeration is at least 30 micro seconds, there is no reason to use a worker pool.
func BenchmarkWorkerPoolMixedConcurrentMap(b *testing.B) {
	runtime.GOMAXPROCS(CORES)

	var wp routerPool
	wp.Initialize()

	go func() {
		for i := 0; i < b.N; i++ {
			//go func() {
			testChannel := make(chan string)
			wp.connections.Set(fmt.Sprint(i), testChannel)
			//}()
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wp.router <- fmt.Sprint(i)
	}

	// for i := 0; i < b.N; i++ {
	// 	<- wp.test_chan
	// }
}

func BenchmarkMixedCustomSyncMap(b *testing.B) {
	runtime.GOMAXPROCS(CORES)
	sm := MutexMap.NewMutexStringChannelMap()

	go func() {
		for i := 0; i < b.N; i++ {
			//go func() {
			testChannel := make(chan string)
			sm.Set(fmt.Sprint(i), testChannel)
			//}()
		}
	}()

	for i := 0; i < b.N; i++ {
		//go func() {
		_, ok := sm.Get(fmt.Sprint(i))
		if ok {

		}
		//}()
	}
}

func BenchmarkMixedLockFreeMap(b *testing.B) {
	runtime.GOMAXPROCS(CORES)
	sm := lockfree.NewHashMap()

	go func() {
		for i := 0; i < b.N; i++ {
			//go func() {
			testChannel := make(chan string)
			sm.Set(fmt.Sprint(i), testChannel)
			//}()
		}
	}()

	for i := 0; i < b.N; i++ {
		//go func() {
		_, ok := sm.Get(fmt.Sprint(i))
		if ok {

		}
		//}()
	}
}

func BenchmarkWriteMultiCores(b *testing.B) {
	runtime.GOMAXPROCS(CORES)
	sm := cmap.New()
	//sm := lockfree.NewHashMap()

	const THREADS = 12

	var wg sync.WaitGroup
	wg.Add(THREADS)

	for thread := 0; thread < THREADS; thread++ {
		go func() {
			for i := 0; i < b.N/THREADS; i++ {
				//go func() {
				testChannel := make(chan string)
				sm.Set(fmt.Sprint(i), testChannel)
				//}()
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkReadMultiCores(b *testing.B) {
	runtime.GOMAXPROCS(CORES)
	sm := cmap.New()
	//sm := lockfree.NewHashMap()

	var THREADS = b.N
	var runs = 1

	var wg sync.WaitGroup
	wg.Add(THREADS * runs)

	go func() {
		for i := 0; i < b.N; i++ {
			//go func() {
			testChannel := make(chan string)
			sm.Set(fmt.Sprint(i), testChannel)
			//}()
		}
	}()

	b.ResetTimer()

	for thread := 0; thread < THREADS; thread++ {
		for j := 0; j < runs; j++ {
			go func() {
				for i := 0; i < b.N/THREADS; i++ {
					_, ok := sm.Get(fmt.Sprint(i))
					if ok {

					}
				}
				wg.Done()
			}()
		}

	}

	wg.Wait()
}

func BenchmarkReadPoolMultiCores(b *testing.B) {
	runtime.GOMAXPROCS(CORES)
	sm := cmap.New()

	var THREADS = b.N
	var runs = 1
	var wg sync.WaitGroup

	pool, _ := ants.NewPoolWithFunc(ants.DefaultAntsPoolSize, func(i interface{}) {
		_, ok := sm.Get(fmt.Sprint(i))
		if ok {

		}
		//time.Sleep(time.Microsecond)
		wg.Done()
	})
	wg.Add(THREADS * runs)

	go func() {
		for i := 0; i < b.N; i++ {
			//go func() {
			testChannel := make(chan string)
			sm.Set(fmt.Sprint(i), testChannel)
			//}()
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < runs; j++ {
			pool.Invoke(i)
		}
	}

	wg.Wait()

}

func BenchmarkAtomicUnsafe(b *testing.B) {
	var ptr unsafe.Pointer

	var wg sync.WaitGroup
	wg.Add(b.N)

	runtime.GOMAXPROCS(0)

	go func() {
		for i := 0; i < b.N; i++ {
			//go func() {
			testChannel := make(chan string)
			atomic.StorePointer(&ptr, unsafe.Pointer(&testChannel))
			//}()
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		//go func() {
		t := (*chan string)(atomic.LoadPointer(&ptr))
		_ = t
		//	//fmt.Println(tc)
		//	wg.Done()
		//}()
	}

	//wg.Wait()
}

func BenchmarkWriteLatencyPerCore(b *testing.B) {
	newCores := 16
	runtime.GOMAXPROCS(newCores)
	sm := cmap.New()

	THREADS := 100

	var wg sync.WaitGroup
	wg.Add(THREADS)

	for thread := 0; thread < THREADS; thread++ {
		go func() {
			for i := 0; i < b.N; i++ {
				testChannel := make(chan string)
				sm.Set(fmt.Sprint(i), testChannel)
			}
		}()
	}

	b.ResetTimer()

	for thread := 0; thread < THREADS; thread++ {
		go func() {
			for i := 0; i < b.N/THREADS; i++ {
				_, ok := sm.Get(fmt.Sprint(i))
				if ok {

				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkAppend(b *testing.B) {
	ary := make([]int, 0, b.N)

	for i := 0; i < b.N; i++ {
		ary = append(ary, i)
	}
}

func BenchmarkAppendSlices(b *testing.B) {
	ary := make([]int, 1000)
	prealloc := make([]int, 1000)

	for i := 0; i < b.N; i++ {
		prealloc = append(prealloc, ary...)
	}
	_ = prealloc
}

func BenchmarkAppendSlicesByCopy(b *testing.B) {
	ary := make([]int, 1000)
	prealloc := make([]int, 2000)
	n := 0

	for i := 0; i < b.N; i++ {
		prealloc[0] = 1
		n = copy(prealloc[1000:], ary)
	}
	_ = n
	//fmt.Println(n, len(prealloc), prealloc)
}

func BenchmarkAppendSlicesByAssign(b *testing.B) {
	ary := make([]int, 1000)
	prealloc := make([]int, 2000)
	n := 0

	for i := 0; i < b.N; i++ {
		prealloc[0] = 1
		for j := 0; j < len(ary); j++ {
			prealloc[1000+j] = ary[j]
		}
	}
	_ = n
	//fmt.Println(n, len(prealloc), prealloc)
}

func BenchmarkPrepend(b *testing.B) {
	ary := make([]int, b.N, b.N+1)

	for i := 0; i < b.N; i++ {
		ary = append([]int{i}, ary...)
	}
}

func BenchmarkPrependWithReverse(b *testing.B) {
	ary := make([]int, b.N)

	for i := 0; i < b.N; i++ {
		ary = append(ary, i)
	}

	for i := 0; i < len(ary)/2; i++ {
		j := len(ary) - i - 1
		ary[i], ary[j] = ary[j], ary[i]
	}
}

func BenchmarkSpread(b *testing.B) {
	ary := make([]int, b.N)

	for i := 0; i < b.N; i++ {
		func(s ...int) {

		}(ary...)
	}
}

func BenchmarkNormalSlice(b *testing.B) {
	ary := make([]int, b.N)

	for i := 0; i < b.N; i++ {
		func(s []int) {

		}(ary)
	}
}

func BenchmarkNormalMap(b *testing.B) {
	m := make(map[string]string)
	m["0"] = "0"

	for i := 0; i < b.N; i++ {
		res, ok := m["0"]
		if res == "0" && ok {
			//fmt.Sprint(res)
		}
	}
}

func BenchmarkReaderReadSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := strings.NewReader("abcdefaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa|ghij")
		r := bufio.NewReader(s)
		_, err := r.ReadSlice('|')
		if err != nil {
			panic(err)
		}
		//fmt.Printf("Token: %q\n", token)
	}
}

func BenchmarkReaderRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := strings.NewReader("abcdefaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa|ghij")
		r := bufio.NewReader(s)
		buf := make([]byte, 48)
		_, err := r.Read(buf)
		if err != nil {
			panic(err)
		}
		//fmt.Printf("Token: %q\n", token)
	}
}

func BenchmarkByteToStringAndBack(b *testing.B) {
	buf := make([]byte, 100)
	var str string = "1000000000000000000000000000000000000000000000000000000000000"

	for i := 0; i < b.N; i++ {
		str = string(buf[:])
		//_ = str
		buf = []byte(str)
		//_ = buf
	}
}

func BenchmarkCopyByteArray(b *testing.B) {
	buf := make([]byte, 1000)
	var buf_copy []byte

	for i := 0; i < b.N; i++ {
		buf_copy = buf
	}

	_ = buf_copy
}

func BenchmarkMapAccess(b *testing.B) {
	m := make(map[string]string)
	m["0"] = "A"
	m["1"] = "B"
	m["2"] = "C"
	m["3"] = "D"
	m["4"] = "E"
	m["5"] = "F"
	m["6"] = "G"
	m["7"] = "H"
	m["8"] = "I"
	m["9"] = "J"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = m["9"]
	}
}

func BenchmarkSliceAccess(b *testing.B) {
	m := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}

	for i := 0; i < b.N; i++ {
		for j := 0; j < len(m); j++ {
			if j == 9 {
				_ = m[9]
			}
		}
	}
}

func BenchmarkSliceRangeAccess(b *testing.B) {
	m := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}

	for i := 0; i < b.N; i++ {
		for j, _ := range m {
			if j == 9 {
				_ = m[j]
			}
		}
	}
}

func BenchmarkIntDataRace(b *testing.B) {

	n := 10000

	var wg sync.WaitGroup
	wg.Add(n)
	value := new(uint64)
	atomic.StoreUint64(value, 0)

	for i := 1; i <= n; i++ {

		go func(c int) {
			atomic.AddUint64(value, uint64(c))
			wg.Done()
		}(i)
	}

	wg.Wait()

	end_val := atomic.LoadUint64(value)
	fmt.Println("END_VAL", end_val)
}

func BenchmarkSliceDataRace(b *testing.B) {

	n := 10000

	var wg sync.WaitGroup
	wg.Add(n)
	value := new(uint64)
	atomic.StoreUint64(value, 0)

	var s []int
	for i := 1; i <= n; i++ {
		s = nil
		s = make([]int, 1)
		s[0] = i
		func(c []int) {
			c = nil
		}(s)
		go func(c []int) {
			atomic.AddUint64(value, uint64(c[0]))
			wg.Done()
		}(s)
	}

	wg.Wait()

	end_val := atomic.LoadUint64(value)
	fmt.Println("END_VAL", end_val)
}

func BenchmarkAppendToFile(b *testing.B) {
	file, _ := os.OpenFile("temp.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	w := bufio.NewWriter(file)
	defer file.Close()
	b.ResetTimer()
	//w.WriteString("String")
	for i := 0; i < b.N; i++ {
		w.WriteString("String:" + fmt.Sprint(i) + "\n")
	}
	w.Flush()
	os.Truncate("temp.txt", 0)
}

func BenchmarkTruncateFile(b *testing.B) {
	file, _ := os.OpenFile("temp.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	defer file.Close()
	b.ResetTimer()
	//w.WriteString("String")
	for i := 0; i < b.N; i++ {
		os.Truncate("temp.txt", 0)
	}
}

func BenchmarkDataRaceObjectAndValue(b *testing.B) {

	type test struct {
		value1 int
		value2 int
	}

	t := &test{
		value1: 1,
		value2: 2,
	}
	go func() {
		fmt.Println("START1")
		for i := 0; i < b.N; i++ {
			t.value1++ // read, increment, write
		}
		fmt.Println("DONE1")
	}()
	fmt.Println("START")
	for i := 0; i < b.N; i++ {
		t.value1++ // read, increment, write
	}
	fmt.Println("DONE")

	fmt.Println(t) // Output: <unspecified>
}
