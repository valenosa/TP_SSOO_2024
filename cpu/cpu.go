package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

//================================| MAIN |================================\\

var registrosCPU structs.RegistrosUsoGeneral

var configJson config.Cpu

func main() {

	// Configura el logger
	config.Logger("CPU.log")

	log.Printf("Soy un logeano")

	// Se establece el handler que se utilizará para las diversas situaciones recibidas por el server
	// http.HandleFunc("PUT /plani", handlerIniciarPlanificacion)
	// http.HandleFunc("DELETE /plani", handlerDetenerPlanificacion)

	http.HandleFunc("POST /exec", handlerEjecutarProceso)
	http.HandleFunc("POST /interrupciones", handlerInterrupcion)

	// Extrae info de config.json
	config.Iniciar("config.json", &configJson)

	// declaro puerto
	port := ":" + strconv.Itoa(configJson.Port)

	//COMIENZO DEL HARDCODEO DEL DEVE.
	/*instruccion := IO_GEN_SLEEP{
		Instruccion:       "IO_GEN_SLEEP",
		NombreInterfaz:    "GenericIO",
		UnidadesDeTrabajo: 10,
	}*/

	//enviarInstruccionIO_GEN_SLEEP(instruccion)
	//FINAL DEL HARDCODEO DEL DEVE.

	// Listen and serve con info del config.json
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Error al esuchar en el puerto " + port)
	}
}

// -------------------------- HANDLERS -----------------------------------
// Funcion test para enviar una instruccion leída al kernel.
/*
func enviarInstruccionIO_GEN_SLEEP(instruccion IO_GEN_SLEEP) {
	body, err := json.Marshal(instruccion)

	//Check si no hay errores al crear el body.
	if err != nil {
		fmt.Printf("error codificando body: %s", err.Error())
		return
	}

	Mandar a ejecutar a la interfaz (Puerto)
	respuesta := config.Request(config.Kernel.Port, config.Kernel. , "POST", "/instruccion", body)

	if respuesta == nil{
		fmt.Println("Fallo en el envío de instrucción desde CPU a Kernel.")
	}
}*/

func handlerIniciarPlanificacion(w http.ResponseWriter, r *http.Request) {

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

func handlerDetenerPlanificacion(w http.ResponseWriter, r *http.Request) {

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

// Contiene el pid del proceso que dispatch mandó a ejecutar
var pidEnEjecucion uint32

// TODO: Hay que pasarla a local
var motivoDeDesalojo string

func handlerEjecutarProceso(w http.ResponseWriter, r *http.Request) {
	// Crea uan variable tipo BodyIniciar (para interpretar lo que se recibe de la pcbRecibido)
	var pcbRecibido structs.PCB

	// Decodifica el request (codificado en formato json)
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

	//Itera el ciclo de instruccion si hay instrucciones a ejecutar y no hay interrupciones
	for !hayInterrupcion && !cicloFinalizado {
		instruccion := fetch(PCB.PID, registrosCPU.PC)
		decodeAndExecute(PCB, instruccion, &registrosCPU.PC, &cicloFinalizado)
	}
	PCB.RegistrosUsoGeneral = registrosCPU

}

// Trae de memoria las instrucciones indicadas por el PC y el PID.
func fetch(PID uint32, PC uint32) string {

	// Se pasan PID y PC a string
	pid := strconv.FormatUint(uint64(PID), 10)
	pc := strconv.FormatUint(uint64(PC), 10)

	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/instrucciones", configJson.Ip_Memory, configJson.Port_Memory)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}

	//Paso como parametros pid y pc.
	q := req.URL.Query()
	q.Add("PID", pid)
	q.Add("PC", pc)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "text/plain")
	respuesta, err := cliente.Do(req)
	if err != nil {
		return ""
	}

	// Verificar el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		return ""
	}

	bodyBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return ""
	}

	return string(bodyBytes)
}

// Ejecuta las instrucciones traidas de memoria.
func decodeAndExecute(PCB *structs.PCB, instruccion string, PC *uint32, cicloFinalizado *bool) {

	var registrosMap = map[string]*uint8{
		"AX": &registrosCPU.AX,
		"BX": &registrosCPU.BX,
		"CX": &registrosCPU.CX,
		"DX": &registrosCPU.DX,
	}

	//Parsea las instrucciones de string a string[]
	variable := strings.Split(instruccion, " ")

	fmt.Println("Instruccion: ", variable[0], " Parametros: ", variable[1:])

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

	*PC++
}

//---------------------FUNCIONES DE INSTRUCCIONES---------------------

// Asigna al registro el valor pasado como parámetro.
func set(reg string, dato string, registroMap map[string]*uint8, PC *uint32) {

	//Checkea si existe el registro obtenido de la instruccion.
	if reg == "PC" {

		valorInt64, err := strconv.ParseUint(dato, 10, 32)

		if err != nil {
			fmt.Println("Dato no valido")
		}

		*PC = uint32(valorInt64) - 1
		return
	}

	registro, encontrado := registroMap[reg]
	if !encontrado {
		fmt.Println("Registro invalido")
		return
	}

	//Parsea string a entero el valor que va a tomar el registro.
	valor, err := strconv.Atoi(dato)

	if err != nil {
		fmt.Println("Dato no valido")
	}

	//Asigna el nuevo valor al registro.
	*registro = uint8(valor)
}

// Suma al Registro Destino el Registro Origen y deja el resultado en el Registro Destino.
func sum(reg1 string, reg2 string, registroMap map[string]*uint8) {
	//Checkea si existe el registro obtenido de la instruccion.
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

	*registro1 += *registro2

}

// Resta al Registro Destino el Registro Origen y deja el resultado en el Registro Destino.
func sub(reg1 string, reg2 string, registroMap map[string]*uint8) {
	//Checkea si existe el registro obtenido de la instruccion.
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

	*registro1 -= *registro2
}

// Si el valor del registro es distinto de cero, actualiza el PC al numero de instruccion pasada por parametro.
func jnz(reg string, valor string, PC *uint32, registroMap map[string]*uint8) {
	registro, encontrado := registroMap[reg]
	if !encontrado {
		fmt.Println("Registro invalido")
		return
	}

	valorInt64, err := strconv.ParseUint(valor, 10, 32)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	nuevoValor := uint32(valorInt64) - 1

	if *registro != 0 {
		*PC = nuevoValor
	}
}

// Envía una request a Kernel con el nombre de una interfaz y las unidades de trabajo a multiplicar. No se hace nada con la respuesta.
func IoGenSleep(nombreInterfaz string, unitWorkTimeString string, registroMap map[string]*uint8, PID uint32) {

	// int(unitWorkTime)
	unitWorkTime, err := strconv.Atoi(unitWorkTimeString)
	if err != nil {
		return
	}

	//Pasa la instruccion a formato JSON.
	body, err := json.Marshal(structs.InstruccionIO{
		PidDesalojado:  PID,
		NombreInterfaz: nombreInterfaz,
		Instruccion:    "IO_GEN_SLEEP",
		UnitWorkTime:   unitWorkTime,
	})
	if err != nil {
		return
	}

	//Envía la request
	respuesta := config.Request(configJson.Port_Kernel, configJson.Ip_Kernel, "POST", "instruccion", body)

	//TODO: Implementar respuesta si es necesario.
	fmt.Print(respuesta)
}
