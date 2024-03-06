package limit

type Limiter interface {
	Pass() error
}
