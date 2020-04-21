package server

import (
	"net/http"
	et "github.com/shipa988/ebitentest"
)

type Game struct {
	
}

func ser()  {
	
}

func main ()  {
	et.BigBang(640,480,1)

	
	http.ListenAndServe("/game",Game)
	mux:=http.NewServeMux()
	mux.HandleFunc("/world", func(writer http.ResponseWriter, request *http.Request) {

	})

}
