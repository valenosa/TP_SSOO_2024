package planificacion

import (
	"encoding/json"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/utils/config"
)

//================================| CLIENT SIDE |===================================================\\

//-------------------------- STRUCTS ------------------------------------

//-------------------------- CALLS --------------------------------------

func Iniciar(configJson config.Kernel) {
	// Enviar request al servidor
	respuesta := config.Request(configJson.Port_CPU, configJson.Ip_CPU, "PUT", "plani")
	// Verificar que no hubo error en la request
	if respuesta == nil {
		return
	}

}

func Detener(configJson config.Kernel) {
	// Enviar request al servidor
	respuesta := config.Request(configJson.Port_CPU, configJson.Ip_CPU, "DELETE", "plani")
	// Verificar que no hubo error en la request
	if respuesta == nil {
		return
	}

}

//================================| SERVER SIDE |===================================================\\

//-------------------------- STRUCTS ------------------------------------

//-------------------------- HANDLERS -----------------------------------

func HandlerIniciar(w http.ResponseWriter, r *http.Request) {

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

func HandlerDetener(w http.ResponseWriter, r *http.Request) {

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
