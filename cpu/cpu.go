package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/sisoputnfrba/tp-golang/utils/APIs/kernel-memoria/proceso"
	"github.com/sisoputnfrba/tp-golang/utils/config"
)

// ================================| MAIN |===================================================\\

type IO_GEN_SLEEP struct {
	Instruccion       string
	NombreInterfaz    string
	UnidadesDeTrabajo int
}

var registrosCPU proceso.RegistrosUsoGeneral

var configJson config.Cpu

func main() {

	// Configura el logger
	config.Logger("CPU.log")

	log.Printf("Soy un logeano")

	// Se establece el handler que se utilizará para las diversas situaciones recibidas por el server
	// http.HandleFunc("PUT /plani", handlerIniciarPlanificacion)
	// http.HandleFunc("DELETE /plani", handlerDetenerPlanificacion)

	http.HandleFunc("POST /exec", handlerEjecutarProceso)

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

func handlerEjecutarProceso(w http.ResponseWriter, r *http.Request) {
	// Crea uan variable tipo BodyIniciar (para interpretar lo que se recibe de la pcbRecibido)
	var pcbRecibido proceso.PCB

	// Decodifica el request (codificado en formato json)
	err := json.NewDecoder(r.Body).Decode(&pcbRecibido)

	// Error Handler de la decodificación
	if err != nil {
		fmt.Printf("Error al decodificar request body: ")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Simula ejecutar el proceso
	ejecutarCiclosDeInstruccion(&pcbRecibido)

	fmt.Println("Se está ejecutando el proceso: ", pcbRecibido.PID)

	// Falta devolver tambien el motivo de desalojo. Tenemos que charlar cuales son los motivos y como lo almacenamos.
	// Codificar Response en un array de bytes (formato json)
	respuesta, err := json.Marshal(pcbRecibido)

	// Error Handler de la codificación
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Envía respuesta (con estatus como header) al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}

func ejecutarCiclosDeInstruccion(PCB *proceso.PCB) {
	var cicloFInalizado bool = false

	for {
		instruccion := fetch(PCB.PID, registrosCPU.PC)
		decodeAndExecute(PCB, instruccion, &registrosCPU.PC, &cicloFInalizado)
		if cicloFInalizado {
			break
		}
		checkInterrupt()
	}
}

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

func decodeAndExecute(PCB *proceso.PCB, instruccion string, PC *uint32, cicloFinalizado *bool) {

	var registrosMap = map[string]*uint8{
		"AX": &registrosCPU.AX,
		"BX": &registrosCPU.BX,
		"CX": &registrosCPU.CX,
		"DX": &registrosCPU.DX,
	}

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
		IoGenSleep(variable[1], variable[2], registrosMap)

	case "EXIT":
		*cicloFinalizado = true
		PCB.RegistrosUsoGeneral = registrosCPU

		//El estado debería pasar a EXIT acá o en Kernel?
		estadoAExit := func() {
			PCB.Estado = "EXIT"
		}

		defer estadoAExit()

		return

	default:
		fmt.Println("------")
	}

	*PC++
}

//---------------------FUNCIONES DE INSTRUCCIONES---------------------

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

type InstruccionIO struct {
	NombreInterfaz string
	Instruccion    string
	UnitWorkTime   int
}

// Envía una request a kernel con el nombre de una interfaz y las unidades de trabajo a multiplicar. No se hace nada con la respuesta.
func IoGenSleep(nombreInterfaz string, unitWorkTimeString string, registroMap map[string]*uint8) {
	//Variable 1: Nombre interfaz
	//Variable 2: UnitWorkTime
	//En segundo lugar va "IO_GEN_SLEEP"

	//Pasa de unitWorkTime de string a int con ATOI
	unitWorkTime, err := strconv.Atoi(unitWorkTimeString)
	if err != nil {
		return
	}

	//Confecciona el body.
	body, err := json.Marshal(InstruccionIO{
		NombreInterfaz: nombreInterfaz,
		Instruccion:    "IO_GEN_SLEEP",
		UnitWorkTime:   unitWorkTime,
	})
	if err != nil {
		return
	}
	//Envía la request
	respuesta := config.Request(configJson.Port_Kernel, configJson.Ip_Kernel, "POST", "instruccion", body)

	// Verificar que no hubo error en la request
	if respuesta == nil {
		return
	}
	//Que hacer con la respuesta????
}

func checkInterrupt() {
	//Checkea si hay interrupciones.
}
