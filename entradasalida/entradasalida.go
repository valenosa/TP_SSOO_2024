package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

// Este mutex hace que se ejecute una sola instrucción de IO a la vez
var mx_interfaz sync.Mutex

//*======================================| MAIN |======================================\\

func main() {

	nombreInterfaz := os.Args[1]
	path := os.Args[2]

	// Configura el logger
	config.Logger("IO.log")

	// Crear interfaz (TESTING)
	conectarInterfazIO(nombreInterfaz, path)
}

//*======================================| FUNCIONES |======================================\\

func conectarInterfazIO(nombre string, filePath string) {

	// Extrae info de config.json
	var configNuevaInterfaz config.IO

	config.Iniciar(filePath, &configNuevaInterfaz)

	// Crea Interfaz base
	var nuevaInterfazIO = structs.Interfaz{TipoInterfaz: configNuevaInterfaz.Type, PuertoInterfaz: configNuevaInterfaz.Port}

	// Crea y codifica la request de conexion a Kernel
	var requestConectarIO = structs.RequestConectarInterfazIO{NombreInterfaz: nombre, Interfaz: nuevaInterfazIO}
	body, marshalErr := json.Marshal(requestConectarIO)
	if marshalErr != nil {
		fmt.Printf("error codificando body: %s", marshalErr.Error())
		return
	}

	// Si todo es correcto envia la request de conexion a Kernel
	respuesta := config.Request(configNuevaInterfaz.Port_Kernel, configNuevaInterfaz.Ip_Kernel, "POST", "interfazConectada", body)
	if respuesta == nil {
		return
	}

	// Levanta el server de la nuevaInterfazIO
	serverErr := iniciarServidorInterfaz(configNuevaInterfaz)
	if serverErr != nil {
		fmt.Printf("Error al iniciar servidor de interfaz: %s", serverErr.Error())
		return
	}
}

func iniciarServidorInterfaz(configInterfaz config.IO) error {

	http.HandleFunc("POST /IO_GEN_SLEEP", handlerIO_GEN_SLEEP(configInterfaz))
	http.HandleFunc("POST /IO_STDIN_READ", handlerIO_STDIN_READ)

	var err = config.IniciarServidor(configInterfaz.Port)
	return err
}

//*======================================| HANDLERS |======================================\\

// Implemantación de la Interfaz Génerica
func handlerIO_GEN_SLEEP(configIO config.IO) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		mx_interfaz.Lock()
		//Crea una variable tipo Interfaz (para interpretar lo que se recibe de la instruccionIO)
		var instruccionIO structs.RequestEjecutarInstruccionIO

		// Decodifica el request (codificado en formato json)
		err := json.NewDecoder(r.Body).Decode(&instruccionIO)

		// Error de la decodificación
		if err != nil {
			fmt.Println(err) ////! Borrar despues.
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Imprime el request por consola (del lado del server)
		fmt.Println("Unidades de Trabajo:", instruccionIO)

		//Ejecuta IO_GEN_SLEEP
		sleepTime := configIO.Unit_Work_Time * instruccionIO.UnitWorkTime
		fmt.Println(instruccionIO.PidDesalojado, " Zzzzzz...")
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)
		fmt.Println("Wakey wakey, ", instruccionIO.PidDesalojado, ", its time for school")
		// Responde al cliente
		w.WriteHeader(http.StatusOK)
		mx_interfaz.Unlock()
	}
}

func handlerIO_STDIN_READ(w http.ResponseWriter, r *http.Request) {
	//--------- REQUEST ---------

	//Crea una variable tipo Interfaz (para interpretar lo que se recibe de la instruccionIO)
	var instruccionIO structs.RequestEjecutarInstruccionIO

	// Decodifica el request (codificado en formato json)
	err := json.NewDecoder(r.Body).Decode(&instruccionIO)

	// Error de la decodificación
	if err != nil {
		fmt.Println(err) ////! Borrar despues.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//--------- EJECUTA IO_STDIN_READ ---------

	//Genera un lector de texto.
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Por favor ingresa un texto:")

	//Lee hasta que haya un salto de linea, y guarda el texto incluyendo '\n'
	input, err := reader.ReadString('\n')

	if err != nil {
		fmt.Println(err) ////! Borrar despues.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("EL MALDITO IO")

	// Eliminar el salto de línea al final de la cadena
	input = input[:len(input)-1]

	//--------- RESPUESTA ---------
	responseInputUsuario := structs.RequestInputSTDIN{TextoUsuario: input}

	respuesta, err := json.Marshal(responseInputUsuario)
	if err != nil {
		fmt.Println(err) //! Borrar despues.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Envía respuesta al cliente
	w.WriteHeader(http.StatusOK)
	w.Write(respuesta)
}
