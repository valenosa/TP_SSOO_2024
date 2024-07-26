package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/kernel/funciones"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/logueano"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

//*======================================| MAIN |======================================\\

func main() {

	configPath := os.Args[1]
	config.Iniciar(configPath, &funciones.ConfigJson)

	funciones.Cont_producirPCB = make(chan int, funciones.ConfigJson.Multiprogramming)
	funciones.Bin_hayPCBenREADY = make(chan int, funciones.ConfigJson.Multiprogramming+1)

	// Inicializar recursos
	funciones.LeerRecursos(funciones.ConfigJson.Resources, funciones.ConfigJson.Resource_Instances)

	// Configura el logger (aux en funciones.go)
	logueano.Logger("kernel.log")

	funciones.Auxlogger = logueano.InitAuxLog("kernel")

	// ======== Iniciamos Planificador ========

	funciones.TogglePlanificador.Lock()
	go funciones.Planificador()

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

	//RECURSOS
	http.HandleFunc("POST /wait", handlerWait)
	http.HandleFunc("POST /signal", handlerSignal)

	// ======== Inicio Serivor de Kernel ========
	config.IniciarServidor(funciones.ConfigJson.Port)

}

//*======================================| HANDLERS |======================================\\

// ----------------------( PLANIFICACION )----------------------\\
func handlerIniciarPlanificacion(w http.ResponseWriter, r *http.Request) {

	if !funciones.PlanificadorIniciado {
		funciones.PlanificadorIniciado = true
		funciones.TogglePlanificador.Unlock()
	}

	w.WriteHeader(http.StatusOK)
}

func handlerDetenerPlanificacion(w http.ResponseWriter, r *http.Request) {

	if funciones.PlanificadorIniciado {
		funciones.PlanificadorIniciado = false
		funciones.TogglePlanificador.Lock()
	}

	w.WriteHeader(http.StatusOK)
}

//----------------------( PROCESOS )----------------------\\

func handlerIniciarProceso(w http.ResponseWriter, r *http.Request) {

	//----------- RECIBE ---------
	//variable que recibirá la request.
	var request structs.IniciarProceso

	// Decodifica en formato JSON la request.
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		logueano.Error(funciones.Auxlogger, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logueano.MensajeConFormato(funciones.Auxlogger, "Path: %s\n", request.Path)

	//----------- EJECUTA ---------

	// Se crea un nuevo PCB en estado NEW
	var nuevoPCB structs.PCB

	nuevoPCB.PID = request.PID

	nuevoPCB.Estado = "NEW"
	funciones.AdministrarQueues(nuevoPCB)

	//----------- Va a memoria ---------
	bodyIniciarProceso, err := json.Marshal(structs.IniciarProceso{PID: nuevoPCB.PID, Path: request.Path})
	if err != nil {
		return
	}

	//Envía el path a memoria para que cree el proceso
	respuesta, err := config.Request(funciones.ConfigJson.Port_Memory, funciones.ConfigJson.Ip_Memory, "PUT", "process", bodyIniciarProceso)
	if err != nil {
		logueano.Error(funciones.Auxlogger, err)
		return
	}

	var respMemoIniciarProceso structs.IniciarProceso
	// Decodifica en formato JSON la request.
	err = json.NewDecoder(respuesta.Body).Decode(&respMemoIniciarProceso)
	if err != nil {
		logueano.Error(funciones.Auxlogger, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//----------------------------

	//Verifica si puede producir un PCB (por Multiprogramacion)
	funciones.Cont_producirPCB <- 0

	//Lo Elimino de la lista de NEW
	funciones.ListaNEW.Extract(nuevoPCB.PID)

	// Si todo es correcto agregamos el PID al PCB
	nuevoPCB.Estado = "READY"

	// Agrega el nuevo PCB a readyQueue
	funciones.AdministrarQueues(nuevoPCB)

	//^ log obligatorio (2/6)
	logueano.CambioDeEstado("NEW", nuevoPCB.Estado, nuevoPCB.PID)

	// ----------- DEVUELVE -----------

	respIniciarProceso, err := json.Marshal(respMemoIniciarProceso.PID)
	if err != nil {
		logueano.Error(funciones.Auxlogger, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respIniciarProceso)
}

func handlerFinalizarProceso(w http.ResponseWriter, r *http.Request) {

	//--------- RECIBE ---------
	pid, err := strconv.ParseUint(r.PathValue("pid"), 10, 32)
	if err != nil {
		logueano.Error(funciones.Auxlogger, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//--------- EJECUTA ---------

	funciones.Interrupt(uint32(pid), "INTERRUPTED_BY_USER") // Interrumpe el proceso

	pcb, encontrado := funciones.ExtraerPCB(uint32(pid))
	if encontrado {
		pcb.Estado = "EXIT"
		funciones.AdministrarQueues(pcb)
	}

	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
}

func handlerListarProceso(w http.ResponseWriter, r *http.Request) {

	//----------- EJECUTA -----------
	//Recorre la lista de NEW
	var listaDeProcesos []structs.ResponseListarProceso

	listaDeProcesos = funciones.AppendListaProceso(listaDeProcesos, &funciones.ListaNEW)
	listaDeProcesos = funciones.AppendListaProceso(listaDeProcesos, &funciones.ListaREADY)
	var procesoExec = structs.ResponseListarProceso{PID: funciones.ProcesoExec.PID, Estado: funciones.ProcesoExec.Estado}
	if procesoExec.Estado == "EXEC" {
		listaDeProcesos = append(listaDeProcesos, procesoExec)
	}
	listaDeProcesos = funciones.AppendMapProceso(listaDeProcesos, &funciones.MapBLOCK)
	listaDeProcesos = funciones.AppendListaProceso(listaDeProcesos, &funciones.ListaEXIT)

	//----------- DEVUELVE -----------

	//Paso a formato JSON la lista de procesos
	respuesta, err := json.Marshal(listaDeProcesos)
	if err != nil {
		logueano.Error(funciones.Auxlogger, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Envía respuesta al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

func handlerEstadoProceso(w http.ResponseWriter, r *http.Request) {

	//--------- RECIBE ---------
	pid, err := strconv.Atoi(r.PathValue("pid"))
	if err != nil {
		logueano.Error(funciones.Auxlogger, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logueano.MensajeConFormato(funciones.Auxlogger, "PID: %d\n", uint32(pid))

	//--------- EJECUTA ---------

	pcb, encontrado := funciones.BuscarPCB(uint32(pid))
	if !encontrado {
		fmt.Println("Proceso no encontrado")
		return
	}

	var respEstadoProceso structs.ResponseEstadoProceso = structs.ResponseEstadoProceso{State: pcb.Estado}

	//--------- DEVUELVE ---------

	//Crea una variable tipo Response
	respuesta, err := json.Marshal(respEstadoProceso)
	if err != nil {
		logueano.Error(funciones.Auxlogger, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

//----------------------( RECURSOS )----------------------\\

func handlerWait(w http.ResponseWriter, r *http.Request) {

	//--------- RECIBE ---------

	// Almaceno el recurso en una variable
	var recursoSolicitado structs.RequestRecurso
	err := json.NewDecoder(r.Body).Decode(&recursoSolicitado)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//--------- EJECUTA ---------

	respAsignacionRecurso := "OK: Recurso asignado"

	//Busco el recurso solicitado
	var recurso, find = funciones.MapRecursos[recursoSolicitado.NombreRecurso]
	if !find {
		//Si no existe el recurso
		respAsignacionRecurso = "ERROR: Recurso no existe"
	} else {

		//Resto uno al la cantidad de instancias del recurso
		recurso.Instancias--
		if recurso.Instancias < 0 {

			//Agrego PID a su lista de bloqueados
			recurso.ListaBlock.Append(recursoSolicitado.PidSolicitante)

			//^log obligatorio (6/6)
			logueano.MotivoBloqueo(recursoSolicitado.PidSolicitante, recursoSolicitado.NombreRecurso)

			respAsignacionRecurso = "BLOQUEAR: Recurso no disponible"
		}

	}

	//--------- DEVUELVE ---------

	respuesta, err := json.Marshal(respAsignacionRecurso)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

func handlerSignal(w http.ResponseWriter, r *http.Request) {

	//--------- RECIBE ---------

	// Almaceno el recurso en una variable
	var recursoLiberado string
	err := json.NewDecoder(r.Body).Decode(&recursoLiberado)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//--------- EJECUTA ---------

	var _, find = funciones.MapRecursos[recursoLiberado]
	if !find {
		http.Error(w, "INVALID_RESOURCE", http.StatusNotFound)
		return
	}

	funciones.LiberarRecurso(recursoLiberado)

	w.WriteHeader(http.StatusOK)
}

//----------------------( I/O )----------------------\\

// Recibe una iterfaz y la guarda en InterfacesConectadas
func handlerConexionInterfazIO(w http.ResponseWriter, r *http.Request) {

	// Almaceno la interfazConectada en una variable
	var interfazConectada structs.RequestConectarInterfazIO
	err := json.NewDecoder(r.Body).Decode(&interfazConectada)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("Interfaz conectada: ", interfazConectada)
	logueano.MensajeConFormato(funciones.Auxlogger, "Interfaz conectada: %s\n", interfazConectada.NombreInterfaz)

	//Guarda la interfazConectada en el map de interfaces conectadas
	funciones.InterfacesConectadas.Set(interfazConectada.NombreInterfaz, interfazConectada.Interfaz)
}

func handlerEjecutarInstruccionEnIO(w http.ResponseWriter, r *http.Request) {

	//--------- RECIBE ---------

	// Se decodifica el cuerpo de la solicitud en formato JSON
	var requestInstruccionIO structs.RequestEjecutarInstruccionIO
	marshalError := json.NewDecoder(r.Body).Decode(&requestInstruccionIO)
	if marshalError != nil {
		http.Error(w, marshalError.Error(), http.StatusBadRequest)
		return
	}

	//--------- EJECUTA ---------

	//--- VALIDA

	// Verifica que la Interfaz este Conectada
	interfazSolicitada, encontrado := funciones.InterfacesConectadas.Get(requestInstruccionIO.NombreInterfaz)
	if !encontrado {
		funciones.DesalojarProcesoIO(requestInstruccionIO.PidDesalojado)
		logueano.Mensaje(funciones.Auxlogger, "Interfaz no conectada.")
		http.Error(w, "Interfaz no conectada.", http.StatusNotFound)
		return
	}

	//Verifica que la instruccion sea compatible con el tipo de interfazConectada
	laInstruccionEsValida := funciones.ValidarInstruccionIO(interfazSolicitada.TipoInterfaz, requestInstruccionIO.Instruccion)
	if !laInstruccionEsValida {
		funciones.DesalojarProcesoIO(requestInstruccionIO.PidDesalojado)
		logueano.Mensaje(funciones.Auxlogger, "Instruccion incompatible.")
		http.Error(w, "Instruccion incompatible.", http.StatusBadRequest)
		return
	}

	//^log obligatorio (6/6)
	logueano.MotivoBloqueo(requestInstruccionIO.PidDesalojado, requestInstruccionIO.NombreInterfaz)

	//--- ENVIA A EJECUTAR A IO

	// Codifica instruccion a ejecutar en JSON
	body, marshalError := json.Marshal(requestInstruccionIO)
	if marshalError != nil {
		http.Error(w, marshalError.Error(), http.StatusInternalServerError)
		return
	}

	// Envía la instrucción a ejecutar a la interfazConectada (Puerto)
	query := interfazSolicitada.TipoInterfaz + "/" + requestInstruccionIO.Instruccion

	respuesta, err := config.Request(interfazSolicitada.PuertoInterfaz, interfazSolicitada.IpInterfaz, "POST", query, body)
	if err != nil {
		logueano.Error(funciones.Auxlogger, err)
		return
	}

	//Si es IO_STDIN_READ, devolver badRequest (implementado para los logueanos).
	if (requestInstruccionIO.Instruccion == "IO_STDIN_READ" || requestInstruccionIO.Instruccion == "IO_FS_READ") && respuesta.StatusCode == http.StatusBadRequest {

		//^log obligatorio (6/6)
		logueano.FinDeProceso(requestInstruccionIO.PidDesalojado, "INVALID_WRITE")
		http.Error(w, "INVALID_WRITE", http.StatusBadRequest)
		return
	}

	if respuesta.StatusCode != http.StatusOK {
		http.Error(w, "Error en la respuesta de I/O.", http.StatusInternalServerError)
		// Si no conecta con la interfaz, la elimina del map de las interfacesConectadas y desaloja el proceso.
		funciones.DesalojarProcesoIO(requestInstruccionIO.PidDesalojado)
		funciones.InterfacesConectadas.Delete(requestInstruccionIO.NombreInterfaz)
		logueano.Mensaje(funciones.Auxlogger, "Interfaz desconectada.")
		http.Error(w, "Interfaz desconectada.", http.StatusInternalServerError)
		return
	}

	//--- VUELVE DE IO

	// Pasa el proceso a READY y lo quita de la lista de bloqueados.
	pcbDesalojado, _ := funciones.MapBLOCK.Delete(requestInstruccionIO.PidDesalojado)

	//^ log obligatorio (2/6)
	logueano.CambioDeEstado(pcbDesalojado.Estado, "READY", pcbDesalojado.PID)
	pcbDesalojado.Estado = "READY"

	// Pasa el proceso a READY_PRIORITARIO si el algoritmo de planificacion es VRR
	if funciones.ConfigJson.Planning_Algorithm == "VRR" {
		pcbDesalojado.Estado = "READY_PRIORITARIO"
	}

	funciones.AdministrarQueues(pcbDesalojado)

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		//fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(bodyBytes)
}
