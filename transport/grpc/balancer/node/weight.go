package node

type PlainWeight struct {
	Node
}

func (p *PlainWeight) Weight() int64 {
	weight := p.InitialWeight()
	if weight != nil {
		return *weight
	}
	return 100 // default weight
}

func (p *PlainWeight) Unwrap() Node {
	return p.Node
}

type WeightedBuilder interface {
	Build(Node) WeightedNode
}

type Plain struct{}

func (p *Plain) Build(node Node) WeightedNode {
	return &PlainWeight{
		Node: node,
	}
}
