package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/proceso"
)

//-------------------------- STRUCTS --------------------------------------------------

// Estructura administrativa que alamcena las instrucciones de un proceso.

// ================================| MAIN |===================================================\\

var configJson config.Memoria

func main() {
	// Extrae info de config.json
	config.Iniciar("config.json", &configJson)

	// Crea e inicializa la memoria de instrucciones
	memoriaInstrucciones := make(map[uint32][]string)

	// Para que no llore Go
	fmt.Println("Memoria: ", memoriaInstrucciones)

	// Configura el logger
	config.Logger("Memoria.log")

	// Se establece el handler que se utilizará para las diversas situaciones recibidas por el server

	http.HandleFunc("PUT /process", handlerIniciarProceso(memoriaInstrucciones))
	http.HandleFunc("DELETE /process/{pid}", handlerFinalizarProceso)
	http.HandleFunc("GET /process/{pid}", handlerEstadoProceso)
	http.HandleFunc("GET /process", handlerListarProceso)
	http.HandleFunc("GET /instrucciones", handlerEnviarInstruccion(memoriaInstrucciones))

	// Extrae info de config.json

	// declaro puerto
	port := ":" + strconv.Itoa(configJson.Port)

	// Listen and serve con info del config.json
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al escuchar en el puerto " + port)
	}
}

//================================| FUNCIONES |===================================================\\

func guardarInstrucciones(pid uint32, path string, memoriaInstrucciones map[uint32][]string) {
	path = configJson.Instructions_Path + "/" + path
	data := extractInstructions(path)
	insertData(pid, memoriaInstrucciones, data)
}

func extractInstructions(path string) []byte {
	// Lee el archivo
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error al leer el archivo de instrucciones")
		return nil
	}

	// Ahora data es un array de bytes con el contenido del archivo
	return file
}

// funciona todo bien con uint?
func insertData(pid uint32, memoriaInstrucciones map[uint32][]string, data []byte) {
	// Separar instrucciones por medio de tokens
	instrucciones := strings.Split(string(data), "\n")
	// Inserta en la memoria de instrucciones
	memoriaInstrucciones[pid] = instrucciones
	fmt.Println("Instrucciones guardadas en memoria: ")
	for pid, instrucciones := range memoriaInstrucciones {
		fmt.Printf("PID: %d\n", pid)
		for _, instruccion := range instrucciones {
			fmt.Println(instruccion)
		}
		fmt.Println()
	}
}

//================================| HANDLERS |====================================================\\

// Wrapper del handler de iniciar proceso. esto permite pasarle parametros al handler para no tener que usar variables globales y poder pasarle parámetros
func handlerIniciarProceso(memoriaInstrucciones map[uint32][]string) func(http.ResponseWriter, *http.Request) {

	// Handler para iniciar un proceso
	return func(w http.ResponseWriter, r *http.Request) {

		//Crea una variable tipo BodyIniciar (para interpretar lo que se recibe de la request)
		var request proceso.BodyIniciar

		// Decodifica el request (codificado en formato json)
		err := json.NewDecoder(r.Body).Decode(&request)

		// Error Handler de la decodificación
		if err != nil {
			fmt.Printf("Error al decodificar request body: ")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Se guardan las instrucciones en memoria
		guardarInstrucciones(request.PID, request.Path, memoriaInstrucciones)

		// Crea una variable tipo Response (para confeccionar una respuesta)
		var respBody proceso.Response = proceso.Response{PID: request.PID}

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

		// // Luego sincronzar para que no se creen varios procesos a la vez
		// proceso.Counter++
		// fmt.Println("Counter:", proceso.Counter)
	}

}

// primera versión de finalizar proceso, no recibe body (solo un path por medio de la url) y envía una respuesta vacía (mandamos status ok y hacemos que printee el valor del pid recibido para ver que ha sido llamada).
// Cuando haya  procesos se busca por el path {pid}
func handlerFinalizarProceso(w http.ResponseWriter, r *http.Request) {

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
func handlerEstadoProceso(w http.ResponseWriter, r *http.Request) {
	//usando el struct de Response envío el estado del proceso

	pid, error := strconv.Atoi(r.PathValue("pid"))

	if error != nil {
		http.Error(w, "Error al obtener el ID del proceso", http.StatusInternalServerError)
		return
	}
	//Crea una variable tipo Response (para confeccionar una respuesta)
	var respBody proceso.Response = proceso.Response{PID: uint32(pid), Estado: "READY"}

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

/*
Se encargará de mostrar por consola y retornar por la api el listado de procesos
que se encuentran en el sistema con su respectivo estado dentro de cada uno de ellos.
*/
func handlerListarProceso(w http.ResponseWriter, r *http.Request) {

	//Harcodea una lista de procesos, más adelante deberá ser dinámico.
	var listaDeProcesos []proceso.Response = []proceso.Response{
		{PID: 0, Estado: "READY"},
		{PID: 1, Estado: "BLOCK"},
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

func handlerEnviarInstruccion(memoriaInstrucciones map[uint32][]string) func(http.ResponseWriter, *http.Request) {

	// Handler para enviar una instruccion
	return func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		pid, errPid := strconv.ParseUint(queryParams.Get("PID"), 10, 32)
		pc, errPC := strconv.ParseUint(queryParams.Get("PC"), 10, 32)

		if errPid != nil || errPC != nil {
			return
		}

		instruccion := memoriaInstrucciones[uint32(pid)][uint32(pc)]
		fmt.Println(instruccion)

		// respuesta, err := json.Marshal(instruccion)
		// fmt.Println(respuesta)

		// if err != nil {
		// 	http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		// 	return
		// }

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(instruccion))
	}
}
