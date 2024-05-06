package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/utils/APIs/kernel-memoria/proceso"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

//================================| MAIN |===================================================\\

//-------------------------- STRUCTS -----------------------------------------

// MOVELO A UTILS (struct tambien usada por entrasalida.go)
type Interfaz struct {
	Nombre string
	Tipo   string
	Puerto int
}

//-------------------------- VARIABLES --------------------------

var ReadyQueue []proceso.PCB
var BlockQueue []proceso.PCB
var CPUActivo bool = false
var planificadorActivo bool = true
var ReadyQueueVacia bool = true
var interfacesConectadas []Interfaz

func main() {

	// Configura el logger
	config.Logger("Kernel.log")

	// Extrae info de config.json
	var configJson config.Kernel

	config.Iniciar("config.json", &configJson)

	// testea la conectividad con otros modulos
	//Conectividad(configJson)

	//Establezco petición
	http.HandleFunc("POST /interfaz", handlerIniciarInterfaz)

	// declaro puerto
	port := ":" + strconv.Itoa(configJson.Port)

	// Listen and serve con info del config.json
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al esuchar en el puerto " + port)
	}

}

//-------------------------- FUNCIONES ---------------------------------------------

func handlerIniciarInterfaz(w http.ResponseWriter, r *http.Request) {

	//Crea una variable tipo Interfaz (para interpretar lo que se recibe de la request)
	var request Interfaz

	// Decodifica el request (codificado en formato json)
	err := json.NewDecoder(r.Body).Decode(&request)

	// Error Handler de la decodificación
	if err != nil {
		fmt.Printf("Error al decodificar request body: ")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Imprime el request por consola (del lado del server)
	fmt.Println("Request path:", request)

	//Guarda la interfaz en la lista de interfaces conectadas.
	interfacesConectadas = append(interfacesConectadas, request)
}

//-------------------------- ADJACENT FUNCTIONS ------------------------------------

func asignarPCB(nuevoPCB proceso.PCB, respuesta proceso.Response) {
	// Crea un nuevo PCB

	nuevoPCB.PID = uint32(respuesta.Pid)
	nuevoPCB.Estado = "READY"
	pcb_estado_viejo := "NEW"

	//log obligatorio (2/6) (NEW->Ready): Cambio de Estado
	logCambioDeEstado(pcb_estado_viejo, nuevoPCB)

	// Agrega el nuevo PCB a la lista de PCBs
	ReadyQueue = append(ReadyQueue, nuevoPCB)

	// for _, pcb := range queuePCB {
	// 	fmt.Print(pcb.PID, "\n")
	// }
}

//-------------------------- TEST --------------------------------------------------

// Testea la conectividad con otros módulos

func Conectividad(configJson config.Kernel) {
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

	//log obligatorio (3/6): Ingreso a READY
	logPidsReady(ReadyQueue)

	//Funcionalidades temporales para testing
	testing := func() {
		// Imprime pid (parámetro de la estructura)
		fmt.Printf("pid: %d\n", response.Pid)

		for _, pcb := range ReadyQueue {
			fmt.Print(pcb.PID, "\n")
		}

		fmt.Println("Counter:", proceso.Counter)
	}

	testing()
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

// Función que según que se haga con un PCB se lo puede enviar a la lista de planificación o a la de bloqueo
func EnviarAPlanificador(pcb proceso.PCB) {
	if pcb.Estado == "READY" {
		ReadyQueue = append(ReadyQueue, pcb)
		ReadyQueueVacia = false
		logPidsReady(ReadyQueue)

	} else if pcb.Estado == "BLOCK" {
		BlockQueue = append(BlockQueue, pcb)
		logPidsBlock(BlockQueue)
	}
}

// Envía continuamente procesos al CPU mientras que el bool planificadorActivo sea TRUE y el CPU esté esperando un proceso.
func planificador() {
	for planificadorActivo {
		if !CPUActivo {
			//Se envía el primer proceso y se hace un pop del mismo de la lista
			if !ReadyQueueVacia {
				dispatch(ReadyQueue[0])
			}
			CPUActivo = true
			ReadyQueue = pop(ReadyQueue)
		}
	}
}

// Elimina el primer PCB de la lista, si esta está vacía, simplemente espera a que vuelva a comenzar
func pop(listaPCB []proceso.PCB) []proceso.PCB {
	l := len(listaPCB)

	if l == 0 {
		log.Print("Esperando nuevos procesos...")
		ReadyQueueVacia = true
		return nil
	}

	return listaPCB[:l-1]
}

func dispatch(pcb proceso.PCB) {
	//envía a CPU el PCB
}

func interrupt() {

}

// -------------------------- LOG´s --------------------------------------------------
// log obligatorio (1/6)
func logNuevoProceso(response proceso.Response, nuevoPCB proceso.PCB) {

	log.Printf("Se crea el proceso %d en estado %s", response.Pid, nuevoPCB.Estado)
}

// log obligatorio (2/6)
func logCambioDeEstado(pcb_estado_viejo string, pcb proceso.PCB) {

	log.Printf("PID: %d - Estado anterior: %s - Estado actual: %s", pcb.PID, pcb_estado_viejo, pcb.Estado)

}

// log obligatorio (3/6)
func logPidsReady(ReadyQueue []proceso.PCB) {
	var pids []uint32
	//Recorre la lista READY y guarda sus PIDs
	for _, pcb := range ReadyQueue {
		pids = append(pids, pcb.PID)
	}

	log.Printf("Cola Ready 'ReadyQueue' : %v", pids)
}

//LUEGO IMPLEMENTAR EN NUESTRO ARCHIVO NO OFICIAL DE LOGS

// log para el manejo de listas BLOCK
func logPidsBlock(BlockQueue []proceso.PCB) {
	var pids []uint32
	//Recorre la lista BLOCK y guarda sus PIDs
	for _, pcb := range BlockQueue {
		pids = append(pids, pcb.PID)
	}

	fmt.Printf("Cola Block 'BlockQueue' : %v", pids)
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
