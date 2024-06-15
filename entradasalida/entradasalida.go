package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

var mx_interfaz sync.Mutex // Mutex para Ejecutar las intrucciones IO en orden FIFO
var configInterfaz config.IO

// *======================================| MAIN |======================================\\
func main() {

	// Configura el logger
	config.Logger("IO.log")

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
	respuesta := config.Request(configInterfaz.Port_Kernel, configInterfaz.Ip_Kernel, "POST", "interfazConectada", body)
	if respuesta == nil {
		return
	}
}

func iniciarServidorInterfaz() error {

	http.HandleFunc("POST /GENERICA /IO_GEN_SLEEP", handlerIO_GEN_SLEEP)
	http.HandleFunc("POST /STDIN/IO_STDIN_READ", handlerIO_STDIN_READ)
	http.HandleFunc("POST /STDIN/IO_STDOUT_WRITE", handlerIO_STDOUT_WRITE)

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
		fmt.Println(err) //! Borrar despues.
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
	inputTruncado := input[0:instruccionIO.RegistroTamaño]

	//--------- REQUEST A MEMORIA ---------

	bodyWriteMemoria := structs.RequestInputSTDIN{
		Pid:               instruccionIO.PidDesalojado,
		TextoUsuario:      []byte(inputTruncado),
		RegistroDireccion: instruccionIO.RegistroDireccion,
	}

	body, err := json.Marshal(bodyWriteMemoria)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Envía la request a memoria
	respuesta := config.Request(configInterfaz.Port_Memory, configInterfaz.Ip_Memory, "POST", "/movout", body) // TODO: Cambiar endpoint de la request a memoria
	if respuesta == nil {
		return
	}

	marcoBytes, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return
	}

	// Convierte el valor de la instrucción a un uint64 bits.
	testInput := string(marcoBytes)

	fmt.Println("Mi input es: ", testInput)

	//--------- RESPUESTA ---------

	// Envía el status al Kernel
	w.WriteHeader(http.StatusOK)
	mx_interfaz.Unlock()

}

func handlerIO_STDOUT_WRITE(w http.ResponseWriter, r *http.Request) {
	mx_interfaz.Lock()

	//--------- RECIBE ---------
	var instruccionIO structs.RequestEjecutarInstruccionIO

	// Decodifica el request (codificado en formato json)
	err := json.NewDecoder(r.Body).Decode(&instruccionIO)
	if err != nil {
		fmt.Println(err) //! Borrar despues.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//--------- REQUEST A MEMORIA ---------

	body, err := json.Marshal(instruccionIO.RegistroDireccion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Envía la request a memoria
	respuesta := config.Request(configInterfaz.Port_Memory, configInterfaz.Ip_Memory, "POST", "/stdout", body) // TODO: Cambiar endpoint de la request a memoria
	if respuesta == nil {
		return
	}

	//--------- EJECUTA ---------
	var inputTruncado string
	err = json.NewDecoder(respuesta.Body).Decode(&inputTruncado)
	if err != nil {
		http.Error(w, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		return
	}

	// Recorta la longitud del input en base al registroTamaño
	inputTruncado = inputTruncado[0:instruccionIO.RegistroTamaño]

	// Muestra por la terminal el dato que se encontraba en la dirección enviada a memoria.
	fmt.Println(inputTruncado) //* No borrar, es parte de STDOUT.

	//--------- RESPUESTA ---------
	// Envía el status al Kernel
	w.WriteHeader(http.StatusOK)
	mx_interfaz.Unlock()
}
