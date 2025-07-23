package main

import (
	"github.com/kyosheek/go-patterns/pkg/factory"
)

type car struct {
	kmHTop  uint16
	kmHTime float32
}

func main() {
	carFactory := factory.New[car]()

	// (B7) 1.6 TDI BMT (105 Hp)
	volkswagenPassat := carFactory.Create()
	volkswagenPassat.kmHTime = 12.2
	volkswagenPassat.kmHTop = 195

	// VIII (XV70, facelift 2020) 2.5 (218 Hp) Hybrid e-CVT
	toyotaCamry := carFactory.Create()
	toyotaCamry.kmHTop = 180
	toyotaCamry.kmHTime = 8.3
}
