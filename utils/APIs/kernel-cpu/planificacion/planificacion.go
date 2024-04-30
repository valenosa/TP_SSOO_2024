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

func enviarRequest(metodo string, query string, port int, ip string) {
	// Se declara un nuevo cliente
	cliente := &http.Client{}

	// Se declara la url a utilizar (depende de una ip y un puerto).
	url := fmt.Sprintf("http://%s:%d/%s", ip, port, query)

	// Se crea una request donde se "efectúa" el metodo (PUT / DELETE / GET / POST) hacia url, enviando el Body si lo hay
	req, err := http.NewRequest(metodo, url, nil)

	// Error Handler de la construcción de la request
	if err != nil {
		fmt.Printf("error creando request a ip: %s puerto: %d\n", ip, port)
		return
	}

	// Se establecen los headers
	req.Header.Set("Content-Type", "application/json")

	// Se envía el request al servidor
	respuesta, err := cliente.Do(req)

	// Error handler de la request
	if err != nil {
		fmt.Printf("error enviando request a ip: %s puerto: %d\n", ip, port)
		return
	}

	// Verificar el código de estado de la respuesta del servidor a nuestra request (de no ser OK)
	if respuesta.StatusCode != http.StatusOK {
		fmt.Printf("Status Error: %d\n", respuesta.StatusCode)
		return
	}

	//Todo salió bien
	fmt.Printf("%s %s exitoso \n", metodo, query)
}

func Iniciar(configJson config.Kernel) {
	enviarRequest("PUT", "plani", configJson.Port_CPU, configJson.Ip_CPU)
}

func Detener(configJson config.Kernel) {
	enviarRequest("DELETE", "plani", configJson.Port_CPU, configJson.Ip_CPU)

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
