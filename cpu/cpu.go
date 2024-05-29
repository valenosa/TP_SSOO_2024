package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

//*======================================| MAIN |=======================================\\

var registrosCPU structs.RegistrosUsoGeneral

var configJson config.Cpu

func main() {

	// Configura el logger
	config.Logger("CPU.log")

	// ======== HandleFunctions ========
	// Se establece el handler que se utilizará para las diversas situaciones recibidas por el server
	http.HandleFunc("PUT /plani", handlerIniciarPlanificacion)
	http.HandleFunc("DELETE /plani", handlerDetenerPlanificacion)

	http.HandleFunc("POST /exec", handlerEjecutarProceso)
	http.HandleFunc("POST /interrupciones", handlerInterrupcion)

	// Extrae info de config.json
	config.Iniciar("config.json", &configJson)

	//inicio el servidor de CPU
	config.IniciarServidor(configJson.Port)

}

// *======================================| HANDLERS |=======================================\\
func handlerIniciarPlanificacion(w http.ResponseWriter, r *http.Request) {

	// Convierte una cadena vacía a JSON
	respuesta, err := json.Marshal("")

	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Envía una respuesta vacía con el estado 200 OK al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

func handlerDetenerPlanificacion(w http.ResponseWriter, r *http.Request) {

	// Convierte una cadena vacía a JSON
	respuesta, err := json.Marshal("")

	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Envía una respuesta vacía con el estado 200 OK al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

// Contiene el pid del proceso que dispatch mandó a ejecutar
var pidEnEjecucion uint32

// TODO: Hay que pasarla a local
var motivoDeDesalojo string

// Maneja la ejecución de un proceso a través de un PCB
// Devuelve al despachador el contexto de ejecución y el motivo del desalojo.
func handlerEjecutarProceso(w http.ResponseWriter, r *http.Request) {
	// Crea una variable tipo BodyIniciar (para interpretar lo que se recibe de la pcbRecibido)
	var pcbRecibido structs.PCB

	// Decodifica el request (codificado en formato JSON)
	err := json.NewDecoder(r.Body).Decode(&pcbRecibido)

	// Error Handler de la decodificación
	if err != nil {
		fmt.Printf("Error al decodificar request body: ")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Ejecuta el ciclo de instrucción.
	pidEnEjecucion = pcbRecibido.PID
	ejecutarCiclosDeInstruccion(&pcbRecibido)

	fmt.Println("Se está ejecutando el proceso: ", pcbRecibido.PID)

	// Devuelve a dispatch el contexto de ejecucion y el motivo del desalojo
	respuesta, err := json.Marshal(structs.RespuestaDispatch{
		MotivoDeDesalojo: motivoDeDesalojo,
		PCB:              pcbRecibido,
	})

	// Error Handler de la codificación
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

var hayInterrupcion bool = false

// Checkea que Kernel no haya enviado interrupciones
func handlerInterrupcion(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	// Está en una global; despues cambiar.
	motivoDeDesalojo = queryParams.Get("interrupt_type")

	PID, errPid := strconv.ParseUint(queryParams.Get("PID"), 10, 32)

	if errPid != nil {
		return
	}

	if uint32(PID) != pidEnEjecucion {
		return
	}

	hayInterrupcion = true

	//TODO: Checkear si es necesario lo de abajo (27/05/24).
	/*en caso de que haya interrupcion,
	se devuelve el Contexto de Ejecución actualizado al Kernel con motivo de la interrupción.*/

	// respuesta, err := json.Marshal(instruccion)
	// fmt.Println(respuesta)

	// if err != nil {
	// 	http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
	// 	return
	// }

	// w.WriteHeader(http.StatusOK)
	// w.Write([]byte(instruccion))
}

//---------------------FUNCIONES CICLO DE INSTRUCCION---------------------

// Ejecuta un ciclo de instruccion.
func ejecutarCiclosDeInstruccion(PCB *structs.PCB) {
	var cicloFinalizado bool = false

	// Itera el ciclo de instrucción si hay instrucciones a ejecutar y no hay interrupciones.
	for !hayInterrupcion && !cicloFinalizado {
		// Obtiene la próxima instrucción a ejecutar.
		instruccion := fetch(PCB.PID, registrosCPU.PC)

		// Decodifica y ejecuta la instrucción.
		decodeAndExecute(PCB, instruccion, &registrosCPU.PC, &cicloFinalizado)
	}

	// Actualiza los registros de uso general del PCB con los registros de la CPU.
	PCB.RegistrosUsoGeneral = registrosCPU

}

// Trae de memoria las instrucciones indicadas por el PC y el PID.
func fetch(PID uint32, PC uint32) string {

	// Convierte el PID y el PC a string
	pid := strconv.FormatUint(uint64(PID), 10)
	pc := strconv.FormatUint(uint64(PC), 10)

	// Crea un cliente HTTP
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/instrucciones", configJson.Ip_Memory, configJson.Port_Memory)

	// Crea una nueva solicitud GET
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}

	// Agrega el PID y el PC como params
	q := req.URL.Query()
	q.Add("PID", pid)
	q.Add("PC", pc)
	req.URL.RawQuery = q.Encode()

	// Establece el tipo de contenido de la solicitud
	req.Header.Set("Content-Type", "text/plain")

	// Realiza la solicitud al servidor de memoria
	respuesta, err := cliente.Do(req)
	if err != nil {
		return ""
	}

	// Verifica el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		return ""
	}

	// Lee el cuerpo de la respuesta
	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return ""
	}

	// Retorna las instrucciones obtenidas como una cadena de texto
	return string(bodyBytes)
}

// Ejecuta las instrucciones traidas de memoria.
func decodeAndExecute(PCB *structs.PCB, instruccion string, PC *uint32, cicloFinalizado *bool) {

	// Mapa de registros para acceder a los registros de la CPU por nombre
	var registrosMap = map[string]*uint8{
		"AX": &registrosCPU.AX,
		"BX": &registrosCPU.BX,
		"CX": &registrosCPU.CX,
		"DX": &registrosCPU.DX,
	}

	// Parsea las instrucciones de la cadena de instrucción
	variable := strings.Split(instruccion, " ")

	// Imprime la instrucción y sus parámetros
	fmt.Println("Instruccion: ", variable[0], " Parametros: ", variable[1:])

	// Switch para determinar la operación a realizar según la instrucción
	switch variable[0] {
	case "SET":
		set(variable[1], variable[2], registrosMap, PC)

	case "SUM":
		sum(variable[1], variable[2], registrosMap)

	case "SUB":
		sub(variable[1], variable[2], registrosMap)

	case "JNZ":
		jnz(variable[1], variable[2], PC, registrosMap)

	case "IO_GEN_SLEEP":
		*cicloFinalizado = true
		PCB.Estado = "BLOCK"
		IoGenSleep(variable[1], variable[2], registrosMap, PCB.PID)

	case "EXIT":
		*cicloFinalizado = true
		PCB.Estado = "EXIT"
		motivoDeDesalojo = "EXIT"

		return

	default:
		fmt.Println("------")
	}

	// Incrementa el Program Counter para apuntar a la siguiente instrucción
	*PC++
}

//---------------------FUNCIONES DE INSTRUCCIONES---------------------

// Asigna al registro el valor pasado como parámetro.
func set(reg string, dato string, registroMap map[string]*uint8, PC *uint32) {

	// Verifica si el registro a asignar es el PC
	if reg == "PC" {

		// Convierte el valor a un entero sin signo de 32 bits
		valorInt64, err := strconv.ParseUint(dato, 10, 32)

		if err != nil {
			fmt.Println("Dato no valido")
		}

		// Asigna el valor al PC (resta 1 ya que el PC se incrementará después de esta instrucción)
		*PC = uint32(valorInt64) - 1
		return
	}

	// Obtiene el puntero al registro del mapa de registros
	registro, encontrado := registroMap[reg]
	if !encontrado {
		fmt.Println("Registro invalido")
		return
	}

	// Convierte el valor de string a entero
	valor, err := strconv.Atoi(dato)

	if err != nil {
		fmt.Println("Dato no valido")
	}

	// Asigna el nuevo valor al registro
	*registro = uint8(valor)
}

// Suma al Registro Destino el Registro Origen y deja el resultado en el Registro Destino.
func sum(reg1 string, reg2 string, registroMap map[string]*uint8) {

	// Verifica si existen los registros especificados en la instrucción.
	registro1, encontrado := registroMap[reg1]
	if !encontrado {
		fmt.Println("Registro invalido")
		return
	}

	registro2, encontrado := registroMap[reg2]
	if !encontrado {
		fmt.Println("Registro invalido")
		return
	}

	// Suma el valor del Registro Origen al Registro Destino.
	*registro1 += *registro2

}

// Resta al Registro Destino el Registro Origen y deja el resultado en el Registro Destino.
func sub(reg1 string, reg2 string, registroMap map[string]*uint8) {

	// Verifica si existen los registros especificados en la instrucción.
	registro1, encontrado := registroMap[reg1]
	if !encontrado {
		fmt.Println("Registro invalido")
		return
	}

	registro2, encontrado := registroMap[reg2]
	if !encontrado {
		fmt.Println("Registro invalido")
		return
	}

	// Resta el valor del Registro Origen al Registro Destino.
	*registro1 -= *registro2
}

// Si el valor del registro es distinto de cero, actualiza el PC al numero de instruccion pasada por parametro.
func jnz(reg string, valor string, PC *uint32, registroMap map[string]*uint8) {

	// Verifica si existe el registro especificado en la instrucción.
	registro, encontrado := registroMap[reg]
	if !encontrado {
		fmt.Println("Registro invalido")
		return
	}

	// Convierte el valor de la instrucción a un uint64 bits.
	valorInt64, err := strconv.ParseUint(valor, 10, 32)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Disminuye el valor de la instrucción en uno para ajustarlo al índice del slice de instrucciones.
	nuevoValor := uint32(valorInt64) - 1

	// Si el valor del registro es distinto de cero, actualiza el PC al nuevo valor.
	if *registro != 0 {
		*PC = nuevoValor
	}
}

// Envía una request a Kernel con el nombre de una interfaz y las unidades de trabajo a multiplicar. No se hace nada con la respuesta.
func IoGenSleep(nombreInterfaz string, unitWorkTimeString string, registroMap map[string]*uint8, PID uint32) {

	// Convierte el tiempo de trabajo de la unidad de cadena a entero.
	unitWorkTime, err := strconv.Atoi(unitWorkTimeString)
	if err != nil {
		return
	}

	// Convierte la instrucción a formato JSON.
	body, err := json.Marshal(structs.InstruccionIO{
		PidDesalojado:  PID,
		NombreInterfaz: nombreInterfaz,
		Instruccion:    "IO_GEN_SLEEP",
		UnitWorkTime:   unitWorkTime,
	})
	if err != nil {
		return
	}

	// Envía la solicitud a Kernel.
	respuesta := config.Request(configJson.Port_Kernel, configJson.Ip_Kernel, "POST", "instruccion", body)

	//TODO: Implementar respuesta si es necesario.
	fmt.Print(respuesta)
}
