package module1

import (
	"log"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func NewActor(size int) (*producer, *consumer) {
	ch := make(chan int, size)
	producer := &producer{ch: ch, stop: make(chan struct{})}
	consumer := &consumer{ch: ch, stop: make(chan struct{})}
	return producer, consumer
}

type producer struct {
	ch   chan<- int
	stop chan struct{}
}

func (p *producer) Start(gap time.Duration) {
	for {
		select {
		case <-p.stop:
			return
		default:
			v := rand.Intn(100)
			log.Printf("Producer sent element: %d", v)
			p.ch <- v
			time.Sleep(gap)
		}
	}
}

func (p *producer) Stop() {
	p.stop <- struct{}{}
	close(p.ch)
	log.Println("Producer Terminated")
}

type consumer struct {
	ch   <-chan int
	stop chan struct{}
}

func (p *consumer) Start(gap time.Duration) {
	for {
		select {
		case v, ok := <-p.ch:
			if !ok {
				continue
			}
			log.Printf("Consumer got element: %d", v)
			time.Sleep(gap)
		case <-p.stop:
			return
		}
	}
}

func (p *consumer) Stop() {
	p.stop <- struct{}{}
	log.Println("Consumer Terminated")
}
