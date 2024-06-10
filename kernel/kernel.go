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

	// Inicializar recursos
	funciones.LeerRecursos(funciones.ConfigJson.Resources, funciones.ConfigJson.Resource_Instances)

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
	http.HandleFunc("POST /instruccionIO", handlerEjecutarInstruccionEnIO)

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
		fmt.Println(err) //! Borrar despues.
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Interfaz Conectada:", interfazConectada) //! Borrar despues

	//Guarda la interfazConectada en el map de interfaces conectadas
	funciones.InterfacesConectadas.Set(interfazConectada.NombreInterfaz, interfazConectada.Interfaz)
}

func handlerEjecutarInstruccionEnIO(w http.ResponseWriter, r *http.Request) {

	//--------- RECIBE ---------

	// Se decodifica el cuerpo de la solicitud en formato JSON
	var requestInstruccionIO structs.RequestEjecutarInstruccionIO
	marshalError := json.NewDecoder(r.Body).Decode(&requestInstruccionIO)
	if marshalError != nil {
		fmt.Println(marshalError) //! Borrar despues.
		http.Error(w, marshalError.Error(), http.StatusBadRequest)
		return
	}

	// Imprime la solicitud
	fmt.Println("Request de ejecutar ", requestInstruccionIO.Instruccion, " por ", requestInstruccionIO.NombreInterfaz) //! Borrar despues

	//--------- EJECUTA ---------

	//--- VALIDA

	// Verifica que la Interfaz este Conectada
	interfazSolicitada, encontrado := funciones.InterfacesConectadas.Get(requestInstruccionIO.NombreInterfaz)
	if !encontrado {
		funciones.DesalojarProcesoIO(requestInstruccionIO.PidDesalojado)
		fmt.Println("Interfaz no conectada.")
		http.Error(w, "Interfaz no conectada.", http.StatusNotFound)
		return
	}

	//Verifica que la instruccion sea compatible con el tipo de interfazConectada
	laInstruccionEsValida := funciones.ValidarInstruccionIO(interfazSolicitada.TipoInterfaz, requestInstruccionIO.Instruccion)
	if !laInstruccionEsValida {
		funciones.DesalojarProcesoIO(requestInstruccionIO.PidDesalojado)
		fmt.Println("Instruccion incompatible.")
		http.Error(w, "Instruccion incompatible.", http.StatusBadRequest)
		return
	}

	//--- ENVIA A EJECUTAR A IO

	// Codifica instruccion a ejecutar en JSON
	body, marshalError := json.Marshal(requestInstruccionIO)
	if marshalError != nil {
		fmt.Println(marshalError) //! Borrar despues.
		http.Error(w, marshalError.Error(), http.StatusInternalServerError)
		return
	}

	// Envía la instrucción a ejecutar a la interfazConectada (Puerto)
	query := interfazSolicitada.TipoInterfaz + " /" + requestInstruccionIO.Instruccion

	respuesta := config.Request(interfazSolicitada.PuertoInterfaz, "localhost", "POST", query, body)
	if respuesta == nil {
		// Si no conecta con la interfaz, la elimina del map de las interfacesConectadas y desaloja el proceso.
		funciones.DesalojarProcesoIO(requestInstruccionIO.PidDesalojado)
		funciones.InterfacesConectadas.Delete(requestInstruccionIO.NombreInterfaz)
		fmt.Println("Interfaz desconectada.")
		http.Error(w, "Interfaz desconectada.", http.StatusInternalServerError)
		return
	}

	if respuesta.StatusCode != http.StatusOK {
		http.Error(w, "Error en la respuesta de I/O.", http.StatusInternalServerError)
		return
	}

	//--- VUELVE DE IO

	// Pasa el proceso a READY y lo quita de la lista de bloqueados.
	pcbDesalojado := funciones.MapBLOCK.Delete(requestInstruccionIO.PidDesalojado)
	pcbDesalojado.Estado = "READY"
	funciones.AdministrarQueues(pcbDesalojado)
}
