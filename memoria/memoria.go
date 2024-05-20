package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/utils/APIs/kernel-memoria/proceso"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

//-------------------------- STRUCTS --------------------------------------------------

// Estructura para la tabla de páginas
type Pagina struct {
	numero uint
	marco  uint
}

// Estructura administrativa que relaciona una/s determinada/s página/s con las instrucciones de un proceso.
type PaginaInstrucciones struct { // Cambiar nombre
	PID             uint32
	pagina          Pagina
	cantidadPaginas uint
}

// ================================| MAIN |===================================================\\

var configJson config.Memoria

func main() {
	// Extrae info de config.json
	config.Iniciar("config.json", &configJson)

	// Crea la memoria que representa el espacio de usuario
	memoria := make([]byte, configJson.Memory_Size)

	// Se crea un índice para saber a partir de donde escribir en memoria
	var memoryIndex uint = 0

	// Tabla de páginas que contienen instrucciones
	var tablaPaginas []PaginaInstrucciones

	// Para que no llore Go
	fmt.Println("Memoria: ", memoria)

	// Configura el logger
	config.Logger("Memoria.log")

	// Se establece el handler que se utilizará para las diversas situaciones recibidas por el server

	http.HandleFunc("PUT /process", handlerIniciarProceso(memoria, memoryIndex, tablaPaginas))
	http.HandleFunc("DELETE /process/{pid}", handlerFinalizarProceso)
	http.HandleFunc("GET /process/{pid}", handlerEstadoProceso)
	http.HandleFunc("GET /process", handlerListarProceso)

	// Extrae info de config.json

	// declaro puerto
	port := ":" + strconv.Itoa(configJson.Port)

	// Listen and serve con info del config.json
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al esuchar en el puerto " + port)
	}
}

//================================| FUNCIONES |===================================================\\

func guardarInstrucciones(pid uint32, path string, memoria []byte, memoryIndex uint, tablaInstrucciones []PaginaInstrucciones) {
	data := extractInstructions(path)
	insertData(pid, memoria, data, memoryIndex, tablaInstrucciones)
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
func insertData(pid uint32, memoria []byte, data []byte, memoryIndex uint, tablaPaginas []PaginaInstrucciones) {

	// Calcula el número de páginas necesarias, redondeando hacia arriba
	numPages := uint(math.Ceil(float64(len(data)) / float64(configJson.Page_Size)))

	// Verifica si hay suficiente espacio en la memoria
	if memoryIndex+uint(numPages)*configJson.Memory_Size > uint(len(memoria)) {
		//Debería "limpiar" memoria con el algoritmo elegido
		fmt.Println("No hay suficiente espacio en la memoria")
		return
	}

	// Copia data en memoria
	copy(memoria[memoryIndex:], data)

	// Actualiza la tabla de páginas
	tablaPaginas = append(tablaPaginas, PaginaInstrucciones{PID: pid, pagina: Pagina{numero: uint(memoryIndex / configJson.Memory_Size), marco: memoryIndex}, cantidadPaginas: numPages})

	// Actualiza currentIndex para la próxima inserción
	memoryIndex += uint(numPages) * configJson.Page_Size

}

//================================| HANDLERS |====================================================\\

// Wrapper del handler de iniciar proceso. esto permite pasarle parametros al handler para no tener que usar variables globales y poder pasarle parámetros
func handlerIniciarProceso(memoria []byte, memoryIndex uint, tablaPaginas []PaginaInstrucciones) func(http.ResponseWriter, *http.Request) {

	// Handler para iniciar un proceso
	return func(w http.ResponseWriter, r *http.Request) {

		//Crea uan variable tipo BodyIniciar (para interpretar lo que se recibe de la request)
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
		guardarInstrucciones(request.PID, request.Path, memoria, memoryIndex, tablaPaginas)

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
