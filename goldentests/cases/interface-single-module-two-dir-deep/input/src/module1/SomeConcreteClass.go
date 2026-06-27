package module1

type ConcreteClass struct {
	x int
}

func (c *ConcreteClass) Foo() {
	c.x++
}

func (c *ConcreteClass) Bar(i int) int {
	return i + c.x
}
