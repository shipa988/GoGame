package server

import (
	"encoding/json"
	"fmt"
	et "github.com/shipa988/ebitentest"
	"net/http"
	"strconv"
)

func logMiddleware(next http.Handler) http.Handler  {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println(request.URL)
		next.ServeHTTP(writer,request)
	})
}
func panicMiddleware(next http.Handler) http.Handler  {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if err:=recover();err!=nil{
				fmt.Println("recovered")
				http.Error(writer,err.(error).Error(),http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(writer,request)
	})
}
func main ()  {
	et.BigBang(640,480,1)

	mainmux:=http.NewServeMux()
	mainmux.HandleFunc("/world/create", CreateWorld)
	mainmux.HandleFunc("/world/born", BornInWorld)
	mainmux.HandleFunc("/world/die", DieInWorld)
	mainmux.HandleFunc("/worlds", GetWorlds)
	mainmux.HandleFunc("/world/event", EventInWorld)
	mainmux.HandleFunc("/", UniverseLoad)
	logHanlder:=logMiddleware(mainmux)
	gameHandler:=panicMiddleware(logHanlder)
	http.ListenAndServe(":8080",gameHandler)
}

func UniverseLoad(writer http.ResponseWriter, request *http.Request) {

}


func EventInWorld(writer http.ResponseWriter, request *http.Request) {

}

func GetWorlds(writer http.ResponseWriter, request *http.Request) {

}

func DieInWorld(writer http.ResponseWriter, request *http.Request) {

}

func BornInWorld(writer http.ResponseWriter, request *http.Request) {

}



func CreateWorld(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPost:
		err:=request.ParseForm()
		if err!=nil{
			writer.WriteHeader(http.StatusBadRequest)
		}
		mapid,err:= strconv.Atoi(request.FormValue("mapid"))
		if err!=nil{
			writer.WriteHeader(http.StatusBadRequest)
		}
		worldid,err:=et.AddWorld(mapid)
		if err!=nil{
			writer.WriteHeader(http.StatusBadRequest)
		}
		w,err:=et.GetWorld(worldid)
		if err!=nil {
			writer.WriteHeader(http.StatusBadRequest)
		}
		answ,err:=json.Marshal(w)
		if err!=nil {
			writer.WriteHeader(http.StatusBadRequest)
		}
		writer.Write(answ)
	case http.MethodGet:
	default:
		writer.WriteHeader(http.StatusBadRequest)
	}
}
