package observer

import (
	"log"

	"github.com/kyosheek/go-patterns/pkg/observer"
)

type obs struct {
	name string
}

func (o *obs) Update(state int) {
	log.Printf("%s is in state %d", o.name, state)
}

func newObserver(name string) observer.Observer[int] {
	return &obs{name: name}
}

func main() {
	subject := observer.NewSubject[int]()
	observer1 := newObserver("observer 1")
	observer2 := newObserver("observer 2")

	subject.Attach(observer1, observer2)
	subject.SetState(2)
}
