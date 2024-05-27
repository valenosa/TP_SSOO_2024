package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/utils/structs"

	"github.com/sisoputnfrba/tp-golang/utils/config"
)

//================================| MAIN |===================================================\\

//-------------------------- STRUCTS -----------------------------------------

type RequestInterfaz struct {
	NombreInterfaz string
	Interfaz       Interfaz
}

// MOVELO A UTILS (struct tambien usada por entrasalida.go)
type Interfaz struct {
	TipoInterfaz   string
	PuertoInterfaz int
	QueueBlock     []uint32
}

// MOVELO A UTILS (struct tambien usada por cpu.go)
type InstruccionIO struct {
	PidDesalojado  uint32
	NombreInterfaz string
	Instruccion    string
	UnitWorkTime   int
}

//-------------------------- VARIABLES --------------------------

var newQueue []structs.PCB                           //Debe tener mutex
var readyQueue []structs.PCB                         //Debe tener mutex
var blockedMap = make(map[uint32]structs.PCB)        //Debe tener mutex
var exitQueue []structs.PCB                          //Debe tener mutex
var structsExec structs.PCB                          //Debe tener mutex
var CPUOcupado bool = false                          //Esto se hace con un sem binario
var planificadorActivo bool = true                   //Esto se hace con un sem binario
var interfacesConectadas = make(map[string]Interfaz) //Debe tener mutex
var readyQueueVacia bool = true                      //Esto se hace con un sem binario
var counter int = 0

var hayInterfaz = make(chan int)

func main() {

	var configJson config.Kernel

	// Extrae info de config.json
	config.Iniciar("config.json", &configJson)

	// Testea la conectividad con otros modulos
	//Conectividad(configJson)

	// Configura el logger
	config.Logger("Kernel.log")

	// Establece los endpoints.
	http.HandleFunc("POST /interfaz", handlerIniciarInterfaz)
	http.HandleFunc("POST /instruccion", handlerInstrucciones)

	// Declara su puerto
	port := ":" + strconv.Itoa(configJson.Port)

	// Inicio el servidor en una go-routine para que no bloquee la ejecución del programa
	go func() {
		err := http.ListenAndServe(port, nil)
		if err != nil {
			fmt.Println("Error al escuchar en el puerto " + port)
		}
	}()
	fmt.Printf("Antes del test")

	// Ahora que el servidor está en ejecución, puedo iniciar el ciclo de instrucción
	<-hayInterfaz
	testCicloDeInstruccion(configJson)

	fmt.Printf("Despues del test")

}

//-------------------------- FUNCIONES ---------------------------------------------

// Recibe una interfaz y la agrega al map de interfaces conectadas.
func handlerIniciarInterfaz(w http.ResponseWriter, r *http.Request) {

	//Crea una variable tipo Interfaz (para interpretar lo que se recibe de la requestInterfaz)
	var requestInterfaz RequestInterfaz

	// Decodifica el request (codificado en formato json)
	err := json.NewDecoder(r.Body).Decode(&requestInterfaz)

	// Error Handler de la decodificación
	if err != nil {
		logErrorDecode()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Imprime el request por consola (del lado del server)
	fmt.Println("Request path:", requestInterfaz)

	//Guarda la interfaz en la lista de interfaces conectadas.
	interfacesConectadas[requestInterfaz.NombreInterfaz] = requestInterfaz.Interfaz

	hayInterfaz <- 0
}

// TODO: implementar para los demás tipos de interfaces (cambiar tipos de datos en request y body)
func handlerInstrucciones(w http.ResponseWriter, r *http.Request) {

	//Crea una variable tipo Interfaz (para interpretar lo que se recibe de la request)
	var request InstruccionIO

	// Decodifica el request (codificado en formato json)
	err := json.NewDecoder(r.Body).Decode(&request)

	// Error Handler de la decodificación
	if err != nil {
		logErrorDecode()
		return
	}

	// Imprime el request por consola (del lado del server)
	fmt.Println("Request path:", request)

	// Busca en la lista de interfacesConectadas
	interfaz, encontrado := interfacesConectadas[request.NombreInterfaz]

	// Si no se encontró la interfaz de la request, se desaloja el structs.
	if !encontrado {

		pcbDesalojado := blockedMap[request.PidDesalojado]
		//TODO: Hacer wrapper de delete
		delete(blockedMap, request.PidDesalojado)
		pcbDesalojado.Estado = "EXIT"
		administrarQueues(pcbDesalojado)

		fmt.Println("Interfaz no conectada.")
		return
	}

	//Verificar que la instruccion sea compatible con el tipo de interfaz.
	isValid := validarInstruccion(interfaz.TipoInterfaz, request.Instruccion)

	// IMPLEMENTAR
	if !isValid {

		//!No repetir logica
		pcbDesalojado := blockedMap[request.PidDesalojado]
		//TODO: Hacer wrapper de delete
		delete(blockedMap, request.PidDesalojado)
		pcbDesalojado.Estado = "EXIT"
		administrarQueues(pcbDesalojado)

		//Mandar structs a EXIT
		fmt.Println("Instruccion no compatible.")
		return
	}

	//Agrega el structs a la cola de bloqueados de la interfaz.
	interfaz.QueueBlock = append(interfaz.QueueBlock, request.PidDesalojado)
	interfacesConectadas[request.NombreInterfaz] = interfaz

	//Preparo la interfaz para enviarla en un body.
	body, err := json.Marshal(request.UnitWorkTime)

	//Checkea que no haya errores al crear el body.
	if err != nil {
		fmt.Printf("error codificando body: %s", err.Error())
		return
	}

	// Manda a ejecutar a la interfaz (Puerto)
	respuesta := config.Request(interfaz.PuertoInterfaz, "localhost", "POST", request.Instruccion, body)

	// Verifica que no hubo error en la request
	if respuesta == nil {
		return
	}

	//Si la interfaz pudo ejecutar la instrucción, pasa el structs a READY.
	if respuesta.StatusCode == http.StatusOK {
		//Pasar structs a ready
		//!No repetir logica
		pcbDesalojado := blockedMap[request.PidDesalojado]
		//TODO: Hacer wrapper de delete
		delete(blockedMap, request.PidDesalojado)
		pcbDesalojado.Estado = "READY"
		administrarQueues(pcbDesalojado)
	}
}

func validarInstruccion(tipo string, instruccion string) bool {
	//Verificar que esa interfaz puede ejecutar la instruccion que le pide el CPU
	switch tipo {
	case "GENERICA":
		return instruccion == "IO_GEN_SLEEP"
	}
	return false
}

//-------------------------- ADJACENT FUNCTIONS ------------------------------------

func asignarPCB(nuevoPCB structs.PCB, respuesta structs.Response) {
	// Crea un nuevo PCB

	nuevoPCB.PID = uint32(respuesta.PID)
	pcb_estado_viejo := nuevoPCB.Estado
	nuevoPCB.Estado = "READY"

	//log obligatorio (2/6) (NEW->Ready): Cambio de Estado
	logCambioDeEstado(pcb_estado_viejo, nuevoPCB)

	// Agrega el nuevo PCB a readyQueue
	administrarQueues(nuevoPCB)
}

//-------------------------- API's --------------------------------------------------

func iniciarstructs(configJson config.Kernel, path string) {

	// Se crea un nuevo PCB en estado NEW
	var nuevoPCB structs.PCB
	nuevoPCB.PID = uint32(counter)
	nuevoPCB.Estado = "NEW"

	// Incrementa el contador de structss
	counter++

	// Codificar Body en un array de bytes (formato json)
	body, err := json.Marshal(structs.BodyIniciar{
		PID:  nuevoPCB.PID,
		Path: path,
	})
	// Error Handler de la codificación
	if err != nil {
		fmt.Printf("error codificando body: %s", err.Error())
		return
	}

	//En realidad me parece que se tendría que mandar el path a memoria solamente si hay "espacio" en la readyQueue (depende del grado de multiprogramación)
	// Enviar request al servidor
	respuesta := config.Request(configJson.Port_Memory, configJson.Ip_Memory, "PUT", "process", body)
	// Verificar que no hubo error en la request
	if respuesta == nil {
		return
	}

	// Se declara una nueva variable que contendrá la respuesta del servidor
	var response structs.Response

	// Se decodifica la variable (codificada en formato json) en la estructura correspondiente
	err = json.NewDecoder(respuesta.Body).Decode(&response)

	// Error Handler para al decodificación
	if err != nil {
		fmt.Printf("Error decodificando\n")
		return
	}

	//log obligatorio(1/6): creacion de structs
	logNuevostructs(nuevoPCB)

	asignarPCB(nuevoPCB, response)
}

func finalizarstructs(configJson config.Kernel) {

	// Establecer pid (hardcodeado)
	pid := 0

	// Enviar request al servidor
	respuesta := config.Request(configJson.Port_Memory, configJson.Ip_Memory, "DELETE", fmt.Sprintf("process/%d", pid))
	// verificamos si hubo error en la request
	if respuesta == nil {
		return
	}
}

func estadostructs(configJson config.Kernel) {

	// Establecer pid (hardcodeado)
	pid := 0

	// Enviar request al servidor
	respuesta := config.Request(configJson.Port_Memory, configJson.Ip_Memory, "GET", fmt.Sprintf("process/%d", pid))
	// verificamos si hubo error en la request
	if respuesta == nil {
		return
	}

	// Se declara una nueva variable que contendrá la respuesta del servidor
	var response structs.Response

	// Se decodifica la variable (codificada en formato json) en la estructura correspondiente
	err := json.NewDecoder(respuesta.Body).Decode(&response)

	// Error Handler para al decodificación
	if err != nil {
		fmt.Printf("Error decodificando\n")
		fmt.Println(err)
		return
	}

	// Imprime pid (parámetro de la estructura)
	fmt.Println(response)
}

func listarstructs(configJson config.Kernel) {

	// Enviar request al servidor
	respuesta := config.Request(configJson.Port_Memory, configJson.Ip_Memory, "GET", "process")
	// verificamos si hubo error en la request
	if respuesta == nil {
		return
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return
	}

	fmt.Println(string(bodyBytes))
}

func iniciarPlanificacion(configJson config.Kernel) {
	// Enviar request al servidor
	respuesta := config.Request(configJson.Port_CPU, configJson.Ip_CPU, "PUT", "plani")
	// Verificar que no hubo error en la request
	if respuesta == nil {
		return
	}
}

func detenerPlanificacion(configJson config.Kernel) {
	// Enviar request al servidor
	respuesta := config.Request(configJson.Port_CPU, configJson.Ip_CPU, "DELETE", "plani")
	// Verificar que no hubo error en la request
	if respuesta == nil {
		return
	}
}

type RespuestaDispatch struct {
	MotivoDeDesalojo string
	PCB              structs.PCB
}

// Le envia un PCB al CPU
func dispatch(pcb structs.PCB, configJson config.Kernel) (structs.PCB, string) {

	//Envia PCB al CPU
	fmt.Println("Se envió el structs", pcb.PID, "al CPU")

	//Pasan cosas de HTTP y se ejecuta el structs
	CPUOcupado = true

	//-------------------Request al CPU------------------------

	// Codificar Body en un array de bytes (formato json)
	body, err := json.Marshal(pcb)
	// Error Handler de la codificación
	if err != nil {
		fmt.Printf("error codificando body: %s", err.Error())
		return structs.PCB{}, "ERROR"
	}

	// Enviar request al servidor
	respuesta := config.Request(configJson.Port_CPU, configJson.Ip_CPU, "POST", "exec", body)
	// Verificar que no hubo error en la request
	if respuesta == nil {
		return structs.PCB{}, "ERROR"
	}

	// Se declara una nueva variable que contendrá la respuesta del servidor
	var respuestaDispatch RespuestaDispatch

	// Se decodifica la variable (codificada en formato json) en la estructura correspondiente
	err = json.NewDecoder(respuesta.Body).Decode(&respuestaDispatch)

	// Error Handler para al decodificación
	if err != nil {
		fmt.Printf("Error decodificando\n")
		return structs.PCB{}, "ERROR"
	}
	//-------------------Fin Request al CPU------------------------
	fmt.Println("Motivo de desalojo:", respuestaDispatch.MotivoDeDesalojo)

	CPUOcupado = false

	fmt.Println("Exit queue:", exitQueue)

	return respuestaDispatch.PCB, respuestaDispatch.MotivoDeDesalojo
}

// Envía una interrupción al ciclo de instrucción.
func interrupt(pid int, tipoDeInterrupcion string, configJson config.Kernel) {

	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/interrupciones", configJson.Ip_CPU, configJson.Port_CPU)
	req, err := http.NewRequest("POST", url, nil)

	if err != nil {
		return
	}

	//string(pid)
	pidString := strconv.Itoa(pid)

	//Pasa como param el tipoDeInterrupcion
	q := req.URL.Query()
	q.Add("pid", string(pidString))
	q.Add("interrupt_type", tipoDeInterrupcion)

	req.URL.RawQuery = q.Encode()

	//Envía la request con el pid y el tipo de interrupcion.
	req.Header.Set("Content-Type", "text/plain")
	respuesta, err := cliente.Do(req)

	// Verifica que no hubo error al enviar la request
	if err != nil {
		fmt.Println("Error al enviar la interrupción a CPU.")
		return
	}

	// Verifica que no hubo error en la respuesta.
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
		newQueue = append(newQueue, pcb)

	case "READY":
		readyQueue = append(readyQueue, pcb)
		readyQueueVacia = false
		logPidsReady(readyQueue)

	//Deberia ser una por cada IO. Revisar
	case "BLOCK":
		//Agrega el pcb en el map de pcb's bloqueados.
		blockedMap[pcb.PID] = pcb

		//TODO: Implementar log para el manejo de listas BLOCK con map
		//logPidsBlock(blockedMap)

	case "EXIT":
		exitQueue = append(exitQueue, pcb)
		//TODO: momentaneamente sera un string constante, pero el motivo de Finalizacion deberá venir con el pcb (o alguna estructura que la contenga)
		//motivoDeFinalizacion := "SUCCESS"
		//logFinDestructs(pcb, motivoDeFinalizacion)
	}
}

// Envía continuamente structss al CPU mientras que el bool planificadorActivo sea TRUE y el CPU esté esperando un structs.
func planificador(configJson config.Kernel) {
	if !CPUOcupado && !readyQueueVacia {
		planificadorActivo = true
	}
	for planificadorActivo {
		//Si el CPU está ocupado, se detiene el planificador
		if CPUOcupado {
			planificadorActivo = false
			break
		}
		//Si no...

		//Si la lista de READY está vacía, se detiene el planificador
		if len(readyQueue) == 0 {
			//Si la lista está vacía, se detiene el planificador
			logEsperaNuevosstructss()
			readyQueueVacia = true
			planificadorActivo = false
			break
		}

		//Si la lista no está vacía, se envía el structs al CPU
		//Se envía el primer structs y se hace un dequeue del mismo de la lista READY
		var poppedPCB structs.PCB
		readyQueue, poppedPCB = dequeuePCB(readyQueue)
		//Debería estar en dispatch?
		estadoAExec(&poppedPCB)
		//Será siempre READY cuando pasa a EXEC?
		logCambioDeEstado("READY", poppedPCB)
		pcbActualizado, motivoDesalojo := dispatch(poppedPCB, configJson)

		//TODO: Usar motivo de desalojo para algo.
		fmt.Println(motivoDesalojo)

		administrarQueues(pcbActualizado)
	}
}

// Desencola el PCB de la lista, si esta está vacía, simplemente espera nuevos structss, y avisa que la lista está vacía
func dequeuePCB(listaPCB []structs.PCB) ([]structs.PCB, structs.PCB) {

	return listaPCB[1:], listaPCB[0]
}

func estadoAExec(pcb *structs.PCB) {
	(*pcb).Estado = "EXEC"
	structsExec = *pcb
}

//-------------------------- TEST --------------------------------------------------

// Testea la conectividad con otros módulos

func testConectividad(configJson config.Kernel) {
	fmt.Println("\nIniciar structs:")
	iniciarstructs(configJson, "path")
	iniciarstructs(configJson, "path")
	iniciarstructs(configJson, "path")
	iniciarstructs(configJson, "path")
	fmt.Println("\nFinalizar structs:")
	finalizarstructs(configJson)
	fmt.Println("\nEstado structs:")
	estadostructs(configJson)
	fmt.Println("\nListar structss:")
	listarstructs(configJson)
	fmt.Println("\nDetener Planificación:")
	detenerPlanificacion(configJson)
	fmt.Println("\nIniciar Planificación:")
	iniciarPlanificacion(configJson)
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
	fmt.Printf("\nSe crean 2 structss-------------\n\n")
	for i := 0; i < 2; i++ {
		path := "structs" + strconv.Itoa(counter) + ".txt"
		iniciarstructs(configJson, path)
	}

	fmt.Printf("\nSe testea el planificador-------------\n\n")
	planificador(configJson)
	printList()

	fmt.Printf("\nSe crean 2 structss-------------\n\n")
	for i := 0; i < 2; i++ {
		path := "structs" + strconv.Itoa(counter) + ".txt"
		iniciarstructs(configJson, path)
	}
}

func testCicloDeInstruccion(configJson config.Kernel) {

	fmt.Printf("\nSe crean 1 structs-------------\n\n")
	iniciarstructs(configJson, "structs_test")

	fmt.Printf("\nSe testea el planificador-------------\n\n")
	planificador(configJson)
}

// -------------------------- LOG's --------------------------------------------------
// log obligatorio (1/6)
func logNuevostructs(nuevoPCB structs.PCB) {

	log.Printf("Se crea el structs %d en estado %s", nuevoPCB.PID, nuevoPCB.Estado)
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
func logFinDestructs(pcb structs.PCB, motivoDeFinalizacion string) {

	log.Printf("Finaliza el structs: %d - Motivo: %s", pcb.PID, motivoDeFinalizacion)

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

func logEsperaNuevosstructss() {

	fmt.Println("Esperando nuevos structss...")

}

func logErrorDecode() {

	fmt.Printf("Error al decodificar request body: ")
}
