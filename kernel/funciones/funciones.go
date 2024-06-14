package funciones

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-golang/kernel/logueano"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// ----------------------( VARIABLES )---------------------------\\
var ConfigJson config.Kernel

// ---------------------------- Recursos
var MapRecursos = make(map[string]*structs.Recurso)

// ----------------------------Listas de Estados
var ListaNEW = ListaSegura{}
var ListaREADY = ListaSegura{}
var ListaEXIT = ListaSegura{}
var MapBLOCK = MapSeguroPCB{m: make(map[uint32]structs.PCB)}
var ProcesoExec structs.PCB

//var procesoExec structs.PCB //* Verificar que sea necesario

// ---------------------------- Semaforos PLANIFICADORES

// Iniciar/Detener
var OnePlani sync.Mutex
var TogglePlanificador bool

// Largo Plazo
var Cont_producirPCB chan int

// Corto Plazo
var Bin_hayPCBenREADY chan int //Se inicializa con el buffer = grado de multiprogramacion +1 ya que los semaforos en go tmb se bloquean si no pueden meter algo en el buffer
var mx_CPUOcupado sync.Mutex

var InterfacesConectadas = MapSeguroInterfaz{m: make(map[string]structs.Interfaz)}
var Mx_ConterPID sync.Mutex
var CounterPID uint32 = 0

//var hayInterfaz = make(chan int)

// Envía una solicitud a memoria para obtener el estado de un proceso específico mediante su PID.
func EstadoProceso(configJson config.Kernel) {

	// PID del proceso a consultar (hardcodeado).
	pid := 0

	// Enviar solicitud a memoria para obtener el estado del proceso.
	respuesta := config.Request(configJson.Port_Memory, configJson.Ip_Memory, "GET", fmt.Sprintf("process/%d", pid))
	if respuesta == nil {
		return
	}

	// Declarar una variable para almacenar la respuesta del servidor.
	var response structs.ResponseListarProceso

	// Decodifica la respuesta del servidor.
	err := json.NewDecoder(respuesta.Body).Decode(&response)

	// Maneja el error para la decodificación.
	if err != nil {
		fmt.Printf("Error decodificando\n")
		fmt.Println(err)
		return
	}

	// Imprimir información sobre el proceso (en este caso, solo el PID).
	fmt.Println(response)
}

// TODO desarrollar la lectura de procesos creados (27/05/24)
// Envía una solicitud al módulo de memoria para obtener y mostrar la lista de todos los procesos
func ListarProceso(configJson config.Kernel) {

	// Enviar solicitud al servidor de memoria
	respuesta := config.Request(configJson.Port_Memory, configJson.Ip_Memory, "GET", "process")
	if respuesta == nil {
		return
	}

	// TODO: Checkear que io.ReadAll no esté deprecada.(27/05/24)
	// Leer el cuerpo de la respuesta.
	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return
	}

	// Imprimir la lista de procesos.
	fmt.Println(string(bodyBytes))
}

//*=======================================| PLANIFICADOR |=======================================\\

// TODO: Verificar el tema del semaforo de hay pcb en ready (31/05/24)
// Envía continuamente Procesos al CPU mientras que el bool planificadorActivo sea TRUE y el CPU esté esperando un structs.
func Planificador() {

	//Espero a que se active el planificador
	for TogglePlanificador {

		//Espero a que el CPU este libre
		mx_CPUOcupado.Lock()

		// Espero que exista PCB en READY
		<-Bin_hayPCBenREADY

		// Proceso READY -> EXEC
		var siguientePCB = ListaREADY.Dequeue()
		siguientePCB.Estado = "EXEC"

		ProcesoExec = siguientePCB

		logueano.CambioDeEstado("READY", siguientePCB)

		// Se envía el proceso al CPU para su ejecución y se recibe la respuesta
		pcbActualizado, motivoDesalojo := dispatch(siguientePCB, ConfigJson)

		fmt.Println("Recursos Retenidos por", pcbActualizado.PID, ": ", pcbActualizado.Recursos) //! Borrar despues

		//Aviso que esta libre el CPU
		mx_CPUOcupado.Unlock()

		administrarInterrupciones(&pcbActualizado, motivoDesalojo)

		// Se administra el PCB devuelto por el CPU
		AdministrarQueues(pcbActualizado)

	}
}

func administrarInterrupciones(pcb *structs.PCB, motivoDesalojo string) {
	switch motivoDesalojo {
	case "Fin de QUANTUM":
		pcb.Estado = "READY"

	case "Finalizar PROCESO":
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

		//PCB --> cola de NEW
		ListaNEW.Append(pcb)

	case "READY":

		//PCB --> cola de READY
		ListaREADY.Append(pcb)

		//Avisa al planificador que hay un PCB en READY (se usa dentro del select para que no se bloquee si ya metieron algo "buffer infinito")
		Bin_hayPCBenREADY <- 0

		//^ log obligatorio (3/6)
		logueano.PidsReady(ListaREADY.List) //!No se si tengo que sync esto

	case "BLOCK":

		//PCB --> mapa de BLOCK
		MapBLOCK.Set(pcb.PID, pcb)

		//logPidsBlock(blockedMap)

	case "EXIT":

		//PCB --> cola de EXIT
		ListaEXIT.Append(pcb)
		<-Cont_producirPCB

	}
}

//----------------------( EJECUTAR PROCESOS EN CPU )----------------------\\

// TODO: Reescribir par funcionamiento con semáforos (sincronización)  (18/5/24)
// Envía un PCB al CPU para su ejecución, Tras volver lo manda a la cola correspondiente
func dispatch(pcb structs.PCB, configJson config.Kernel) (structs.PCB, string) {

	//Envia PCB al CPU.
	fmt.Println("Se envió el proceso", pcb.PID, "al CPU")

	//-------------------Request al CPU------------------------

	// Codifica el cuerpo en un arreglo de bytes (formato JSON).
	body, err := json.Marshal(pcb)

	// Maneja los errores para la codificación.
	if err != nil {
		fmt.Printf("error codificando body: %s", err.Error())
		return structs.PCB{}, "ERROR"
	}

	/*
		*Si el algoritmo de planificación es Round Robin, "contabiliza" el quantum
		?Es correcto?
	*/
	if configJson.Planning_Algorithm == "RR" {
		go roundRobin(pcb.PID, configJson.Quantum)
	}

	// Envía una solicitud al servidor CPU.
	respuesta := config.Request(configJson.Port_CPU, configJson.Ip_CPU, "POST", "exec", body)
	if respuesta == nil {
		return structs.PCB{}, "ERROR"
	}

	// Se declara una nueva variable que contendrá la respuesta del servidor
	var respuestaDispatch structs.RespuestaDispatch

	// Se decodifica la variable (codificada en formato JSON) en la estructura correspondiente
	err = json.NewDecoder(respuesta.Body).Decode(&respuestaDispatch)

	// Maneja los errores para la decodificación
	if err != nil {
		fmt.Printf("Error decodificando\n")
		return structs.PCB{}, "ERROR"
	}

	//-------------------Fin Request al CPU------------------------

	// Imprime el motivo de desalojo.
	fmt.Println("Proceso", respuestaDispatch.PCB.PID, "desalojado por:", respuestaDispatch.MotivoDeDesalojo)

	// Retorna el PCB y el motivo de desalojo.
	return respuestaDispatch.PCB, respuestaDispatch.MotivoDeDesalojo
}

// TODO: La función no está en uso. (27/05/24)
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
		fmt.Println("Error al enviar la interrupción a CPU.")
		return
	}

	// Verifica si hubo un error en la respuesta
	if respuesta.StatusCode != http.StatusOK {
		fmt.Println("Error al interpretar el motivo de desalojo.")
		return
	}

	fmt.Printf("Interrupción tipo %s enviada correctamente.\n", tipoDeInterrupcion)
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
	}
	return false
}

// Toma un pid del map general de BLOCK y manda un proceso a EXIT.
func DesalojarProcesoIO(pid uint32) {
	pcbDesalojado := MapBLOCK.Delete(pid)
	pcbDesalojado.Estado = "EXIT"
	AdministrarQueues(pcbDesalojado)
}

// *=======================================| RECURSOS |=======================================\\

func LeerRecursos(recursos []string, instancia_recursos []int) {
	//Tomo de resources y resource_instances los recursos y sus instancias y los guardo en Recursos
	for i, recurso := range recursos {
		MapRecursos[recurso] = &structs.Recurso{Instancias: instancia_recursos[i]}
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

// TODO: Manejar el error en caso de que la lista esté vacía (18/5/24)
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

func (sMap *MapSeguroPCB) Delete(key uint32) structs.PCB {
	sMap.mx.Lock()
	var pcb = sMap.m[key]
	delete(sMap.m, key)
	sMap.mx.Unlock()

	return pcb
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
