package limiter

type Limiter interface {
	Pass() error
}
