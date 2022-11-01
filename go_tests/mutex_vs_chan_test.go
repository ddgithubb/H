package main

import (
	"fmt"
	"sync"
	"testing"
)

// Reason why map is one sharded is because channels can in theory be sharded as well!

const reader_over_writer_fraction = 1000

type table struct {
	table map[int]string
	sync.RWMutex
}

func BenchmarkMutex(b *testing.B) {

	var wg sync.WaitGroup

	m := &table{
		table: make(map[int]string),
	}

	wg.Add(b.N)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if i % reader_over_writer_fraction == 0 {
			go func() {
				m.Lock()
				m.table[i] = "write"
				m.Unlock()
				wg.Done()
			}()
		} else {
			go func() {
				m.RLock()
				_ = m.table[i]
				m.RUnlock()
				wg.Done()
			}()
		}
	}

	wg.Wait()
}

func BenchmarkChannel(b *testing.B) {
	write_chan := make(chan int, 1)
	send_chan := make(chan int, 1)

	go func() {

		m := make(map[int]string)

		var wi int
		var pi int

		for {
			select {
			case wi = <-write_chan:
				m[wi] = "write"
			case pi = <-send_chan:
				res := m[pi]
				go func(res string) {
					_ = res
				}(res)
			}
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if i % reader_over_writer_fraction == 0 {
			write_chan<-i
		} else {
			send_chan<-i
		}
	}
}

// Effectively a write ONLY lock, no read locks
// Cons are that there must only be 1 in sync reader
// CHANNELS ARE LOCKS THEMSELVES!
func BenchmarkReadWithChanWrite(b *testing.B) {
	write_chan := make(chan int, 1)
	m := make(map[int]string)

	write_count := 0
	read_count := 0

	go func() {
		for i := 0; i < b.N/reader_over_writer_fraction; i++ {
			write_chan<-i
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// if i % reader_over_writer_fraction == 0 {
		// 	write_chan<-i
		// }
		select {
		case recv := <-write_chan:
			write_count++
			//fmt.Println(i)
			m[recv] = "write"
		default:
			read_count++
			res := m[i]
			go func(res string) {
				_ = res
			}(res)
		}
	}

	fmt.Println("COUNT:", read_count, write_count)
}


//////////////////////////////////////////////////////////////////
/////////////////////// READ IN SYNC /////////////////////////////
//////////////////////////////////////////////////////////////////

func BenchmarkSyncMutex(b *testing.B) {

	m := &table{
		table: make(map[int]string),
	}

	go func() {
		for i := 0; i < b.N/reader_over_writer_fraction; i++ {
			m.Lock()
			m.table[i] = "write"
			m.Unlock()
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.RLock()
		_ = m.table[i]
		m.RUnlock()
	}
}

// Effectively a write ONLY lock, no read locks
// Cons are that there must only be 1 in sync reader
// CHANNELS ARE LOCKS THEMSELVES!
func BenchmarkSyncReadWithChanWrite(b *testing.B) {
	write_chan := make(chan int, 1)
	m := make(map[int]string)

	write_count := 0
	read_count := 0

	go func() {
		for i := 0; i < b.N/reader_over_writer_fraction; i++ {
			write_chan<-i
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// if i % reader_over_writer_fraction == 0 {
		// 	write_chan<-i
		// }
		select {
		case recv := <-write_chan:
			write_count++
			//fmt.Println(i)
			m[recv] = "write"
		default:
			read_count++
			_ = m[i]
		}
	}

	fmt.Println("COUNT:", read_count, write_count)
}