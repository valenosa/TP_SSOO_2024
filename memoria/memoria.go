package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Declaración temporal. Próximamente las estructuras compartidas se encontrarán unificadas en un archivo
type BodyIniciarProceso struct {
	// Path del archivo que se utilizará como base para ejecutar un nuevo proceso
	Path string `json:"path"`
}

// Declaración temporal. Próximamente las estructuras compartidas se encontrarán unificadas en un archivo
// La cambio porque se necesita para varias respuestas, no solo para iniciar proceso
type ResponseProceso struct {
	Pid    int    `json:"pid"`
	Estado string `json:"estado"`
}

func main() {
	// Declaración temporal.

	// Se establece el handler que se utilizará para las diversas situaciones recibidas por el server
	http.HandleFunc("PUT /process", handler_iniciar_proceso)
	http.HandleFunc("DELETE /process/{pid}", handler_finalizar_proceso)
	http.HandleFunc("GET /process/{pid}", handler_estado_proceso)
	http.ListenAndServe(":8080", nil)
}

func handler_iniciar_proceso(w http.ResponseWriter, r *http.Request) {

	//Crea uan variable tipo BodyIniciarProceso (para interpretar lo que se recibe de la request)
	var request BodyIniciarProceso

	// Decodifica el request (codificado en formato json)
	err := json.NewDecoder(r.Body).Decode(&request)

	// Error Handler de la decodificación
	if err != nil {
		fmt.Printf("Error al decodificar request body: ")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Imprime el request por consola (del lado del server)
	fmt.Printf("Request path: %s\n", request)

	//Crea una variable tipo ResponseIniciarProceso (para confeccionar una respuesta)
	var respBody ResponseProceso = ResponseProceso{Pid: 0}

	// Codificar Response en un array de bytes (formato json)
	respuesta, err := json.Marshal(respBody)

	// Error Handler de la codificación
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

// primera versión de finalizar proceso, com no recive ni devuelve solo le mandamos status ok.
// Cuando haya  procesos se busca por el %s del path {pid}
func handler_finalizar_proceso(w http.ResponseWriter, r *http.Request) {
	res, err := json.Marshal("")

	if err != nil {
		http.Error(w, "fallo", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

// primera versión de estado proceso, como es GET no necesita recibir nada
// Cuando haya  procesos se busca por el %s del path {pid}
func handler_estado_proceso(w http.ResponseWriter, r *http.Request) {
	//usando el struct de ResponseProceso envío el estado del proceso
	var respBody ResponseProceso = ResponseProceso{Estado: "ready"}

	respuesta, err := json.Marshal(respBody)

	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}
	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}
