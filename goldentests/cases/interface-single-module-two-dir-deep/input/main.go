package main

import "interface-single-module/src/module1"

func main() {
	s := module1.NewSomethingThatUses(&module1.ConcreteClass{})
	s.Run()
}
