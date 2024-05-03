package proceso

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/sisoputnfrba/tp-golang/utils/config"
)

//CLIENT SIDE/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//---VARIABLES && STRUCTS--

type BodyIniciar struct {
	// Path del archivo que se utilizará como base para ejecutar un nuevo proceso
	Path string `json:"path"`
}

//--FUNCIONES AUX--

//--CALLS--

// Solamente esqueleto
func Iniciar(configJson config.Kernel) {

	// Codificar Body en un array de bytes (formato json)
	body, err := json.Marshal(BodyIniciar{
		Path: "string",
	})
	// Error Handler de la codificación
	if err != nil {
		fmt.Printf("error codificando body: %s", err.Error())
		return
	}

	// Enviar request al servidor
	respuesta := config.EnviarBodyRequest("PUT", "process", body, configJson.Port_Memory, configJson.Ip_Memory)
	// Verificar que no hubo error en la request
	if respuesta == nil {
		return
	}

	// Se declara una nueva variable que contendrá la respuesta del servidor
	var response Response

	// Se decodifica la variable (codificada en formato json) en la estructura correspondiente
	err = json.NewDecoder(respuesta.Body).Decode(&response)

	// Error Handler para al decodificación
	if err != nil {
		fmt.Printf("Error decodificando\n")
		return
	}

	// Imprime pid (parámetro de la estructura)
	fmt.Printf("pid: %d\n", response.Pid)
}

// Solamente esqueleto
func Finalizar(configJson config.Kernel) {

	// Establecer pid (hardcodeado)
	pid := 0

	// Enviar request al servidor
	respuesta := config.EnviarRequest("DELETE", fmt.Sprintf("process/%d", pid), configJson.Port_Memory, configJson.Ip_Memory)
	// verificamos si hubo error en la request
	if respuesta == nil {
		return
	}

}

func Estado(configJson config.Kernel) {

	// Establecer pid (hardcodeado)
	pid := 0

	// Enviar request al servidor
	respuesta := config.EnviarRequest("GET", fmt.Sprintf("process/%d", pid), configJson.Port_Memory, configJson.Ip_Memory)
	// verificamos si hubo error en la request
	if respuesta == nil {
		return
	}

	// Se declara una nueva variable que contendrá la respuesta del servidor
	var response Response

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

func Listar(configJson config.Kernel) {

	// Enviar request al servidor
	respuesta := config.EnviarRequest("GET", "process", configJson.Port_Memory, configJson.Ip_Memory)
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

//SERVER SIDE/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//--VARIABLES && STRUCTS--

type Response struct {
	Pid    int    `json:"pid"`
	Estado string `json:"estado"`
}

// Estructura de los PCB
type PCB struct {
	PID     uint32
	PC      uint32
	Quantum uint16
	RegistrosUsoGeneral
}

// Estructura de los registros de uso general (para tener info del contexto de ejecución de cada PCB)
type RegistrosUsoGeneral struct {
	AX  uint8
	BX  uint8
	CX  uint8
	DX  uint8
	EAX uint16
	EBX uint16
	ECX uint16
	EDX uint16
	SI  uint32
	DI  uint32
}

// Variable global para llevar la cuenta de los procesos (y así poder nombrarlos de manera correcta)
var Counter int = 0

// Lista que contiene los PCBs (procesos)
var listaPCB []PCB

//--HANDLERS--

func HandlerIniciar(w http.ResponseWriter, r *http.Request) {

	//Crea uan variable tipo BodyIniciar (para interpretar lo que se recibe de la request)
	var request BodyIniciar

	// Decodifica el request (codificado en formato json)
	err := json.NewDecoder(r.Body).Decode(&request)

	// Error Handler de la decodificación
	if err != nil {
		fmt.Printf("Error al decodificar request body: ")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Imprime el request por consola (del lado del server)
	fmt.Printf("Request path: %s\n", request)

	//Crea una variable tipo Response (para confeccionar una respuesta)
	var respBody Response = Response{Pid: Counter}

	// Codificar Response en un array de bytes (formato json)
	respuesta, err := json.Marshal(respBody)

	// Error Handler de la codificación
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)

	// Crea un nuevo PCB
	var nuevoPCB PCB
	nuevoPCB.PID = uint32(Counter)

	// Agrega el nuevo PCB a la lista de PCBs
	listaPCB = append(listaPCB, nuevoPCB)
	for _, pcb := range listaPCB {
		fmt.Print(pcb.PID, "\n")
	}

	// Incrementa el contador de procesos
	Counter++
}

// primera versión de finalizar proceso, no recibe body (solo un path por medio de la url) y envía una respuesta vacía (mandamos status ok y hacemos que printee el valor del pid recibido para ver que ha sido llamada).
// Cuando haya  procesos se busca por el path {pid}
func HandlerFinalizar(w http.ResponseWriter, r *http.Request) {

	//es posible que en un futuro sea necesario convertir esta string a un int
	pid := r.PathValue("pid")

	// Imprime el pid (solo para pruebas)
	fmt.Printf("pid: %s", pid)

	// Respuesta vacía significa que manda una respuesta vacía, o que no hay respuesta?
	respuesta, err := json.Marshal("")

	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

// primera versión de estado proceso, como es GET no necesita recibir nada
// Cuando haya  procesos se busca por el %s del path {pid}
func HandlerEstado(w http.ResponseWriter, r *http.Request) {
	//usando el struct de Response envío el estado del proceso

	pid, error := strconv.Atoi(r.PathValue("pid"))

	if error != nil {
		http.Error(w, "Error al obtener el ID del proceso", http.StatusInternalServerError)
		return
	}
	//Crea una variable tipo Response (para confeccionar una respuesta)
	var respBody Response = Response{Pid: pid, Estado: "READY"}

	// Codificar Response en un array de bytes (formato json)
	respuesta, err := json.Marshal(respBody)

	// Error Handler de la codificación
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

/*
Se encargará de mostrar por consola y retornar por la api el listado de procesos
que se encuentran en el sistema con su respectivo estado dentro de cada uno de ellos.
*/
func HandlerListar(w http.ResponseWriter, r *http.Request) {

	//Harcodea una lista de procesos, más adelante deberá ser dinámico.
	var listaDeProcesos []Response = []Response{
		{Pid: 0, Estado: "READY"},
		{Pid: 1, Estado: "BLOCK"},
	}

	//Paso a formato JSON la lista de procesos.
	respuesta, err := json.Marshal(listaDeProcesos)

	//Check si hubo algún error al parsear el JSON.
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Envía respuesta (con estatus como header) al cliente.
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}
