package main

import (
	"log"
	"net/http"

  "github.com/BrandonSharratt/starter-snake-go/api"
)

func Start(res http.ResponseWriter, req *http.Request) {
	decoded := api.SnakeRequest{}
	err := api.DecodeSnakeRequest(req, &decoded)
	if err != nil {
		log.Printf("Bad start request: %v", err)
	}
	//dump(decoded)

	respond(res, api.StartResponse{
		Color: "#006400",
	})
}

func Move(res http.ResponseWriter, req *http.Request) {
	decoded := api.SnakeRequest{}
	err := api.DecodeSnakeRequest(req, &decoded)
	if err != nil {
		log.Printf("Bad move request: %v", err)
	}
	//dump(decoded)

	respond(res, api.MoveResponse{
		Move: api.GetMove(decoded),
	})
}

func End(res http.ResponseWriter, req *http.Request) {
	println("I died")
	println("Previous estimates - down:", api.PreviousEstimates["down"], " up: ", api.PreviousEstimates["up"], " left: ", api.PreviousEstimates["left"], " right: ", api.PreviousEstimates["right"])
	return
}

func Ping(res http.ResponseWriter, req *http.Request) {
	respond(res, "Hello World")
}
