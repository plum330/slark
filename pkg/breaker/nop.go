package breaker

type Noop struct{}

func (n *Noop) Allow() error {
	return nil
}

func (n *Noop) Fail() {

}

func (n *Noop) Succeed() {

}
