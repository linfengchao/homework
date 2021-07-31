package http

type IHttp interface {
	Start()
	ShutDown()
	GetDieChan() chan struct{}
}
