package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/memoria/funciones"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

//================================| MAIN |================================\\

var configJson config.Memoria

func main() {

	// Extrae info de config.json
	config.Iniciar("config.json", &configJson)

	// Crea e inicializa la memoria de instrucciones
	memoriaInstrucciones := make(map[uint32][]string)

	// Configura el logger
	config.Logger("Memoria.log")

	// ======== HandleFunctions ========
	http.HandleFunc("PUT /process", handlerIniciarProceso(memoriaInstrucciones))
	http.HandleFunc("DELETE /process/{pid}", handlerFinalizarProceso)
	http.HandleFunc("GET /process/{pid}", handlerEstadoProceso)
	http.HandleFunc("GET /process", handlerListarProceso)
	http.HandleFunc("GET /instrucciones", handlerEnviarInstruccion(memoriaInstrucciones))

	//inicio el servidor de Memoria
	go config.IniciarServidor(configJson.Port)
}

//================================| HANDLERS |====================================================\\

// Wrapper que crea un PCB con el pid recibido.
func handlerIniciarProceso(memoriaInstrucciones map[uint32][]string) func(http.ResponseWriter, *http.Request) {

	// Handler para iniciar un proceso.
	return func(w http.ResponseWriter, r *http.Request) {

		//variable que recibirá la request.
		var request structs.BodyIniciar

		// Decodifica en formato JSON la request.
		err := json.NewDecoder(r.Body).Decode(&request)

		// Error Handler de la decodificación
		if err != nil {
			fmt.Printf("Error al decodificar request body: ")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Se guardan las instrucciones en un map de memoria.
		funciones.GuardarInstrucciones(request.PID, request.Path, memoriaInstrucciones)

		// Crea una variable tipo Response (para confeccionar una respuesta)
		var respBody structs.ResponseIniciarProceso = structs.ResponseIniciarProceso{PID: request.PID}

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

}

// TODO: busca el pid y lo interrumpe si está en ejecución y lo pasa a EXIT, de estar encolado, solamente lo desencola y lo pasa a EXIT
func handlerFinalizarProceso(w http.ResponseWriter, r *http.Request) {

	//es posible que en un futuro sea necesario convertir esta string a un int
	pid := r.PathValue("pid")

	// Imprime el pid (solo para pruebas)
	fmt.Printf("pid: %s", pid)

	// Pasa a JSON un string vacío.
	respuesta, err := json.Marshal("")

	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

// TODO: Busca el proceso deseado y devuelve el estado en el que se encuentra
func handlerEstadoProceso(w http.ResponseWriter, r *http.Request) {
	//usando el struct de Response envío el estado del proceso

	pid, error := strconv.Atoi(r.PathValue("pid"))

	if error != nil {
		http.Error(w, "Error al obtener el ID del proceso", http.StatusInternalServerError)
		return
	}
	//Crea una variable tipo Response (para confeccionar una respuesta)
	var respBody structs.ResponseIniciarProceso = structs.ResponseIniciarProceso{PID: uint32(pid), Estado: "READY"}

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

// TODO tomar los procesos creados (BLock, Ready y Exec) y devolverlos en una lista.
func handlerListarProceso(w http.ResponseWriter, r *http.Request) {

	//Harcodea una lista de procesos, más adelante deberá ser dinámico.
	var listaDeProcesos []structs.ResponseIniciarProceso = []structs.ResponseIniciarProceso{
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

// Envía a CPU la instrucción correspondiente al pid y el pc del map de memoria
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

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(instruccion))
	}
}
