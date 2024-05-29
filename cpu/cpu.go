package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/cpu/funciones"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

//*======================================| MAIN |=======================================\\

func main() {

	// Configura el logger
	config.Logger("CPU.log")

	// ======== HandleFunctions ========
	// Se establece el handler que se utilizará para las diversas situaciones recibidas por el server
	http.HandleFunc("POST /exec", handlerEjecutarProceso)
	http.HandleFunc("POST /interrupciones", handlerInterrupcion)

	// Extrae info de config.json
	config.Iniciar("config.json", &funciones.ConfigJson)

	//inicio el servidor de CPU
	config.IniciarServidor(funciones.ConfigJson.Port)
}

// *======================================| HANDLERS |=======================================\\

// Maneja la ejecución de un proceso a través de un PCB
// Devuelve al despachador el contexto de ejecución y el motivo del desalojo.
func handlerEjecutarProceso(w http.ResponseWriter, r *http.Request) {
	// Crea una variable tipo BodyIniciar (para interpretar lo que se recibe de la pcbRecibido)
	var pcbRecibido structs.PCB

	// Decodifica el request (codificado en formato JSON)
	err := json.NewDecoder(r.Body).Decode(&pcbRecibido)

	// Error Handler de la decodificación
	if err != nil {
		fmt.Printf("Error al decodificar request body: ")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Ejecuta el ciclo de instrucción.
	funciones.PidEnEjecucion = pcbRecibido.PID
	funciones.EjecutarCiclosDeInstruccion(&pcbRecibido)

	fmt.Println("Se está ejecutando el proceso: ", pcbRecibido.PID)

	// Devuelve a dispatch el contexto de ejecucion y el motivo del desalojo
	respuesta, err := json.Marshal(structs.RespuestaDispatch{
		MotivoDeDesalojo: funciones.MotivoDeDesalojo,
		PCB:              pcbRecibido,
	})

	// Error Handler de la codificación
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

// Checkea que Kernel no haya enviado interrupciones
func handlerInterrupcion(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	// Está en una global; despues cambiar.
	funciones.MotivoDeDesalojo = queryParams.Get("interrupt_type")

	PID, errPid := strconv.ParseUint(queryParams.Get("PID"), 10, 32)

	if errPid != nil {
		return
	}

	if uint32(PID) != funciones.PidEnEjecucion {
		return
	}

	funciones.HayInterrupcion = true

	//TODO: Checkear si es necesario lo de abajo (27/05/24).
	/*En caso de que haya interrupcion,
	se devuelve el Contexto de Ejecución actualizado al Kernel con motivo de la interrupción.*/

	// respuesta, err := json.Marshal(instruccion)
	// fmt.Println(respuesta)

	// if err != nil {
	// 	http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
	// 	return
	// }

	// w.WriteHeader(http.StatusOK)
	// w.Write([]byte(instruccion))
}
