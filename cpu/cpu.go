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
	ejecutarCiclosDeInstruccion(pcbRecibido)
	fmt.Println("Se está ejecutando el proceso: ", pcbRecibido.PID)

	// Responde que se terminó de ejecutar el proceso (respuesta en caso de que se haya podido terminar de ejecutar el proceso)
	var respBody string = "Se termino de ejecutar el proceso: " + strconv.FormatUint(uint64(pcbRecibido.PID), 10) + "\n" //el choclo este convierte uint32 a string

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

func ejecutarCiclosDeInstruccion(PCB proceso.PCB) {

	for true {
		text := fetch(PCB.PID, registrosCPU.PC)
		decodeAndExecute(text)
	}
}

func fetch(PID uint32, PC uint32) string {

	// Se pasan PID y PC a string
	pid := strconv.FormatUint(uint64(PID), 10)
	pc := strconv.FormatUint(uint64(PC), 10)

	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%s/instrucciones", configJson.Ip_Memory, configJson.Port_Memory)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}

	q := req.URL.Query()
	q.Add("PID", pid)
	q.Add("PC", pc)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
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

func decodeAndExecute(text string) {
	variable := strings.Split(text, " ")
	fmt.Println("Instrucción: ", variable[0], " Parametros: ", variable[1:])
	switch variable[0] {
	case "SET":
		// set(variable[1], variable[2], registrosMap)
		fmt.Println("------------------------")
	default:
		fmt.Println("------")
	}
}
