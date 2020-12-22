package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const N = 50

type User struct {
	id    int
	calls chan<- *User
	taxi  chan *Taxi
	wg    *sync.WaitGroup
}

type Taxi struct {
	id         int
	listen     <-chan *Dispatcher
	driveCheck chan *User
}

type Dispatcher struct {
	calls    <-chan *User
	taxiChan chan<- *Dispatcher
	listen   chan *Taxi
}

func (user *User) callDispatcher() {
	n := rand.Intn(100)
	time.Sleep(time.Duration(n) * time.Millisecond)
	user.calls <- user
	taxi := <-user.taxi
	if taxi != nil {
		taxi.driveCheck <- user
	} else {
		user.wg.Done()
	}
}

func (taxi *Taxi) drive(user *User) {
	defer user.wg.Done()
	n := rand.Intn(50) + 100
	time.Sleep(time.Duration(n) * time.Millisecond)
	fmt.Printf("Taxi %4d done with user %d\n", taxi.id, user.id)
}

func (taxi *Taxi) wait() {
	for {
		n := rand.Intn(N)
		select {
		case dispatcher := <-taxi.listen:
			dispatcher.listen <- taxi
		case user := <-taxi.driveCheck:
			taxi.drive(user)
		default:
			time.Sleep(time.Duration(n) * time.Millisecond)
		}
	}
}

func (disp *Dispatcher) dispatch() {
	for {
		user := <-disp.calls
		disp.taxiChan <- disp
		timeout := time.After(500 * time.Millisecond)
		select {
		case taxi := <-disp.listen:
			user.taxi <- taxi
		case <-timeout:
			user.taxi <- nil
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	const NTaxi = 500
	const NUsers = NTaxi * 10

	taxiChan := make(chan *Dispatcher)
	calls := make(chan *User)

	for i := 0; i < NTaxi; i++ {
		t := &Taxi{
			id:         i,
			listen:     taxiChan,
			driveCheck: make(chan *User),
		}
		go t.wait()
	}

	disp := Dispatcher{
		calls:    calls,
		taxiChan: taxiChan,
		listen:   make(chan *Taxi),
	}
	go disp.dispatch()

	var wg sync.WaitGroup
	wg.Add(NUsers)

	for i := 0; i < NUsers; i++ {
		user := &User{
			id:    i,
			calls: calls,
			taxi:  make(chan *Taxi),
			wg:    &wg,
		}
		go user.callDispatcher()
	}

	wg.Wait()
}
