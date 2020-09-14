package main

import (
	"encoding/json"
	"fmt"
	_ "job-backend-trainee-assignment/docs"
	"job-backend-trainee-assignment/internal/app"
	"net/http"
)


func Handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	b, _ := json.Marshal(&app.User{Id: 1, Balance: 100.0})
	w.Write(b)
}

func main() {

	fmt.Printf("running main")
	http.ListenAndServe(":9000", http.HandlerFunc(Handler))
}
