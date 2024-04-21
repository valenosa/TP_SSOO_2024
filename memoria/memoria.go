package main

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////IMPORTS//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////STRUCTS//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type MemoriaConfig struct {
	Port              int    `json:"port"`
	Memory_Size       int    `json:"memory_size"`
	Page_Size         int    `json:"page_size"`
	Instructions_Path string `json:"instructions_path"`
	Delay_Response    int    `json:"delay_response"`
}

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

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////MAIN///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func main() {
	// Declaración temporal.

	// Se establece el handler que se utilizará para las diversas situaciones recibidas por el server
	http.HandleFunc("PUT /process", handler_iniciar_proceso)
	http.HandleFunc("DELETE /process/{pid}", handler_finalizar_proceso)
	http.HandleFunc("GET /process/{pid}", handler_estado_proceso)
	http.HandleFunc("GET /process", handler_listar_procesos)

	// Extrae info de config.json
	config := iniciarConfiguracion("config.json")

	// declaro puerto
	port := ":" + strconv.Itoa(config.Port)

	// Listen and serve con info del config.json
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al esuchar en el puerto " + port)
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////FUNCIONES/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func iniciarConfiguracion(filePath string) *MemoriaConfig {
	//En el tp0 usan punteros y guardan la variable en un archivo "globals".
	// No estoy seguro del motivo, y por ahora no lo veo necesario
	var config *MemoriaConfig

	// Abre el archivo
	configFile, err := os.Open(filePath)
	if err != nil {
		// log.Fatal(err.Error())
		fmt.Println("Error: ", err)
	}
	// Cierra el archivo una vez que la función termina (ejecuta el return)
	defer configFile.Close()

	// Decodifica la info del json en la variable config
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	// Devuelve config
	return config
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////HANDLERS/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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

	//Crea una variable tipo ResponseProceso (para confeccionar una respuesta)
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

// primera versión de finalizar proceso, no recibe body (solo un path por medio de la url) y envía una respuesta vacía (mandamos status ok y hacemos que printee el valor del pid recibido para ver que ha sido llamada).
// Cuando haya  procesos se busca por el path {pid}
func handler_finalizar_proceso(w http.ResponseWriter, r *http.Request) {

	//es posible que en un futuro sea necesario convertir esta string a un int
	pid := r.PathValue("pid")

	// Imprime el pid (solo para pruebas)
	fmt.Printf("pid: %s", pid)

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

// primera versión de estado proceso, como es GET no necesita recibir nada
// Cuando haya  procesos se busca por el %s del path {pid}
func handler_estado_proceso(w http.ResponseWriter, r *http.Request) {
	//usando el struct de ResponseProceso envío el estado del proceso

	pid, error := strconv.Atoi(r.PathValue("pid"))

	if error != nil {
		http.Error(w, "Error al obtener el ID del proceso", http.StatusInternalServerError)
		return
	}
	//Crea una variable tipo ResponseProceso (para confeccionar una respuesta)
	var respBody ResponseProceso = ResponseProceso{Pid: pid, Estado: "READY"}

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

func handler_listar_procesos(w http.ResponseWriter, r *http.Request) {

	//Harcodea una lista de procesos, más adelante deberá ser dinámico.
	var listaDeProcesos []ResponseProceso = []ResponseProceso{
		{Pid: 0, Estado: "READY"},
		{Pid: 1, Estado: "BLOCK"},
	}

	//Paso a formato JSON la lista de procesos.
	respuesta, err := json.Marshal(listaDeProcesos)

	//Check si hubo algún error al parsear el JSON.
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Envía respuesta (con estatus como header) al cliente.
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}
