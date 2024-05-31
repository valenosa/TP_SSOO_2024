package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/kernel/funciones"
	"github.com/sisoputnfrba/tp-golang/kernel/logueano"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

//*======================================| MAIN |======================================\\

func main() {

	config.Iniciar("config.json", &funciones.ConfigJson)
	funciones.Cont_producirPCB = make(chan int, funciones.ConfigJson.Multiprogramming)
	funciones.Bin_hayPCBenREADY = make(chan int, funciones.ConfigJson.Multiprogramming+1)

	// Configura el logger
	config.Logger("Kernel.log")

	// ======== HandleFunctions ========

	//PLANIFICACION
	http.HandleFunc("PUT /plani", handlerIniciarPlanificacion)
	http.HandleFunc("DELETE /plani", handlerDetenerPlanificacion)

	//PROCESOS
	http.HandleFunc("GET /process/{pid}", handlerEstadoProceso)
	http.HandleFunc("GET /process", handlerListarProceso)

	http.HandleFunc("PUT /process", handlerIniciarProceso)
	http.HandleFunc("DELETE /process/{pid}", handlerFinalizarProceso)

	//ENTRADA SALIDA
	http.HandleFunc("POST /interfazConectada", handlerConexionInterfazIO)
	http.HandleFunc("POST /instruccion", handlerEjecutarInstruccionEnIO)

	//Inicio el servidor de Kernel
	config.IniciarServidor(funciones.ConfigJson.Port)

}

//*======================================| HANDLERS |======================================\\

//----------------------( PLANIFICACION )----------------------\\

func handlerIniciarPlanificacion(w http.ResponseWriter, r *http.Request) {

	fmt.Println("IniciarPlanificacion-------------------------")
	funciones.TogglePlanificador = true

	funciones.OnePlani.Lock()
	go funciones.Planificador()

	w.WriteHeader(http.StatusOK)
}

func handlerDetenerPlanificacion(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("DetenerPlanificacion-------------------------")

	funciones.TogglePlanificador = false
	funciones.OnePlani.Unlock()

	w.WriteHeader(http.StatusOK)
}

//----------------------( PROCESOS )----------------------\\

func handlerIniciarProceso(w http.ResponseWriter, r *http.Request) {

	fmt.Println("IniciarProceso-------------------------")

	//----------- RECIBE ---------
	//variable que recibirá la request.
	var request structs.RequestIniciarProceso

	// Decodifica en formato JSON la request.
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		fmt.Println(err) //! Borrar despues.
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("Path: %s\n", request.Path)

	//----------- EJECUTA ---------

	// Se crea un nuevo PCB en estado NEW
	var nuevoPCB structs.PCB

	funciones.Mx_ConterPID.Lock()
	nuevoPCB.PID = funciones.CounterPID
	funciones.Mx_ConterPID.Unlock()

	nuevoPCB.Estado = "NEW"

	//----------- Va a memoria ---------
	bodyIniciarProceso, err := json.Marshal(structs.BodyIniciarProceso{PID: nuevoPCB.PID, Path: request.Path})
	if err != nil {
		return
	}

	//Envía el path a memoria para que cree el proceso
	respuesta := config.Request(funciones.ConfigJson.Port_Memory, funciones.ConfigJson.Ip_Memory, "PUT", "process", bodyIniciarProceso)
	if respuesta == nil {
		return
	}

	var respMemoIniciarProceso structs.BodyIniciarProceso
	// Decodifica en formato JSON la request.
	err = json.NewDecoder(respuesta.Body).Decode(&respMemoIniciarProceso)
	if err != nil {
		fmt.Println(err) ////! Borrar despues.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//----------------------------

	//Asigna un nuevo valor pid para la proxima response.
	funciones.Mx_ConterPID.Lock()
	funciones.CounterPID++
	funciones.Mx_ConterPID.Unlock()

	//Verifica si puede producir un PCB (por Multiprogramacion)
	funciones.Cont_producirPCB <- 0

	// Si todo es correcto agregamos el PID al PCB
	nuevoPCB.Estado = "READY"

	// Agrega el nuevo PCB a readyQueue
	funciones.AdministrarQueues(nuevoPCB)

	//^ log obligatorio (2/6) (NEW->Ready): Cambio de Estado
	logueano.CambioDeEstado("NEW", nuevoPCB)

	// ----------- DEVUELVE -----------

	respIniciarProceso, err := json.Marshal(respMemoIniciarProceso.PID)
	if err != nil {
		fmt.Println(err) //! Borrar despues.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respIniciarProceso)
}

// TODO: Probar
func handlerFinalizarProceso(w http.ResponseWriter, r *http.Request) {

	fmt.Println("DetenerEstadoProceso-------------------------")

	//--------- RECIBE ---------
	pid, err := strconv.ParseUint(r.PathValue("pid"), 10, 32)
	if err != nil {
		fmt.Println(err) //! Borrar despues.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	funciones.Interrupt(uint32(pid), "Finalizar PROCESO") // Interrumpe el proceso

	//--------- EJECUTA ---------

	//* Busca el Proceso (PID) lo desencola y lo pasa a EXIT (si esta en EXEC, lo interrumpe y lo pasa a EXIT)

	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
}

// TODO: Tomar los procesos creados (BLock, Ready y Exec) y devolverlos en una lista
func handlerListarProceso(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("ListarProceso-------------------------")

	//----------- EJECUTA -----------

	//Harcodea una lista de procesos, más adelante deberá ser dinámico
	var listaDeProcesos []structs.ResponseListarProceso = []structs.ResponseListarProceso{
		{PID: 0, Estado: "READY"},
		{PID: 1, Estado: "BLOCK"},
	}

	//----------- DEVUELVE -----------

	//Paso a formato JSON la lista de procesos
	respuesta, err := json.Marshal(listaDeProcesos)
	if err != nil {
		fmt.Println(err) //! Borrar despues.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Envía respuesta al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

// TODO: Busca el proceso deseado y devuelve el estado en el que se encuentra
func handlerEstadoProceso(w http.ResponseWriter, r *http.Request) {

	fmt.Println("DetenerEstadoProceso-------------------------")

	//--------- RECIBE ---------
	pid, err := strconv.Atoi(r.PathValue("pid"))
	if err != nil {
		fmt.Println(err) //! Borrar despues.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("PID:", pid)

	//--------- EJECUTA ---------

	//TODO: Busca en base al pid el proceso en todas las colas (y el map de BLOCK) y devuelvo el estado
	var respEstadoProceso structs.ResponseEstadoProceso = structs.ResponseEstadoProceso{State: "ANASHE"}

	//--------- DEVUELVE ---------
	//Crea una variable tipo Response
	respuesta, err := json.Marshal(respEstadoProceso)

	// Error Handler de la codificación
	if err != nil {
		fmt.Println(err) //! Borrar despues.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

//----------------------( I/O )----------------------\\

// Recibe una iterfaz y la guarda en InterfacesConectadas
func handlerConexionInterfazIO(w http.ResponseWriter, r *http.Request) {

	fmt.Println("ConexionInterfazIO-------------------------")

	// Almaceno la interfazConectada en una variable
	var interfazConectada structs.RequestConectarInterfazIO
	err := json.NewDecoder(r.Body).Decode(&interfazConectada)
	if err != nil {
		fmt.Println(err) //! Borrar despues.
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Imprime la solicitud
	fmt.Println("Request path:", interfazConectada) //! Borrar despues

	//Guarda la interfazConectada en el map de interfaces conectadas
	funciones.InterfacesConectadas.Set(interfazConectada.NombreInterfaz, interfazConectada.Interfaz)

	//go planificarInterfaz()
}

var Bin_hayInterfaz = make(chan int, 1)

func planificarInterfaz() {

	for {
		//Espero a que la interfaz tenga elementos en su cola de BLOCK
		<-Bin_hayInterfaz

	}

}

// TODO: Implementar para los demás tipos de interfaces (cambiar tipos de datos en request y body)
func handlerEjecutarInstruccionEnIO(w http.ResponseWriter, r *http.Request) {

	// Se crea una variable para almacenar la instrucción recibida en la solicitud
	var requestInstruccionIO structs.RequestEjecutarInstruccionIO

	// Se decodifica el cuerpo de la solicitud en formato JSON
	marshalError := json.NewDecoder(r.Body).Decode(&requestInstruccionIO)
	if marshalError != nil {
		fmt.Println(marshalError) //! Borrar despues.
		http.Error(w, marshalError.Error(), http.StatusBadRequest)
		return
	}

	// Imprime la solicitud
	fmt.Println("Request de ejecutar ", requestInstruccionIO.Instruccion, " por ", requestInstruccionIO.NombreInterfaz) //!Borrar despues

	// Verifica que la Interfaz este Conectada
	interfazConectada, encontrado := funciones.InterfacesConectadas.Get(requestInstruccionIO.NombreInterfaz)
	if !encontrado {
		funciones.DesalojarProcesoIO(requestInstruccionIO.PidDesalojado)
		fmt.Println("Interfaz no conectada.")
		http.Error(w, "Interfaz no conectada.", http.StatusNotFound)
		return
	}

	//Verifica que la instruccion sea compatible con el tipo de interfazConectada
	laInstruccionEsValida := funciones.ValidarInstruccionIO(interfazConectada.TipoInterfaz, requestInstruccionIO.Instruccion)
	if !laInstruccionEsValida {
		funciones.DesalojarProcesoIO(requestInstruccionIO.PidDesalojado)
		fmt.Println("Interfaz incompatible.")
		http.Error(w, "Interfaz incompatible.", http.StatusBadRequest)
		return
	}

	// Agrega el Proceso a la cola de bloqueados de la interfazConectada
	interfazConectada.QueueBlock = append(interfazConectada.QueueBlock, requestInstruccionIO.PidDesalojado)
	//Actualiza la lista de interfaces conectadas
	funciones.InterfacesConectadas.Set(requestInstruccionIO.NombreInterfaz, interfazConectada)

	//Bin_hayInterfaz <- 0

	//TODO: ----------------------------------------(ESTO LO HACE EL PLANIFICADOR)

	// Manda a ejecutar a la interfaz
	body, marshalError := json.Marshal(requestInstruccionIO)
	if marshalError != nil {
		fmt.Println(marshalError) //! Borrar despues.
		http.Error(w, marshalError.Error(), http.StatusInternalServerError)
		return
	}

	// Envía la instrucción a ejecutar a la interfazConectada (Puerto)
	respuesta := config.Request(interfazConectada.PuertoInterfaz, "localhost", "POST", requestInstruccionIO.Instruccion, body)

	// Verifica que no hubo error en la request
	if respuesta == nil {
		fmt.Println(respuesta) //! Borrar despues.
		http.Error(w, "Respuesta vacia.", http.StatusInternalServerError)
		return
	}

	// Si la interfazConectada pudo ejecutar la instrucción, pasa el Proceso a READY.
	if respuesta.StatusCode == http.StatusOK {
		// Pasa el proceso a READY y lo quita de la lista de bloqueados.
		funciones.DesalojarProcesoIO(requestInstruccionIO.PidDesalojado)
		pcbDesalojado := funciones.MapBLOCK.Delete(requestInstruccionIO.PidDesalojado)
		pcbDesalojado.Estado = "READY"
		funciones.AdministrarQueues(pcbDesalojado)
		return
	}
}
