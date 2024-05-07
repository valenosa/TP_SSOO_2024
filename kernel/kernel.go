package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/sisoputnfrba/tp-golang/utils/APIs/kernel-memoria/proceso"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

//================================| MAIN |===================================================\\

//-------------------------- STRUCTS -----------------------------------------

//-------------------------- VARIABLES ---------------------------------------

var readyQueue []proceso.PCB
var blockQueue []proceso.PCB
var CPUOcupado bool = false
var planificadorActivo bool = true
var readyQueueVacia bool = true

func main() {
	var configJson config.Kernel

	// Configura el logger
	config.Logger("Kernel.log")

	// Extrae info de config.json

	config.Iniciar("config.json", &configJson)

	// teste la conectividad con otros modulos
	testPlanificacion(configJson)

	// //Establezco petición
	// http.HandleFunc("GET /holamundo", kernel)

	// // declaro puerto
	// port := ":" + strconv.Itoa(configJson.Port)

	// // Listen and serve con info del config.json
	// err := http.ListenAndServe(port, nil)
	// if err != nil {
	// 	fmt.Println("Error al esuchar en el puerto " + port)
	// }

}

//-------------------------- FUNCIONES ---------------------------------------------

func kernel(w http.ResponseWriter, r *http.Request) {

	respuesta, err := json.Marshal("Hello world! Soy una consola del kernel")

	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)

	fmt.Println("Hello world! Soy una consola del kernel")
}

//-------------------------- ADJACENT FUNCTIONS ------------------------------------

func asignarPCB(nuevoPCB proceso.PCB, respuesta proceso.Response) {
	// Crea un nuevo PCB

	nuevoPCB.PID = uint32(respuesta.Pid)
	pcb_estado_viejo := nuevoPCB.Estado
	nuevoPCB.Estado = "READY"

	//log obligatorio (2/6) (NEW->Ready): Cambio de Estado
	logCambioDeEstado(pcb_estado_viejo, nuevoPCB)

	// Agrega el nuevo PCB a readyQueue
	enviarAPlanificador(nuevoPCB)

}

//-------------------------- API's --------------------------------------------------

func iniciarProceso(configJson config.Kernel) {

	// Codificar Body en un array de bytes (formato json)
	body, err := json.Marshal(proceso.BodyIniciar{
		Path: "string",
	})
	// Error Handler de la codificación
	if err != nil {
		fmt.Printf("error codificando body: %s", err.Error())
		return
	}

	// Enviar request al servidor
	respuesta := config.Request(configJson.Port_Memory, configJson.Ip_Memory, "PUT", "process", body)
	// Verificar que no hubo error en la request
	if respuesta == nil {
		return
	}
	// Se crea un nuevo PCB en estado NEW
	var nuevoPCB proceso.PCB
	nuevoPCB.Estado = "NEW"

	// Se declara una nueva variable que contendrá la respuesta del servidor
	var response proceso.Response

	// Se decodifica la variable (codificada en formato json) en la estructura correspondiente
	err = json.NewDecoder(respuesta.Body).Decode(&response)

	// Error Handler para al decodificación
	if err != nil {
		fmt.Printf("Error decodificando\n")
		return
	}

	//log obligatorio(1/6): creacion de proceso
	logNuevoProceso(response, nuevoPCB)

	asignarPCB(nuevoPCB, response)
}

func finalizarProceso(configJson config.Kernel) {

	// Establecer pid (hardcodeado)
	pid := 0

	// Enviar request al servidor
	respuesta := config.Request(configJson.Port_Memory, configJson.Ip_Memory, "DELETE", fmt.Sprintf("process/%d", pid))
	// verificamos si hubo error en la request
	if respuesta == nil {
		return
	}
}

func estadoProceso(configJson config.Kernel) {

	// Establecer pid (hardcodeado)
	pid := 0

	// Enviar request al servidor
	respuesta := config.Request(configJson.Port_Memory, configJson.Ip_Memory, "GET", fmt.Sprintf("process/%d", pid))
	// verificamos si hubo error en la request
	if respuesta == nil {
		return
	}

	// Se declara una nueva variable que contendrá la respuesta del servidor
	var response proceso.Response

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

func listarProceso(configJson config.Kernel) {

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

func dispatch(pcb proceso.PCB, configJson config.Kernel) {
	//Envia PCB al CPU
	fmt.Println("Se envió el proceso", pcb.PID, "al CPU")

	//Pasan cosas de HTTP y se ejecuta el proceso
	CPUOcupado = true

	//-------------------Request al CPU------------------------
	// Codificar Body en un array de bytes (formato json)
	body, err := json.Marshal(pcb)
	// Error Handler de la codificación
	if err != nil {
		fmt.Printf("error codificando body: %s", err.Error())
		return
	}

	// Enviar request al servidor
	respuesta := config.Request(configJson.Port_CPU, configJson.Ip_CPU, "POST", "exec", body)
	// Verificar que no hubo error en la request
	if respuesta == nil {
		return
	}

	// Se declara una nueva variable que contendrá la respuesta del servidor
	var response string

	// Se decodifica la variable (codificada en formato json) en la estructura correspondiente
	err = json.NewDecoder(respuesta.Body).Decode(&response)

	// Error Handler para al decodificación
	if err != nil {
		fmt.Printf("Error decodificando\n")
		return
	}
	//-------------------Fin Request al CPU------------------------

	//Se muestra la respuesta del CPU
	fmt.Println(response)
	CPUOcupado = false
}

func interrupt() {
}

//-------------------------- PLANIFICACIÓN --------------------------------------------------

// Función que según que se haga con un PCB se lo puede enviar a la lista de planificación o a la de bloqueo
func enviarAPlanificador(pcb proceso.PCB) {
	if pcb.Estado == "READY" {
		readyQueue = append(readyQueue, pcb)
		readyQueueVacia = false
		logPidsReady(readyQueue)

	} else if pcb.Estado == "BLOCK" {
		blockQueue = append(blockQueue, pcb)
		logPidsBlock(blockQueue)
	}
}

// Envía continuamente procesos al CPU mientras que el bool planificadorActivo sea TRUE y el CPU esté esperando un proceso.
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
			fmt.Println("Esperando nuevos procesos...") //Se puede implementar un log en nuestro archivo no-oficial de logs
			readyQueueVacia = true
			planificadorActivo = false
			break
		}

		//Si la lista no está vacía, se envía el proceso al CPU
		//Se envía el primer proceso y se hace un dequeue del mismo de la lista READY
		var poppedPCB proceso.PCB
		readyQueue, poppedPCB = dequeuePCB(readyQueue)
		estadoAExec(&poppedPCB)
		dispatch(poppedPCB, configJson)
	}
}

// Desencola el PCB de la lista, si esta está vacía, simplemente espera nuevos procesos, y avisa que la lista está vacía
func dequeuePCB(listaPCB []proceso.PCB) ([]proceso.PCB, proceso.PCB) {

	return listaPCB[1:], listaPCB[0]
}

func estadoAExec(pcb *proceso.PCB) {
	(*pcb).Estado = "EXEC"
}

//-------------------------- TEST --------------------------------------------------

// Testea la conectividad con otros módulos

func testConectividad(configJson config.Kernel) {
	fmt.Println("\nIniciar Proceso:")
	iniciarProceso(configJson)
	iniciarProceso(configJson)
	iniciarProceso(configJson)
	iniciarProceso(configJson)
	fmt.Println("\nFinalizar Proceso:")
	finalizarProceso(configJson)
	fmt.Println("\nEstado Proceso:")
	estadoProceso(configJson)
	fmt.Println("\nListar Procesos:")
	listarProceso(configJson)
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
	fmt.Println("\nSe crean 2 procesos-------------\n")
	for i := 0; i < 2; i++ {
		iniciarProceso(configJson)
	}

	fmt.Println("\nSe testea el planificador-------------\n")
	planificador(configJson)
	printList()

	fmt.Println("\nSe crean 2 procesos-------------\n")
	for i := 0; i < 2; i++ {
		iniciarProceso(configJson)
	}
}

// -------------------------- LOG's --------------------------------------------------
// log obligatorio (1/6)
func logNuevoProceso(response proceso.Response, nuevoPCB proceso.PCB) {

	log.Printf("Se crea el proceso %d en estado %s", response.Pid, nuevoPCB.Estado)
}

// log obligatorio (2/6)
func logCambioDeEstado(pcb_estado_viejo string, pcb proceso.PCB) {

	log.Printf("PID: %d - Estado anterior: %s - Estado actual: %s", pcb.PID, pcb_estado_viejo, pcb.Estado)

}

// log obligatorio (3/6)
func logPidsReady(readyQueue []proceso.PCB) {
	var pids []uint32
	//Recorre la lista READY y guarda sus PIDs
	for _, pcb := range readyQueue {
		pids = append(pids, pcb.PID)
	}

	log.Printf("Cola Ready 'readyQueue' : %v", pids)
}

//LUEGO IMPLEMENTAR EN NUESTRO ARCHIVO NO OFICIAL DE LOGS

// log para el manejo de listas BLOCK
func logPidsBlock(blockQueue []proceso.PCB) {
	var pids []uint32
	//Recorre la lista BLOCK y guarda sus PIDs
	for _, pcb := range blockQueue {
		pids = append(pids, pcb.PID)
	}

	fmt.Printf("Cola Block 'blockQueue' : %v", pids)
}

// log para el manejo de listas EXEC
func logPidsExec(ExecQueue []proceso.PCB) {
	var pids []uint32
	//Recorre la lista EXEC y guarda sus PIDs
	for _, pcb := range ExecQueue {
		pids = append(pids, pcb.PID)
	}

	fmt.Printf("Cola Executing 'ExecQueue' : %v", pids)
}
