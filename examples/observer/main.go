package main

import (
	"log"
	"math"

	"github.com/kyosheek/go-patterns/pkg/observer"
)

type obsBehaviour struct {
	name      string
	behaviour string
}

type obsPercent struct {
	name   string
	change float64
}

func (o *obsBehaviour) Update(state, prevState int) {
	if state > prevState {
		o.behaviour = "increase"
	} else if state < prevState {
		o.behaviour = "decrease"
	} else {
		o.behaviour = "steady"
	}

	log.Println(o.name, o.behaviour)
}

func (o *obsPercent) Update(state, prevState int) {
	if state == prevState {
		o.change = 0
	} else if state == 0 {
		if prevState < 0 {
			o.change = 100
		} else if prevState > 0 {
			o.change = -100
		}
	} else if prevState == 0 {
		if state > 0 {
			o.change = 100
		} else {
			o.change = -100
		}
	} else {
		diff := float64(state) / float64(prevState)

		if math.Abs(diff) < 1 {
			if state > 0 && prevState < 0 {
				diff += 1
			} else if state < 0 && prevState > 0 {
				diff -= 1
			}
		}

		if state > 0 && prevState < 0 {
			diff *= -1
			diff += 1
		} else if state < 0 && prevState > 0 {
			diff -= 1
		} else if state > 0 && prevState > 0 {
			if state > prevState {
				diff -= 1
			} else {
				diff = 1 - diff
			}
		} else if state < 0 && prevState < 0 {
			diff -= 1
			diff *= -1
		}

		o.change = math.Round(diff*100*100) / 100
	}

	log.Println(o.name, o.change, "%")
}

func main() {
	subject := observer.NewSubject[int]()
	subject.SetState(4)
	observer1 := &obsBehaviour{
		name:      "Behaviour observer",
		behaviour: "steady",
	}
	observer2 := &obsPercent{
		name:   "Percent observer",
		change: 0,
	}

	subject.Attach(observer1, observer2)
	subject.SetState(-2)
	subject.SetState(-2)
	subject.SetState(4)
	subject.SetState(5)
	subject.SetState(-15)
	subject.SetState(-16)
}
