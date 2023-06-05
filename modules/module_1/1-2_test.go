package module1

import (
	"testing"
	"time"
)

func Test1_2(t *testing.T) {
	size := 10
	oneSecGap := time.Second
	producer, consumer := NewActor(size)
	go producer.Start(oneSecGap)
	go consumer.Start(oneSecGap)
	time.Sleep(5 * time.Second)
	producer.Stop()
	consumer.Stop()
}
