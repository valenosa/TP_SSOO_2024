package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/logueano"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var mx_interfaz sync.Mutex // Mutex para Ejecutar las intrucciones IO en orden FIFO
var configInterfaz config.IO

var Auxlogger *logueano.AuxLogger

// Inicializa el logger auxiliar
func init() {
	var err error
	Auxlogger, err = logueano.NewLogger("entradasalida")
	if err != nil {
		panic(err)
	}
}

// *======================================| MAIN |======================================\\
func main() {

	// Configura el logger
	logueano.Logger("entradasalida.log")

	//Toma los parametros pasados por argumento
	nombreInterfaz := os.Args[1]
	configPath := os.Args[2]

	config.Iniciar(configPath, &configInterfaz)

	//----------( INICIAMOS INTERFAZ )----------

	// Envio a Kernel la nueva interfaz
	conectarInterfazIO(nombreInterfaz)

	// Levanta el server de la nuevaInterfazIO
	serverErr := iniciarServidorInterfaz()
	if serverErr != nil {
		fmt.Printf("Error al iniciar servidor de interfaz: %s", serverErr.Error())
		return
	}
}

//*======================================| CONEXION CON KERNEL |======================================\\

func conectarInterfazIO(nombre string) {

	// Crea Interfaz base
	var nuevaInterfazIO = structs.Interfaz{TipoInterfaz: configInterfaz.Type, PuertoInterfaz: configInterfaz.Port}

	// Crea y codifica la request de conexion a Kernel
	var requestConectarIO = structs.RequestConectarInterfazIO{NombreInterfaz: nombre, Interfaz: nuevaInterfazIO}
	body, marshalErr := json.Marshal(requestConectarIO)
	if marshalErr != nil {
		fmt.Printf("error codificando body: %s", marshalErr.Error())
		return
	}

	// Envia la request de conexion a Kernel
	_, err := config.Request(configInterfaz.Port_Kernel, configInterfaz.Ip_Kernel, "POST", "interfazConectada", body)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func iniciarServidorInterfaz() error {

	http.HandleFunc("POST /GENERICA/IO_GEN_SLEEP", handlerIO_GEN_SLEEP)
	http.HandleFunc("POST /STDIN/IO_STDIN_READ", handlerIO_STDIN_READ)
	http.HandleFunc("POST /STDOUT/IO_STDOUT_WRITE", handlerIO_STDOUT_WRITE)

	var err = config.IniciarServidor(configInterfaz.Port)
	return err
}

//*======================================| INTERFACES |======================================\\

//*---------------( GENERICA )------------------

func handlerIO_GEN_SLEEP(w http.ResponseWriter, r *http.Request) {

	mx_interfaz.Lock()

	//--------- RECIBE ---------

	var instruccionIO structs.RequestEjecutarInstruccionIO

	// Decodifica el request (codificado en formato json).
	err := json.NewDecoder(r.Body).Decode(&instruccionIO)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Unidades de Trabajo:", instruccionIO.UnitWorkTime) //! Borrar despues.

	//--------- EJECUTA ---------

	sleepTime := configInterfaz.Unit_Work_Time * instruccionIO.UnitWorkTime

	fmt.Println(instruccionIO.PidDesalojado, " Zzzzzz...")
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	fmt.Println("Wakey wakey, ", instruccionIO.PidDesalojado, ", its time for school")

	//--------- RESPUESTA ---------

	w.WriteHeader(http.StatusOK)
	mx_interfaz.Unlock()
}

//*---------------( STDIN )--------------------

func handlerIO_STDIN_READ(w http.ResponseWriter, r *http.Request) {

	mx_interfaz.Lock()

	//--------- RECIBE ---------
	var instruccionIO structs.RequestEjecutarInstruccionIO

	// Decodifica el request (codificado en formato json)
	err := json.NewDecoder(r.Body).Decode(&instruccionIO)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//--------- EJECUTA ---------

	// Prepara el reader para leer el input de la terminal
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Por favor ingresa un texto:")

	//Lee hasta que haya un salto de linea
	input, err := reader.ReadString('\n')

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Elimina el salto de línea al final de la cadena
	input = input[:len(input)-1]

	// Recorta la longitud del input en base al registroTamaño
	if len(input) > int(instruccionIO.Tamaño) {
		input = input[:instruccionIO.Tamaño]
	}

	//--------- REQUEST A MEMORIA ---------

	bodyWriteMemoria := structs.RequestMovOUT{
		Pid:  instruccionIO.PidDesalojado,
		Dir:  instruccionIO.Direccion,
		Data: []byte(input),
	}

	body, err := json.Marshal(bodyWriteMemoria)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Envía la request a memoria
	_, err = config.Request(configInterfaz.Port_Memory, configInterfaz.Ip_Memory, "POST", "memoria/movout", body) // TODO: Cambiar endpoint de la request a memoria
	if err != nil {
		fmt.Println(err)
		return
	}

	//--------- RESPUESTA ---------

	// Envía el status al Kernel
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(":)"))

	mx_interfaz.Unlock()

}

func handlerIO_STDOUT_WRITE(w http.ResponseWriter, r *http.Request) {
	mx_interfaz.Lock()

	//--------- RECIBE ---------
	var instruccionIO structs.RequestEjecutarInstruccionIO
	err := json.NewDecoder(r.Body).Decode(&instruccionIO)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//--------- REQUEST A MEMORIA ---------

	// Crea un cliente HTTP
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/memoria/movin", configInterfaz.Ip_Memory, configInterfaz.Port_Memory)

	// Crea una nueva solicitud GET
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	//Parsea la direccion física de uint32 a string.
	direccionFisicaStr := strconv.FormatUint(uint64(instruccionIO.Direccion), 10)
	pidEnEjecucionStr := strconv.FormatUint(uint64(instruccionIO.PidDesalojado), 10)
	longitud := strconv.FormatUint(uint64(instruccionIO.Tamaño), 10)

	// Agrega el PID y el PC como params
	q := req.URL.Query()
	q.Add("pid", pidEnEjecucionStr)
	q.Add("dir", direccionFisicaStr)
	q.Add("size", longitud)
	req.URL.RawQuery = q.Encode()

	// Establece el tipo de contenido de la solicitud
	req.Header.Set("Content-Type", "text/plain")

	// Realiza la solicitud al servidor de memoria
	respuesta, err := cliente.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	//--------- EJECUTA ---------

	data, err := io.ReadAll(respuesta.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var inputTruncado = string(data)

	// Muestra por la terminal el dato que se encontraba en la dirección enviada a memoria.
	fmt.Println(inputTruncado) //* No borrar, es parte de STDOUT.

	//--------- RESPUESTA ---------
	// Envía el status al Kernel
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(":("))

	mx_interfaz.Unlock()

}
