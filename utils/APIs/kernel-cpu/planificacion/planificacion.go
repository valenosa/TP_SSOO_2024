package planificacion

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/utils/config"
)

//CLIENT SIDE/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

////STRUCTS/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

////CALLS/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func Iniciar(configJson config.Kernel) {

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/plani", configJson.Ip_CPU, configJson.Port_CPU)

	// Genera una petición HTTP.
	req, err := http.NewRequest("PUT", url, nil)

	// Check error generando una request.
	if err != nil {
		fmt.Printf("Error creando request: %s\n", err.Error())
		return
	}

	// Se envía la request al servidor.
	respuesta, err := cliente.Do(req)

	// Check request enviada.
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", configJson.Ip_CPU, configJson.Port_CPU)
		return
	}

	//Espera a que la respuesta se termine de utilizar para liberarla de memoria.
	defer respuesta.Body.Close()

	// Check response recibida.
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}

	// Todo salió bien, la planificación se inició correctamente.
	fmt.Println("Planificación iniciada exitosamente.")
}

func Detener(configJson config.Kernel) {

	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/plani", configJson.Ip_CPU, configJson.Port_CPU)

	// Genera una petición HTTP.
	req, err := http.NewRequest("DELETE", url, nil)

	// Check error generando una request.
	if err != nil {
		fmt.Printf("Error creando request: %s\n", err.Error())
		return
	}

	// Se envía la request al servidor.
	respuesta, err := cliente.Do(req)

	// Check request enviada.
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", configJson.Ip_CPU, configJson.Port_CPU)
		return
	}

	//Espera a que la respuesta se termine de utilizar para liberarla de memoria.
	defer respuesta.Body.Close()

	// Check response recibida.
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}

	// Todo salió bien, la planificación se detuvo correctamente.
	fmt.Println("Planificación detenida exitosamente.")
}

//SERVER SIDE/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

////STRUCTS/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

////HANDLERS/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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
