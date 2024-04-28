package main

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////IMPORTS//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/utils/config"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////MAIN///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func main() {

	// Se establece el handler que se utilizará para las diversas situaciones recibidas por el server
	http.HandleFunc("PUT /plani", handler_iniciar_planificacion)
	http.HandleFunc("DELETE /plani", handler_detener_planificacion)

	// Extrae info de config.json
	var configJson config.Cpu

	config.Iniciar("config.json", &configJson)

	// declaro puerto
	port := ":" + strconv.Itoa(configJson.Port)

	// Listen and serve con info del config.json
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al esuchar en el puerto " + port)
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////FUNCIONES/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////HANDLERS//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func handler_iniciar_planificacion(w http.ResponseWriter, r *http.Request) {

	// Respuesta vacía significa que manda una respuesta vacía, o que no hay respuesta?
	respuesta, err := json.Marshal("")

	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

func handler_detener_planificacion(w http.ResponseWriter, r *http.Request) {

	// Respuesta vacía significa que manda una respuesta vacía, o que no hay respuesta?
	respuesta, err := json.Marshal("")

	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}
