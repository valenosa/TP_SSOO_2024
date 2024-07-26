package funciones

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-golang/utils/logueano"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// ----------------------( VARIABLES )---------------------------\\
var ConfigJson config.Kernel

var Auxlogger *log.Logger

// ---------------------------- Recursos
var MapRecursos = make(map[string]*structs.Recurso)

// ----------------------------Listas de Estados
var ListaREADY = ListaSegura{}
var MapBLOCK = MapSeguroPCB{m: make(map[uint32]structs.PCB)}

// Solo para mostrarlo en listarProcesos
var ListaNEW = ListaSegura{}
var ListaEXIT = ListaSegura{}
var ProcesoExec structs.PCB

// ---------------------------- VRR
var ListaREADY_PRIORITARIO = ListaSegura{}

// ---------------------------- Semáforos PLANIFICADORES

// Iniciar/Detener
var PlanificadorIniciado = false
var TogglePlanificador sync.Mutex

// Largo Plazo
var Cont_producirPCB chan int

// Corto Plazo
var Bin_hayPCBenREADY chan int //Se inicializa con el buffer = grado de multiprogramacion +1 ya que los semaforos en go tmb se bloquean si no pueden meter algo en el buffer
var mx_CPUOcupado sync.Mutex

var InterfacesConectadas = MapSeguroInterfaz{m: make(map[string]structs.Interfaz)}

//*=======================================| PLANIFICADOR |=======================================\\

// Envía continuamente Procesos al CPU mientras que el bool planificadorActivo sea TRUE y el CPU esté esperando un structs.
func Planificador() {

	//Espero a que se active el planificador
	for {

		//Espero a que el CPU este libre
		mx_CPUOcupado.Lock()

		// Espero que exista PCB en READY (Tanto en READY como en READY_PRIORITARIO)
		<-Bin_hayPCBenREADY

		TogglePlanificador.Lock()

		var siguientePCB structs.PCB      // PCB a enviar al CPU
		var tiempoInicioQuantum time.Time // Tiempo de inicio del Quantum

		//Ejecuta VRR si hay procesos priotirarios.
		if strings.ToUpper(ConfigJson.Planning_Algorithm) == "VRR" && len(ListaREADY_PRIORITARIO.List) > 0 {

			siguientePCB = ListaREADY_PRIORITARIO.Dequeue()
			go roundRobin(siguientePCB.PID, int(siguientePCB.Quantum))

		} else {
			siguientePCB = ListaREADY.Dequeue()

			//Si el algoritmo de planificación es Round Robin, "contabiliza" el quantum
			if strings.ToUpper(ConfigJson.Planning_Algorithm) != "FIFO" {
				go roundRobin(siguientePCB.PID, ConfigJson.Quantum)
			}
		}

		//Guardo tiempo de inicio para Virtual RR
		tiempoInicioQuantum = time.Now()

		// Proceso READY -> EXEC
		siguientePCB.Estado = "EXEC"
		ProcesoExec = siguientePCB

		//^ log obligatorio (2/6)
		logueano.CambioDeEstado("READY", siguientePCB.Estado, siguientePCB.PID)

		// Se envía el proceso al CPU para su ejecución y espera a que se lo devuelva actualizado
		pcbActualizado, motivoDesalojo := dispatch(siguientePCB, ConfigJson)

		ProcesoExec = structs.PCB{} //Limpia a proceso Exec para listarProcesos

		// Si se usa VRR y el proceso se desalojo por IO se guarda el Quantum no usado por el proceso
		if ConfigJson.Planning_Algorithm == "VRR" && motivoDesalojo == "IO" {

			tiempoCorteQuantum := time.Now()

			tiempoUsado := tiempoCorteQuantum.Sub(tiempoInicioQuantum)

			pcbActualizado.Quantum = uint16(ConfigJson.Quantum) - uint16(tiempoUsado.Milliseconds())
		}

		TogglePlanificador.Unlock()

		//Aviso que esta libre el CPU
		mx_CPUOcupado.Unlock()

		administrarMotivoDesalojo(&pcbActualizado, motivoDesalojo)

		// Se administra el PCB devuelto por el CPU
		AdministrarQueues(pcbActualizado)

	}
}

func administrarMotivoDesalojo(pcb *structs.PCB, motivoDesalojo string) {

	switch motivoDesalojo {

	case "Fin de QUANTUM":

		//^ log obligatorio (5/6)
		logueano.FinDeQuantum(*pcb)

		//^ log obligatorio (2/6)
		logueano.CambioDeEstado(pcb.Estado, "READY", pcb.PID)
		pcb.Estado = "READY"

	case "IO", "WAIT":

		//^ log obligatorio (2/6)
		logueano.CambioDeEstado(pcb.Estado, "BLOCK", pcb.PID)
		pcb.Estado = "BLOCK"

	case "INTERRUPTED_BY_USER", "INVALID_RESOURCE", "OUT_OF_MEMORY", "PAGE_FAULT", "INVALID_WRITE", "INVALID_READ", "SUCCESS":

		//^ log obligatorio (2/6)
		logueano.CambioDeEstado(pcb.Estado, "EXIT", pcb.PID)

		//^ log obligatorio (2/6)
		logueano.FinDeProceso(pcb.PID, motivoDesalojo)

		pcb.Estado = "EXIT"

	}
}

//----------------------( ROUND ROBIN )----------------------\\

func roundRobin(PID uint32, quantum int) {
	time.Sleep(time.Duration(quantum) * time.Millisecond)
	Interrupt(PID, "Fin de QUANTUM")
}

//----------------------( ADMINISTRAR COLAS )----------------------\\

// Administra las colas de los procesos según el estado indicado en el PCB
func AdministrarQueues(pcb structs.PCB) {

	switch pcb.Estado {
	case "NEW":

		//^ log obligatorio (1/6)
		logueano.NuevoProceso(pcb)

		//PCB --> cola de NEW
		ListaNEW.Append(pcb)

		//logueano.PidsNew(Auxlogger, ListaNEW.List)

	case "READY":

		//PCB --> cola de READY
		ListaREADY.Append(pcb)

		//Avisa al planificador que hay un PCB en READY (se usa dentro del select para que no se bloquee si ya metieron algo "buffer infinito")
		Bin_hayPCBenREADY <- 0

		//^ log obligatorio (3/6)
		logueano.PidsReady(ListaREADY.List)

	case "READY_PRIORITARIO":
		ListaREADY_PRIORITARIO.Append(pcb)
		logueano.PidsReadyPrioritarios(Auxlogger, pcb)
		Bin_hayPCBenREADY <- 0

	case "BLOCK":

		//PCB --> mapa de BLOCK
		MapBLOCK.Set(pcb.PID, pcb)

		//logPidsBlock(blockedMap)
		//logueano.PidsBlock(Auxlogger, MapBLOCK.m)

	case "EXIT":

		//PCB --> cola de EXIT
		ListaEXIT.Append(pcb)
		//logueano.PidsExit(Auxlogger, ListaEXIT.List)
		LiberarProceso(pcb)
		<-Cont_producirPCB
	}
}

func LiberarProceso(pcb structs.PCB) {

	//-------------- Libera las estructuras de Memoria --------------
	// Crea un cliente HTTP
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/process", ConfigJson.Ip_Memory, ConfigJson.Port_Memory)

	// Crea una nueva solicitud PUT
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	// Agrega el PID y la PAGINA como params
	q := req.URL.Query()
	q.Add("pid", fmt.Sprint(pcb.PID))
	req.URL.RawQuery = q.Encode()

	// Realiza la solicitud al servidor de memoria
	_, err = cliente.Do(req)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	//-------------- Liberar Recursos ------------------------------

	for _, recurso := range pcb.Recursos {
		LiberarRecurso(recurso)
	}
}

func ExtraerPCB(pid uint32) (structs.PCB, bool) {

	pcb, encontrado := MapBLOCK.Delete(pid)
	if encontrado {
		//Retornar PCB
		return pcb, true
	}

	pcb, encontrado = ListaREADY.Extract(pid)
	if encontrado {
		//Retornar PCB y declarar que hay un proceso menos en READY
		<-Bin_hayPCBenREADY
		return pcb, true
	}

	pcb, encontrado = ListaNEW.Extract(pid)
	if encontrado {
		//Retornar PCB
		return pcb, true
	}

	return structs.PCB{}, false
}

func BuscarPCB(pid uint32) (structs.PCB, bool) {

	pcb, encontrado := MapBLOCK.Get(pid)
	if encontrado {
		//Retornar PCB
		return pcb, true
	}

	pcb, encontrado = ListaREADY.Search(pid)
	if encontrado {
		//Retornar PCB
		return pcb, true
	}

	pcb, encontrado = ListaNEW.Search(pid)
	if encontrado {
		//Retornar PCB
		return pcb, true
	}

	if ProcesoExec.PID == pid {
		//Eliminar proceso exec

		//Retornar PCB
		return ProcesoExec, true
	}

	return structs.PCB{}, false
}

//----------------------( EJECUTAR PROCESOS EN CPU )----------------------\\

// Envía un PCB (indicado por el planificador) al CPU para su ejecución, Tras volver lo devuelve al planificador
func dispatch(pcb structs.PCB, configJson config.Kernel) (structs.PCB, string) {

	//-------------------Request al CPU------------------------

	// Codifica el cuerpo en un arreglo de bytes (formato JSON).
	body, err := json.Marshal(pcb)
	if err != nil {
		logueano.MensajeConFormato(Auxlogger, "error codificando body: %s", err.Error())
		return structs.PCB{}, "ERROR"
	}

	// Envía una solicitud al servidor CPU.
	respuesta, err := config.Request(configJson.Port_CPU, configJson.Ip_CPU, "POST", "exec", body)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return structs.PCB{}, "ERROR"
	}

	// Se declara una nueva variable que contendrá la respuesta del servidor
	var respuestaDispatch structs.RespuestaDispatch

	// Se decodifica la variable (codificada en formato JSON) en la estructura correspondiente
	err = json.NewDecoder(respuesta.Body).Decode(&respuestaDispatch)

	// Maneja los errores para la decodificación
	if err != nil {
		logueano.Mensaje(Auxlogger, "Error decodificando respuesta del CPU.")
		return structs.PCB{}, "ERROR"
	}

	//-------------------Fin Request al CPU------------------------

	// Retorna el PCB y el motivo de desalojo.
	return respuestaDispatch.PCB, respuestaDispatch.MotivoDeDesalojo
}

// Desaloja el Proceso enviando una interrupción al CPU
func Interrupt(PID uint32, tipoDeInterrupcion string) {
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/interrupciones", ConfigJson.Ip_CPU, ConfigJson.Port_CPU)
	req, err := http.NewRequest("POST", url, nil)

	if err != nil {
		return
	}

	// Convierte el PID a string
	pidString := strconv.FormatUint(uint64(PID), 10)

	// Agrega el PID y el tipo de interrupción como parámetros de la URL
	q := req.URL.Query()
	q.Add("PID", string(pidString))
	q.Add("interrupt_type", tipoDeInterrupcion)

	req.URL.RawQuery = q.Encode()

	// Envía la solicitud con el PID y el tipo de interrupción
	req.Header.Set("Content-Type", "text/plain")
	respuesta, err := cliente.Do(req)

	// Verifica si hubo un error al enviar la solicitud
	if err != nil {
		logueano.Mensaje(Auxlogger, "Error al enviar la interrupción al CPU.")
		return
	}

	// Verifica si hubo un error en la respuesta
	if respuesta.StatusCode != http.StatusOK {
		logueano.Mensaje(Auxlogger, "Error en la respuesta del CPU.")
		return
	}
}

//*======================================| ENTRADA SALIDA (I/O) |=======================================\\

// Verificar que esa interfazConectada puede ejecutar la instruccion que le pide el CPU
func ValidarInstruccionIO(tipo string, instruccion string) bool {
	switch tipo {
	case "GENERICA":
		return instruccion == "IO_GEN_SLEEP"

	case "STDIN":
		return instruccion == "IO_STDIN_READ"

	case "STDOUT":
		return instruccion == "IO_STDOUT_WRITE"

	case "DIALFS":
		return instruccion == "IO_FS_CREATE" || instruccion == "IO_FS_READ" || instruccion == "IO_FS_WRITE" || instruccion == "IO_FS_DELETE" || instruccion == "IO_FS_TRUNCATE"
	}
	return false
}

// Toma un pid del map general de BLOCK y manda un proceso a EXIT.
func DesalojarProcesoIO(pid uint32) {
	pcbDesalojado, _ := MapBLOCK.Delete(pid)
	pcbDesalojado.Estado = "EXIT"

	//^ log obligatorio (2/6)
	logueano.CambioDeEstado("BLOCK", pcbDesalojado.Estado, pcbDesalojado.PID)
	AdministrarQueues(pcbDesalojado)
}

// *=======================================| RECURSOS |=======================================\\

func LeerRecursos(recursos []string, instancia_recursos []int) {
	//Tomo de resources y resource_instances los recursos y sus instancias y los guardo en Recursos
	for i, recurso := range recursos {
		MapRecursos[recurso] = &structs.Recurso{Instancias: instancia_recursos[i]}
	}
}

func LiberarRecurso(nombreRecurso string) {

	recurso := MapRecursos[nombreRecurso]

	// Si hay procesos bloqueados por el recurso, se desbloquea al primero
	if len(recurso.ListaBlock.List) > 0 {

		// Tomo el primer PID de la lista de BLOCK (del recurso)
		pid := recurso.ListaBlock.Dequeue()

		pcbDesbloqueado, find := MapBLOCK.Delete(pid)
		if !find {
			LiberarRecurso(nombreRecurso)
			return
		}

		//^ log obligatorio (2/6)
		logueano.CambioDeEstado(pcbDesbloqueado.Estado, "READY", pcbDesbloqueado.PID)

		//Se agrega el recurso a la lista de recursos del proceso
		pcbDesbloqueado.Recursos = append(pcbDesbloqueado.Recursos, nombreRecurso)
		//Se pasa el proceso a de BLOCK -> READY
		pcbDesbloqueado.Estado = "READY"
		AdministrarQueues(pcbDesbloqueado)

	} else {
		recurso.Instancias++
	}
}

// *=======================================| TADs SINCRONIZACION |=======================================\\

// ----------------------( LISTA )----------------------\\
type ListaSegura struct {
	Mx   sync.Mutex
	List []structs.PCB
}

func (sList *ListaSegura) Append(value structs.PCB) {
	sList.Mx.Lock()
	sList.List = append(sList.List, value)
	sList.Mx.Unlock()
}

func (sList *ListaSegura) Dequeue() structs.PCB {
	sList.Mx.Lock()
	var pcb = sList.List[0]
	sList.List = sList.List[1:]
	sList.Mx.Unlock()

	return pcb
}

func AppendListaProceso(listadoProcesos []structs.ResponseListarProceso, listaEspecifica *ListaSegura) []structs.ResponseListarProceso {
	listaEspecifica.Mx.Lock()
	for i := range listaEspecifica.List {
		elemento := structs.ResponseListarProceso{PID: listaEspecifica.List[i].PID, Estado: listaEspecifica.List[i].Estado}
		listadoProcesos = append(listadoProcesos, elemento)
	}
	listaEspecifica.Mx.Unlock()

	return listadoProcesos
}

func AppendMapProceso(listadoProcesos []structs.ResponseListarProceso, mapEspecifico *MapSeguroPCB) []structs.ResponseListarProceso {
	mapEspecifico.mx.Lock()
	for _, value := range mapEspecifico.m {
		elemento := structs.ResponseListarProceso{PID: value.PID, Estado: value.Estado}
		listadoProcesos = append(listadoProcesos, elemento)
	}
	mapEspecifico.mx.Unlock()

	return listadoProcesos
}

// Busca un PCB en la lista segura y lo elimina si lo encuentra
func (sList *ListaSegura) Extract(pcbID uint32) (structs.PCB, bool) {
	sList.Mx.Lock()
	defer sList.Mx.Unlock()

	for i, pcb := range sList.List {
		if pcb.PID == pcbID {
			// Elimina el PCB de la lista
			sList.List = append(sList.List[:i], sList.List[i+1:]...)
			return pcb, true
		}
	}

	return structs.PCB{}, false
}

func (sList *ListaSegura) Search(pcbID uint32) (structs.PCB, bool) {
	sList.Mx.Lock()
	defer sList.Mx.Unlock()

	for _, pcb := range sList.List {
		if pcb.PID == pcbID {
			return pcb, true
		}
	}

	return structs.PCB{}, false
}

// ----------------------( MAP PCB )----------------------\\
type MapSeguroPCB struct {
	mx sync.Mutex
	m  map[uint32]structs.PCB
}

func (sMap *MapSeguroPCB) Set(key uint32, value structs.PCB) {
	sMap.mx.Lock()
	sMap.m[key] = value
	sMap.mx.Unlock()
}

func (sMap *MapSeguroPCB) Delete(key uint32) (structs.PCB, bool) {
	sMap.mx.Lock()
	var pcb, find = sMap.m[key]
	if find {
		delete(sMap.m, key)
	}
	sMap.mx.Unlock()

	return pcb, find
}

func (sMap *MapSeguroPCB) Get(key uint32) (structs.PCB, bool) {
	sMap.mx.Lock()
	var pcb, find = sMap.m[key]
	sMap.mx.Unlock()

	return pcb, find
}

// ----------------------( MAP Interfaz )----------------------\\
type MapSeguroInterfaz struct {
	mx sync.Mutex
	m  map[string]structs.Interfaz
}

func (sMap *MapSeguroInterfaz) Set(key string, value structs.Interfaz) {
	sMap.mx.Lock()
	sMap.m[key] = value
	sMap.mx.Unlock()
}

func (sMap *MapSeguroInterfaz) Delete(key string) structs.Interfaz {
	sMap.mx.Lock()
	var pcb = sMap.m[key]
	delete(sMap.m, key)
	sMap.mx.Unlock()

	return pcb
}

func (sMap *MapSeguroInterfaz) Get(key string) (structs.Interfaz, bool) {
	sMap.mx.Lock()
	var interfaz, find = sMap.m[key]
	sMap.mx.Unlock()

	return interfaz, find
}
