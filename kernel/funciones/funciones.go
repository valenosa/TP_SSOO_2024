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

//----------------------( VARIABLES )---------------------------\\

var ConfigJson config.Kernel

// ----------------------------Listas de Estados
var mx_NEW sync.Mutex
var listaNEW []structs.PCB

var mx_READY sync.Mutex
var listaREADY []structs.PCB

var mx_BLOCK sync.Mutex
var mapBLOK = make(map[uint32]structs.PCB)

var mx_EXIT sync.Mutex
var listaEXIT []structs.PCB

var procesoExec structs.PCB //Verificar que sea necesario

//---------------------------- Semaforos Sincronizacion

// Planificadores (Largo y Corto Plazo)
var Cont_producirPCB chan int
var Bin_hayPCBenREADY = make(chan int)
var mx_CPUOcupado sync.Mutex

var InterfacesConectadas = make(map[string]structs.Interfaz) //TODO: Debe tener mutex
var CounterPID uint32 = 0                                    //TODO: Debe tener mutex

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

	mx_BLOCK.Lock()
	pcbDesalojado := mapBLOK[pid]
	delete(mapBLOK, pid)
	mx_BLOCK.Unlock()

	pcbDesalojado.Estado = estado
	AdministrarQueues(pcbDesalojado)
	logueano.FinDeProceso(pcbDesalojado, estado)
}

//*======================================================| PLANIFICADOR |======================================================\\

// TODO: Reescribir par funcionamiento con semáforos (sincronización)  (18/5/24)
// Envía continuamente Procesos al CPU mientras que el bool planificadorActivo sea TRUE y el CPU esté esperando un structs.
func Planificador() {

	for {

		//Espero a que el CPU este libre
		mx_CPUOcupado.Lock()

		// Espero que exista PCB en READY
		<-Bin_hayPCBenREADY

		// Proceso READY -> EXEC
		var siguientePCB structs.PCB
		mx_READY.Lock()
		listaREADY, siguientePCB = dequeuePCB(listaREADY)
		mx_READY.Unlock()

		// ? Debería estar en dispatch?
		estadoAExec(&siguientePCB)
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
		mx_NEW.Lock()
		listaNEW = append(listaNEW, pcb)
		mx_NEW.Unlock()

	case "READY":

		//PCB --> cola de READY
		mx_READY.Lock()
		listaREADY = append(listaREADY, pcb)
		mx_READY.Unlock()

		//^ log obligatorio (3/6)
		logueano.PidsReady(listaREADY)

	case "BLOCK":

		//PCB --> mapa de BLOCK
		mx_BLOCK.Lock()
		mapBLOK[pcb.PID] = pcb
		mx_BLOCK.Unlock()

		//logPidsBlock(blockedMap)

	case "EXIT":

		//PCB --> cola de EXIT
		<-Cont_producirPCB
		mx_EXIT.Lock()
		listaEXIT = append(listaEXIT, pcb)
		mx_EXIT.Unlock()

	}
}

// ? ES NECESARIA ESTA FUNCION
// Desencola el PCB de la lista, si esta está vacía, simplemente espera nuevos Procesos, y avisa que la lista está vacía
func estadoAExec(pcb *structs.PCB) {

	// Cambia el estado del PCB a "EXEC"
	(*pcb).Estado = "EXEC"

	// Registra el proceso que está en ejecución
	procesoExec = *pcb
}

// TODO: Manejar el error en caso de que la lista esté vacía (18/5/24)
func dequeuePCB(listaPCB []structs.PCB) ([]structs.PCB, structs.PCB) {
	return listaPCB[1:], listaPCB[0]
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
