package main

import (
	"time"
	MyHttp "week03/http"
)

func main(){
	s:=MyHttp.NewHttpServer()
	time.AfterFunc(time.Second*10, func() {
		s.GetDieChan()<- struct{}{}
	})
	defer s.ShutDown()
	s.Start()
}