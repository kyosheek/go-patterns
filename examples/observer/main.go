package observer

import (
	"log"

	"github.com/kyosheek/go-patterns/pkg/observer"
)

type obs struct {
	name  string
	state int
}

func (o *obs) Update(state int) {
	o.state = state
	log.Printf("%s is in state %d", o.name, o.state)
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
