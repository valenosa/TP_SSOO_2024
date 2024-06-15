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

	// ======== Make ========

	TLB := make(funciones.TLB)                                                                   // TLB (Translation Lookaside Buffer).
	prioridadTLB := make([]funciones.ElementoPrioridad, funciones.ConfigJson.Number_Felling_tlb) // Prioridad de la TLB (para el algoritmo de reemplazo de páginas

	// ======== HandleFunctions ========
	// Se establece el handler que se utilizará para las diversas situaciones recibidas por el server
	http.HandleFunc("POST /exec", handlerEjecutarProceso(&TLB, &prioridadTLB))
	http.HandleFunc("POST /interrupciones", handlerInterrupcion)

	// Extrae info de config.json
	config.Iniciar("config.json", &funciones.ConfigJson)

	//inicio el servidor de CPU
	config.IniciarServidor(funciones.ConfigJson.Port)
}

// *======================================| HANDLERS |=======================================\\

// Maneja la ejecución de un proceso a través de un PCB
// Devuelve a dispatch el contexto de ejecución y el motivo del desalojo.
func handlerEjecutarProceso(TLB *funciones.TLB, prioridadesTLB *[]funciones.ElementoPrioridad) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		//--------- REQUEST ---------

		// Crea una variable tipo BodyIniciar (para interpretar lo que se recibe de la pcbRecibido)
		var pcbRecibido structs.PCB

		// Decodifica el request (codificado en formato JSON)
		err := json.NewDecoder(r.Body).Decode(&pcbRecibido)

		// Error Handler de la decodificación
		if err != nil {
			fmt.Println(err) //! Borrar despues.
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//--------- EJECUTAR ---------

		fmt.Println("Se está ejecutando el proceso: ", pcbRecibido.PID)

		funciones.PidEnEjecucion = pcbRecibido.PID

		// Ejecuta el ciclo de instrucción.
		funciones.RegistrosCPU = pcbRecibido.RegistrosUsoGeneral
		funciones.EjecutarCiclosDeInstruccion(&pcbRecibido, TLB, prioridadesTLB)

		//--------- RESPUESTA ---------

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
}

// Checkea que Kernel no haya enviado interrupciones
func handlerInterrupcion(w http.ResponseWriter, r *http.Request) {

	//--------- REQUEST ---------

	queryParams := r.URL.Query()

	funciones.MotivoDeDesalojo = queryParams.Get("interrupt_type")
	fmt.Println("Se recibió una interrupción de tipo: ", funciones.MotivoDeDesalojo) //!Solamente para chequear que reciba todo bien, después se borra.

	PID, errPid := strconv.ParseUint(queryParams.Get("PID"), 10, 32)

	if errPid != nil {
		return
	}

	if uint32(PID) != funciones.PidEnEjecucion {
		return
	}

	funciones.HayInterrupcion = true
}
