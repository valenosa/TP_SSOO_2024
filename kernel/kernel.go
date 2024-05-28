package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/kernel/funciones"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

//-------------------------- VARIABLES --------------------------

var newQueue []structs.PCB                                   //TODO: Debe tener mutex
var readyQueue []structs.PCB                                 //TODO: Debe tener mutex
var blockedMap = make(map[uint32]structs.PCB)                //TODO: Debe tener mutex
var exitQueue []structs.PCB                                  //TODO: Debe tener mutex
var procesoExec structs.PCB                                  //TODO: Debe tener mutex
var CPUOcupado bool = false                                  //TODO: Esto se hace con un sem binario
var planificadorActivo bool = true                           //TODO: Esto se hace con un sem binario
var interfacesConectadas = make(map[string]structs.Interfaz) //TODO: Debe tener mutex
var readyQueueVacia bool = true                              //TODO: Esto se hace con un sem binario
var counter int = 0

var hayInterfaz = make(chan int)

//================================| MAIN |================================\\

func main() {

	// Se declara una variable para almacenar la configuración del Kernel.
	var configJson config.Kernel

	// Extrae info de config.json
	config.Iniciar("config.json", &configJson)

	// Testea la conectividad con otros modulos
	//Conectividad(configJson)

	// Configura el logger
	config.Logger("Kernel.log")

	// ======== HandleFunctions ========
	http.HandleFunc("POST /interfazConectada", handlerIniciarInterfaz)
	http.HandleFunc("POST /instruccion", handlerInstrucciones)

	//inicio el servidor de Kern
	go config.IniciarServidor(configJson.Port)

	fmt.Printf("Antes del test")

	// Espera a que haya una interfaz conectada.
	<-hayInterfaz

	// Ahora que el servidor está en ejecución y hay una interfaz conectada, se puede iniciar el ciclo de instrucción.
	testCicloDeInstruccion(configJson)

	fmt.Printf("Despues del test")

}

//-------------------------- FUNCIONES ---------------------------------------------

// Recibe una interfazConectada y la agrega al map de interfaces conectadas.
func handlerIniciarInterfaz(w http.ResponseWriter, r *http.Request) {

	// Se crea una variable para almacenar la interfaz recibida en la solicitud.
	var requestInterfaz structs.RequestInterfaz

	// Se decodifica el cuerpo de la solicitud en formato JSON.
	err := json.NewDecoder(r.Body).Decode(&requestInterfaz)

	// Maneja el error en la decodificación.
	if err != nil {
		logErrorDecode()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Imprime la solicitud
	fmt.Println("Request path:", requestInterfaz)

	//Guarda la interfazConectada en la lista de interfaces conectadas.
	interfacesConectadas[requestInterfaz.NombreInterfaz] = requestInterfaz.Interfaz

	// Envía una señal al canal 'hayInterfaz' para indicar que hay una nueva interfaz conectada.
	hayInterfaz <- 0
}

// TODO: implementar para los demás tipos de interfaces (cambiar tipos de datos en request y body)
func handlerInstrucciones(w http.ResponseWriter, r *http.Request) {

	// Se crea una variable para almacenar la instrucción recibida en la solicitud.
	var request structs.InstruccionIO

	// Se decodifica el cuerpo de la solicitud en formato JSON.
	err := json.NewDecoder(r.Body).Decode(&request)

	// Maneja el error de la decodificación
	if err != nil {
		logErrorDecode()
		return
	}

	// Imprime la solicitud
	fmt.Println("Request path:", request)

	// Busca la interfaz conectada en el mapa de interfacesConectadas.
	interfazConectada, encontrado := interfacesConectadas[request.NombreInterfaz]

	// Si no se encontró la interfazConectada de la request, se desaloja el structs.
	if !encontrado {

		funciones.DesalojarProceso(request.PidDesalojado, "EXIT")
		fmt.Println("Interfaz no conectada.")
		return
	}

	//Verifica que la instruccion sea compatible con el tipo de interfazConectada.
	isValid := funciones.ValidarInstruccion(interfazConectada.TipoInterfaz, request.Instruccion)

	// Si la instrucción no es compatible, se desaloja el proceso y se marca como "EXIT".
	if !isValid {

		funciones.DesalojarProceso(request.PidDesalojado, "EXIT")
		fmt.Println("Interfaz incompatible.")
		return
	}

	// Agrega el Proceso a la cola de bloqueados de la interfazConectada.
	interfazConectada.QueueBlock = append(interfazConectada.QueueBlock, request.PidDesalojado)
	interfacesConectadas[request.NombreInterfaz] = interfazConectada

	// Prepara la interfazConectada para enviarla en el body.
	body, err := json.Marshal(request.UnitWorkTime)

	// Maneja los errores al crear el body.
	if err != nil {
		fmt.Printf("error codificando body: %s", err.Error())
		return
	}

	// Envía la instrucción a ejecutar a la interfazConectada (Puerto).
	respuesta := config.Request(interfazConectada.PuertoInterfaz, "localhost", "POST", request.Instruccion, body)

	// Verifica que no hubo error en la request
	if respuesta == nil {
		return
	}

	// Si la interfazConectada pudo ejecutar la instrucción, pasa el Proceso a READY.
	if respuesta.StatusCode == http.StatusOK {
		// Pasa el proceso a READY y lo quita de la lista de bloqueados.
		funciones.DesalojarProceso(request.PidDesalojado, "READY")
		return
	}
}

func validarInstruccion(tipo string, instruccion string) bool {
	//Verificar que esa interfazConectada puede ejecutar la instruccion que le pide el CPU
	switch tipo {
	case "GENERICA":
		return instruccion == "IO_GEN_SLEEP"
	}
	return false
}

// ! SE PUEDE BORRAR TODO ESTO, YA ESTÁ UTILIZANDO LOS PAQUETES "FUNCIONES" Y "LOG". LOS TESTS NO FUERON MIGRADOS.
// -------------------------- ADJACENT FUNCTIONS ------------------------------------
// Asigna un PCB recién creado a la lista de PCBs en estado READY.
func asignarPCBAReady(nuevoPCB structs.PCB, respuesta structs.ResponseIniciarProceso) {

	// Crea un nuevo PCB en base a un pid
	nuevoPCB.PID = uint32(respuesta.PID)

	// Almacena el estado viejo de un PCB
	pcb_estado_viejo := nuevoPCB.Estado
	nuevoPCB.Estado = "READY"

	//log obligatorio (2/6) (NEW->Ready): Cambio de Estado
	logCambioDeEstado(pcb_estado_viejo, nuevoPCB)

	// Agrega el nuevo PCB a readyQueue
	administrarQueues(nuevoPCB)
}

//-------------------------- API's --------------------------------------------------

func iniciarProceso(configJson config.Kernel, path string) {

	// Se crea un nuevo PCB en estado NEW
	var nuevoPCB structs.PCB
	nuevoPCB.PID = uint32(counter)
	nuevoPCB.Estado = "NEW"

	// Incrementa el contador de Procesos
	counter++

	// Codificar Body en un array de bytes (formato json)
	body, err := json.Marshal(structs.BodyIniciar{
		PID:  nuevoPCB.PID,
		Path: path,
	})

	// Maneja errores de codificación.
	if err != nil {
		fmt.Printf("error codificando body: %s", err.Error())
		return
	}

	//TODO: Quizá debería mandar el path a memoria solamente si hay "espacio" en la readyQueue (depende del grado de multiprogramación)
	// Enviar solicitud al servidor de memoria para almacenar el proceso.
	respuesta := config.Request(configJson.Port_Memory, configJson.Ip_Memory, "PUT", "process", body)
	// Verificar que no hubo error en la request
	if respuesta == nil {
		return
	}

	// Se declara una nueva variable que contendrá la respuesta del servidor.
	var responseIniciarProceso structs.ResponseIniciarProceso

	// Se decodifica la variable (codificada en formato json) en la estructura correspondiente.
	err = json.NewDecoder(respuesta.Body).Decode(&responseIniciarProceso)

	// Maneja errores para al decodificación.
	if err != nil {
		fmt.Printf("Error decodificando\n")
		return
	}

	//log obligatorio(1/6): creacion de Proceso
	//logNuevoProceso(nuevoPCB)

	// Asigna un PCB al proceso recién creado.
	asignarPCBAReady(nuevoPCB, responseIniciarProceso)
}

// Envía una solicitud a memoria para finalizar un proceso específico mediante su PID.
func finalizarProceso(configJson config.Kernel) {

	// PID del proceso a finalizar (hardcodeado).
	pid := 0

	// Enviar solicitud al servidor de memoria para finalizar el proceso.
	respuesta := config.Request(configJson.Port_Memory, configJson.Ip_Memory, "DELETE", fmt.Sprintf("process/%d", pid))

	// Verifica si ocurrió un error en la solicitud.
	if respuesta == nil {
		return
	}
}

// Envía una solicitud a memoria para obtener el estado de un proceso específico mediante su PID.
func estadoProceso(configJson config.Kernel) {

	// PID del proceso a consultar (hardcodeado).
	pid := 0

	// Enviar solicitud a memoria para obtener el estado del proceso.
	respuesta := config.Request(configJson.Port_Memory, configJson.Ip_Memory, "GET", fmt.Sprintf("process/%d", pid))

	// Verifica si ocurrió un error en la solicitud.
	if respuesta == nil {
		return
	}

	// Declarar una variable para almacenar la respuesta del servidor.
	var response structs.ResponseIniciarProceso

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

// TODO desarrollar la lectura de procesos creados. La función no está en uso. (27/05/24)
// Envía una solicitud al módulo de memoria para obtener y mostrar la lista de todos los procesos
func listarProceso(configJson config.Kernel) {

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

// TODO: La función no está en uso. (27/05/24)
// Envía una solicitud al módulo de CPU para iniciar el proceso de planificación.
func iniciarPlanificacion(configJson config.Kernel) {

	// Enviar solicitud al servidor de CPU para iniciar la planificación.
	respuesta := config.Request(configJson.Port_CPU, configJson.Ip_CPU, "PUT", "plani")

	// Verificar si ocurrió un error en la solicitud.
	if respuesta == nil {
		return
	}
}

// TODO: La función no está en uso. (27/05/24)
// Envía una solicitud al módulo de CPU para detener el proceso de planificación.
func detenerPlanificacion(configJson config.Kernel) {

	// Enviar solicitud al servidor de CPU para detener la planificación.
	respuesta := config.Request(configJson.Port_CPU, configJson.Ip_CPU, "DELETE", "plani")

	// Verificar si ocurrió un error en la solicitud.
	if respuesta == nil {
		return
	}
}

// Dispatch envía un PCB al CPU para su ejecución y maneja la respuesta del servidor CPU.
func dispatch(pcb structs.PCB, configJson config.Kernel) (structs.PCB, string) {

	//Envia PCB al CPU.
	fmt.Println("Se envió el proceso", pcb.PID, "al CPU")

	// Se realizan las acciones necesarias para la comunicación HTTP y la ejecución del proceso.
	CPUOcupado = true

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

	// Actualiza el estado del CPU.
	CPUOcupado = false

	fmt.Println("Exit queue:", exitQueue)

	// Retorna el PCB y el motivo de desalojo.
	return respuestaDispatch.PCB, respuestaDispatch.MotivoDeDesalojo
}

// TODO: La función no está en uso. (27/05/24)
// Envía una interrupción al ciclo de instrucción del CPU.
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

//-------------------------- PLANIFICACIÓN --------------------------------------------------

// Función que según que se haga con un PCB se lo puede enviar a la lista de planificación o a la de bloqueo
func administrarQueues(pcb structs.PCB) {

	switch pcb.Estado {
	case "NEW":

		// Agrega el PCB a la cola de nuevos procesos
		newQueue = append(newQueue, pcb)

	case "READY":

		// Agrega el PCB a la cola de procesos listos
		readyQueue = append(readyQueue, pcb)
		readyQueueVacia = false
		logPidsReady(readyQueue)

	//TODO: Deberia ser una por cada IO.
	case "BLOCK":

		// Agrega el PCB al mapa de procesos bloqueados
		blockedMap[pcb.PID] = pcb

		//TODO: Implementar log para el manejo de listas BLOCK con map
		//logPidsBlock(blockedMap)

	case "EXIT":

		// Agrega el PCB a la cola de procesos finalizados
		exitQueue = append(exitQueue, pcb)
		//TODO: momentaneamente sera un string constante, pero el motivo de Finalizacion deberá venir con el PCB (o alguna estructura que la contenga)
		//motivoDeFinalizacion := "SUCCESS"
		//logFinDeProceso(pcb, motivoDeFinalizacion)
	}
}

// Envía continuamente Procesos al CPU mientras que el bool planificadorActivo sea TRUE y el CPU esté esperando un structs.
func planificador(configJson config.Kernel) {

	// Verifica si el CPU no está ocupado y la lista de procesos listos no está vacía.
	if !CPUOcupado && !readyQueueVacia {
		planificadorActivo = true
	}
	for planificadorActivo {
		// Si el CPU está ocupado, detiene el planificador
		if CPUOcupado {
			planificadorActivo = false
			break
		}

		// Si la lista de procesos en READY está vacía, se detiene el planificador.
		if len(readyQueue) == 0 {
			// Si la lista está vacía, se detiene el planificador.
			logEsperaNuevosProcesos()
			readyQueueVacia = true
			planificadorActivo = false
			break
		}

		// Si la lista no está vacía, se envía el Proceso al CPU.
		// Se envía el primer Proceso y se hace un dequeue del mismo de la lista READY.
		var poppedPCB structs.PCB
		readyQueue, poppedPCB = dequeuePCB(readyQueue)

		// ? Debería estar en dispatch?
		estadoAExec(&poppedPCB)
		// ? Será siempre READY cuando pasa a EXEC?
		logCambioDeEstado("READY", poppedPCB)

		// Se envía el proceso al CPU para su ejecución y se recibe la respuesta
		pcbActualizado, motivoDesalojo := dispatch(poppedPCB, configJson)

		// Se actualizan las colas de procesos según la respuesta del CPU
		administrarQueues(pcbActualizado)

		// TODO: Usar motivo de desalojo para algo.
		fmt.Println(motivoDesalojo)

	}
}

// Desencola el PCB de la lista, si esta está vacía, simplemente espera nuevos Procesos, y avisa que la lista está vacía
func dequeuePCB(listaPCB []structs.PCB) ([]structs.PCB, structs.PCB) {
	//TODO: Manejar el error en caso de que la lista esté vacía.
	return listaPCB[1:], listaPCB[0]
}

func estadoAExec(pcb *structs.PCB) {

	// Cambia el estado del PCB a "EXEC"
	(*pcb).Estado = "EXEC"

	// Registra el proceso que está en ejecución
	procesoExec = *pcb
}

//-------------------------- TEST --------------------------------------------------
// !ESTO NO SE MIGRÓ A NINGÚN PAQUETE.
// Testea la conectividad con otros módulos

func testConectividad(configJson config.Kernel) {
	fmt.Println("\nIniciar Proceso:")
	funciones.IniciarProceso(configJson, "path")
	funciones.IniciarProceso(configJson, "path")
	funciones.IniciarProceso(configJson, "path")
	funciones.IniciarProceso(configJson, "path")
	fmt.Println("\nFinalizar Proceso:")
	funciones.FinalizarProceso(configJson)
	fmt.Println("\nEstado Proceso:")
	funciones.EstadoProceso(configJson)
	fmt.Println("\nListar Procesos:")
	funciones.ListarProceso(configJson)
	fmt.Println("\nDetener Planificación:")
	funciones.DetenerPlanificacion(configJson)
	fmt.Println("\nIniciar Planificación:")
	funciones.IniciarPlanificacion(configJson)
}

func testPlanificacion(configJson config.Kernel) {

	printList := func() {
		fmt.Println("readyQueue:")
		var ready []uint32
		for _, pcb := range readyQueue {
			ready = append(ready, pcb.PID)
		}
		fmt.Println(ready)
	}

	//
	fmt.Printf("\nSe crean 2 procesos-------------\n\n")
	for i := 0; i < 2; i++ {
		path := "procesos" + strconv.Itoa(counter) + ".txt"
		funciones.IniciarProceso(configJson, path)
	}

	fmt.Printf("\nSe testea el planificador-------------\n\n")
	funciones.Planificador(configJson)
	printList()

	fmt.Printf("\nSe crean 2 procesos-------------\n\n")
	for i := 0; i < 2; i++ {
		path := "proceso" + strconv.Itoa(counter) + ".txt"
		funciones.IniciarProceso(configJson, path)
	}
}

func testCicloDeInstruccion(configJson config.Kernel) {

	fmt.Printf("\nSe crean 1 proceso-------------\n\n")
	funciones.IniciarProceso(configJson, "proceso_test")

	fmt.Printf("\nSe testea el planificador-------------\n\n")
	funciones.Planificador(configJson)
}

// -------------------------- LOG's --------------------------------------------------
// !SE MIGRÓ AL PAQUETE "LOG"
// log obligatorio (1/6)
func logNuevoProceso(nuevoPCB structs.PCB) {

	log.Printf("Se crea el proceso %d en estado %s", nuevoPCB.PID, nuevoPCB.Estado)
}

// log obligatorio (2/6)
func logCambioDeEstado(pcb_estado_viejo string, pcb structs.PCB) {

	log.Printf("PID: %d - Estado anterior: %s - Estado actual: %s", pcb.PID, pcb_estado_viejo, pcb.Estado)

}

// log obligatorio (3/6)
func logPidsReady(readyQueue []structs.PCB) {
	var pids []uint32
	//Recorre la lista READY y guarda sus PIDs
	for _, pcb := range readyQueue {
		pids = append(pids, pcb.PID)
	}

	log.Printf("Cola Ready 'readyQueue' : %v", pids)
}

// log obligatorio (4/6)
func logFinDeProceso(pcb structs.PCB, motivoDeFinalizacion string) {

	log.Printf("Finaliza el proceso: %d - Motivo: %s", pcb.PID, motivoDeFinalizacion)

}

//LUEGO IMPLEMENTAR EN NUESTRO ARCHIVO NO OFICIAL DE LOGS ----------------------------

// TODO: Implementar para blockedMap.
func logPidsBlock(blockQueue []structs.PCB) {
	var pids []uint32
	//Recorre la lista BLOCK y guarda sus PIDs
	for _, pcb := range blockQueue {
		pids = append(pids, pcb.PID)
	}

	fmt.Printf("Cola Block 'blockQueue' : %v", pids)
}

// log para el manejo de listas EXEC
func logPidsExec(ExecQueue []structs.PCB) {
	var pids []uint32
	//Recorre la lista EXEC y guarda sus PIDs
	for _, pcb := range ExecQueue {
		pids = append(pids, pcb.PID)
	}

	fmt.Printf("Cola Executing 'ExecQueue' : %v", pids)
}

func logEsperaNuevosProcesos() {

	fmt.Println("Esperando nuevos procesos...")

}

func logErrorDecode() {

	fmt.Printf("Error al decodificar request body: ")
}
