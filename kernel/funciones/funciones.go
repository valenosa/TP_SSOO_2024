package funciones

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/sisoputnfrba/tp-golang/kernel/logueano"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// ----------------------( VARIABLES )---------------------------\\
var ConfigJson config.Kernel

// ----------------------------Listas de Estados
var listaNEW = ListaSegura{}
var listaREADY = ListaSegura{}
var listaEXIT = ListaSegura{}
var mapBLOK = MapSeguroPCB{m: make(map[uint32]structs.PCB)}

//var procesoExec structs.PCB //* Verificar que sea necesario

// ---------------------------- Semaforos PLANIFICADORES

// Iniciar/Detener
var Bin_togglePlanificador = make(chan int, 2) //TODO: implementar

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

	// Verifica si ocurrió un error en la solicitud.
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

	// Verificar si ocurrió un error en la solicitud.
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

//*======================================| ENTRADA SALIDA (I/O) |=======================================\\

// Verificar que esa interfazConectada puede ejecutar la instruccion que le pide el CPU
func ValidarInstruccion(tipo string, instruccion string) bool {
	switch tipo {
	case "GENERICA":
		return instruccion == "IO_GEN_SLEEP"
	}
	return false
}

// Cambia el estado del PCB y lo envía a encolar segun el nuevo estado.
func DesalojarProceso(pid uint32, estado string) {

	pcbDesalojado := mapBLOK.Delete(pid)

	pcbDesalojado.Estado = estado
	AdministrarQueues(pcbDesalojado)
	logueano.FinDeProceso(pcbDesalojado, estado)
}

//*=======================================| PLANIFICADOR |=======================================\\

// TODO: Verificar el tema del semaforo de hay pcb en ready (31/05/24)
// Envía continuamente Procesos al CPU mientras que el bool planificadorActivo sea TRUE y el CPU esté esperando un structs.
func Planificador() {

	for {
		//Espero a que se active el planificador

		//Espero a que el CPU este libre
		mx_CPUOcupado.Lock()

		// Espero que exista PCB en READY
		<-Bin_hayPCBenREADY

		// Proceso READY -> EXEC
		var siguientePCB = listaREADY.Dequeue()
		siguientePCB.Estado = "EXEC"

		logueano.CambioDeEstado("READY", siguientePCB)

		// Se envía el proceso al CPU para su ejecución y se recibe la respuesta
		pcbActualizado, motivoDesalojo := dispatch(siguientePCB, ConfigJson)

		//Aviso que esta libre el CPU
		mx_CPUOcupado.Unlock()

		// Se administra el PCB devuelto por el CPU
		AdministrarQueues(pcbActualizado)

		// TODO: Usar motivo de desalojo para algo.
		fmt.Println(motivoDesalojo)

	}
}

//----------------------( ADMINISTRAR COLAS )----------------------\\

// Administra las colas de los procesos según el estado indicado en el PCB
func AdministrarQueues(pcb structs.PCB) {

	switch pcb.Estado {
	case "NEW":
		//PCB --> cola de NEW
		listaNEW.Append(pcb)

	case "READY":

		//PCB --> cola de READY
		listaREADY.Append(pcb)

		//Avisa al planificador que hay un PCB en READY (se usa dentro del select para que no se bloquee si ya metieron algo "buffer infinito")
		Bin_hayPCBenREADY <- 0

		//^ log obligatorio (3/6)
		logueano.PidsReady(listaREADY.list) //!No se si tengo que sync esto

	case "BLOCK":

		//PCB --> mapa de BLOCK
		mapBLOK.Set(pcb.PID, pcb)

		//logPidsBlock(blockedMap)

	case "EXIT":

		//PCB --> cola de EXIT
		listaEXIT.Append(pcb)
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

	// Envía una solicitud al servidor CPU.
	respuesta := config.Request(configJson.Port_CPU, configJson.Ip_CPU, "POST", "exec", body)

	// Verifica si hubo un error en la solicitud.
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
	fmt.Println("Motivo de desalojo:", respuestaDispatch.MotivoDeDesalojo)

	// Retorna el PCB y el motivo de desalojo.
	return respuestaDispatch.PCB, respuestaDispatch.MotivoDeDesalojo
}

// TODO: La función no está en uso. (27/05/24)
// Desaloja el Proceso enviando una interrupción al CPU
func interrupt(pid int, tipoDeInterrupcion string, configJson config.Kernel) {

	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/interrupciones", configJson.Ip_CPU, configJson.Port_CPU)
	req, err := http.NewRequest("POST", url, nil)

	if err != nil {
		return
	}

	// Convierte el PID a string
	pidString := strconv.Itoa(pid)

	// Agrega el PID y el tipo de interrupción como parámetros de la URL
	q := req.URL.Query()
	q.Add("pid", string(pidString))
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

	fmt.Println("Interrupción enviada correctamente.")
}

// *=======================================| TADs SINCRONIZACION |=======================================\\

// ----------------------( LISTA )----------------------\\
type ListaSegura struct {
	mx   sync.Mutex
	list []structs.PCB
}

func (sList *ListaSegura) Append(value structs.PCB) {
	sList.mx.Lock()
	sList.list = append(sList.list, value)
	sList.mx.Unlock()
}

// TODO: Manejar el error en caso de que la lista esté vacía (18/5/24)
func (sList *ListaSegura) Dequeue() structs.PCB {
	sList.mx.Lock()
	var pcb = sList.list[0]
	sList.list = sList.list[1:]
	sList.mx.Unlock()

	return pcb
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
	var pcb, find = sMap.m[key]
	sMap.mx.Unlock()

	return pcb, find
}
