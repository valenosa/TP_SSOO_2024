package main

import (
	"encoding/json"
	"net/http"
)

type Proceso struct {
	pID int `json:"pid"`
	//state string //READY, EXEC, BLOCK, EXIT
}

func main() {
	http.HandleFunc("PUT /process", iniciar_proceso)
	http.HandleFunc("DELETE /process/{pid}", finalizar_proceso)
	http.ListenAndServe(":8080", nil)
}

func iniciar_proceso(w http.ResponseWriter, r *http.Request) {

	p := Proceso{pID: 0}

	res, err := json.Marshal(p)

	if err != nil {
		http.Error(w, "fallo", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func finalizar_proceso(w http.ResponseWriter, r *http.Request) {
	res, err := json.Marshal("")

	if err != nil {
		http.Error(w, "fallo", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}
