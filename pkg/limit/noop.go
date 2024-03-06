package limit

type Noop struct{}

func (n *Noop) Pass() error {
	return nil
}
