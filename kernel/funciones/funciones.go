package funciones

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/kernel/logueano"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

//----------------------( VARIABLES )---------------------------\\

var NewQueue []structs.PCB                                   //TODO: Debe tener mutex
var ReadyQueue []structs.PCB                                 //TODO: Debe tener mutex
var blockedMap = make(map[uint32]structs.PCB)                //TODO: Debe tener mutex
var exitQueue []structs.PCB                                  //TODO: Debe tener mutex
var procesoExec structs.PCB                                  //TODO: Debe tener mutex
var CPUOcupado bool = false                                  //TODO: Esto se hace con un sem binario
var planificadorActivo bool = true                           //TODO: Esto se hace con un sem binario
var InterfacesConectadas = make(map[string]structs.Interfaz) //TODO: Debe tener mutex
var readyQueueVacia bool = true                              //TODO: Esto se hace con un sem binario
var CounterPID uint32 = 0

var hayInterfaz = make(chan int)

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
	pcbDesalojado := blockedMap[pid]
	//TODO: Hacer wrapper de delete
	delete(blockedMap, pid)
	pcbDesalojado.Estado = estado
	administrarQueues(pcbDesalojado)
	logueano.FinDeProceso(pcbDesalojado, estado)
}

//*======================================================| PLANIFICADORES |======================================================\\

//*======================================[ PLANI LARGO PLAZO ]=======================================\\

//----------------------( CREAR PROCESOS )----------------------\\

func IniciarProceso(configJson config.Kernel, path string) {

	// Se crea un nuevo PCB en estado NEW
	var nuevoPCB structs.PCB
	nuevoPCB.PID = uint32(CounterPID)
	nuevoPCB.Estado = "NEW"

	// Incrementa el contador de Procesos
	CounterPID++

	// Codificar Body en un array de bytes (formato json)
	body, err := json.Marshal(structs.BodyIniciarProceso{
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
	var responseIniciarProceso structs.ResponseListarProceso

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
	asignarPCB(nuevoPCB, responseIniciarProceso)
}

// Asigna un PCB al Proceso recién creado y lo envía a la lista de READY para su ejecución
func asignarPCB(nuevoPCB structs.PCB, respuesta structs.ResponseListarProceso) {

	// Crea un nuevo PCB en base a un pid
	nuevoPCB.PID = uint32(respuesta.PID)

	// Almacena el estado viejo de un PCB
	pcb_estado_viejo := nuevoPCB.Estado
	nuevoPCB.Estado = "READY"

	//log obligatorio (2/6) (NEW->Ready): Cambio de Estado
	logueano.CambioDeEstado(pcb_estado_viejo, nuevoPCB)

	// Agrega el nuevo PCB a readyQueue
	administrarQueues(nuevoPCB)
}

//----------------------( FINALIZAR PROCESOS )----------------------\\

// Envía una solicitud a memoria para finalizar un proceso específico mediante su PID
func FinalizarProceso(configJson config.Kernel) {

	// PID del proceso a finalizar (hardcodeado).
	pid := 0

	// Enviar solicitud al servidor de memoria para finalizar el proceso.
	respuesta := config.Request(configJson.Port_Memory, configJson.Ip_Memory, "DELETE", fmt.Sprintf("process/%d", pid))

	// Verifica si ocurrió un error en la solicitud.
	if respuesta == nil {
		return
	}
}

//*=======================================[ PLANI CORTO PLAZO ]=======================================\\

// TODO: Reescribir par funcionamiento con semáforos (sincronización)  (18/5/24)
// Envía continuamente Procesos al CPU mientras que el bool planificadorActivo sea TRUE y el CPU esté esperando un structs.
func Planificador(configJson config.Kernel) {

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
		if len(ReadyQueue) == 0 {
			// Si la lista está vacía, se detiene el planificador.
			logueano.EsperaNuevosProcesos()
			readyQueueVacia = true
			planificadorActivo = false
			break
		}

		// Si la lista no está vacía, se envía el Proceso al CPU.
		// Se envía el primer Proceso y se hace un dequeue del mismo de la lista READY.
		var poppedPCB structs.PCB
		ReadyQueue, poppedPCB = dequeuePCB(ReadyQueue)

		// ? Debería estar en dispatch?
		estadoAExec(&poppedPCB)
		// ? Será siempre READY cuando pasa a EXEC?
		logueano.CambioDeEstado("READY", poppedPCB)

		// Se envía el proceso al CPU para su ejecución y se recibe la respuesta
		pcbActualizado, motivoDesalojo := dispatch(poppedPCB, configJson)

		// Se administra el PCB devuelto por el CPU
		administrarQueues(pcbActualizado)

		// TODO: Usar motivo de desalojo para algo.
		fmt.Println(motivoDesalojo)

	}
}

//----------------------( ADMINISTRAR COLAS LOCALES )----------------------\\

// Función que según que se haga con un PCB se lo puede enviar a la lista de planificación o a la de bloqueo
func administrarQueues(pcb structs.PCB) {

	switch pcb.Estado {
	case "NEW":

		// Agrega el PCB a la cola de nuevos procesos
		NewQueue = append(NewQueue, pcb)

	case "READY":

		// Agrega el PCB a la cola de procesos listos
		ReadyQueue = append(ReadyQueue, pcb)
		readyQueueVacia = false
		logueano.PidsReady(ReadyQueue)

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
