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

func main() {

	// Configura el logger
	config.Logger("Kernel.log")

	log.Printf("Soy un logeano")

	// Extrae info de config.json
	var configJson config.Kernel

	config.Iniciar("config.json", &configJson)

	// teste la conectividad con otros modulos
	Conectividad(configJson)

	//Establezco petición
	http.HandleFunc("GET /holamundo", kernel)

	// declaro puerto
	port := ":" + strconv.Itoa(configJson.Port)

	// Listen and serve con info del config.json
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al esuchar en el puerto " + port)
	}

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
	nuevoPCB.Estado = "READY"

	// Agrega el nuevo PCB a la lista de PCBs
	proceso.ReadyQueue = append(proceso.ReadyQueue, nuevoPCB) //AAAAAAAAAAAAAAAAAAAAAAAAA
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

	asignarPCB(nuevoPCB, response)

	//Funcionalidades temporales para testing
	testing := func() {
		// Imprime pid (parámetro de la estructura)
		fmt.Printf("pid: %d\n", response.Pid)

		for _, pcb := range proceso.ReadyQueue { //AAAAAAAAAAAAAAAAAAAAAAAAA
			fmt.Print(pcb.PID, "\n")
		}

		fmt.Println("Counter:", proceso.Counter) //AAAAAAAAAAAAAAAAAAAAAAAAA
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

func dispatch() {

}

func interrupt() {

}
