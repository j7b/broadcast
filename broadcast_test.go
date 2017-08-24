package broadcast

import (
	"fmt"
	"sync"
	"testing"
)

type number int

func (number) Broadcast() {}

func TestBroadcast(t *testing.T) {
	fanin := make(chan int)
	s := New()
	receive := func(r Receiver, c chan int) {
		b := r.Receive()
		c <- int(b.(number))
	}
	go receive(s.Receiver(), fanin)
	go receive(s.Receiver(), fanin)
	s.Send(number(2))
	if i := <-fanin; i != 2 {
		t.Fatalf("want 2 got %v", i)
	}
	if i := <-fanin; i != 2 {
		t.Fatalf("want 2 got %v", i)
	}
}

func TestSharedSender(t *testing.T) {

	fanin := make(chan int)
	f := func(r Receiver, c chan int) {
		b := r.Receive()
		c <- int(b.(number))
		b = r.Receive()
		c <- int(b.(number))
		b = r.Receive()
		c <- int(b.(number))
		b = r.Receive()
		c <- int(b.(number))
	}
	ss := NewShared()
	go f(ss.Receiver(), fanin)
	go f(ss.Receiver(), fanin)
	go func() {
		ss.Send(number(2))
		ss.Send(number(1))
	}()
	go func() {
		ss.Send(number(1))
		ss.Send(number(2))
	}()
	ints := make([]int, 8)
	for i := range ints {
		n := <-fanin
		switch n {
		case 1, 2:
		default:
			t.Fatalf("want 1 or 2 got %v", n)
		}
		ints[i] = n
	}
}

func Example() {
	wg := new(sync.WaitGroup)
	s := New()
	f := func(r Receiver) {
		defer wg.Done()
		counter := 0
		for {
			b := r.Receive()
			n, ok := b.(number)
			if !ok {
				panic("not a number")
			}
			if int(n) != counter {
				panic("number != counter")
			}
			counter++
			if counter == 10 {
				return
			}
		}
	}
	wg.Add(8000)
	for i := 0; i < 8000; i++ {
		go f(s.Receiver())
	}
	for i := 0; i < 10; i++ {
		s.Send(number(i))
	}
	wg.Wait()
	fmt.Println("Goroutines done")
	// Output:
	// Goroutines done
}

func BenchmarkBroadcast(b *testing.B) {
	wg := new(sync.WaitGroup)
	s := New()
	for n := 0; n < 8000; n++ {
		wg.Add(1)
		go func(r Receiver) {
			defer wg.Done()
			for {
				b := r.Receive()
				if b == nil {
					return
				}
			}
		}(s.Receiver())
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s.Send(number(n))
	}
	s.Send(nil)
	wg.Wait()
}
