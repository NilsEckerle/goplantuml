package module1

type SomethingThatUses struct {
	someClass ISomeClass
}

func NewSomethingThatUses(someClass ISomeClass) SomethingThatUses {
	return SomethingThatUses{
		someClass: someClass,
	}
}

func (s SomethingThatUses) Run() {
	s.someClass.Foo()
	k := s.someClass.Bar(5)
	k--
}
