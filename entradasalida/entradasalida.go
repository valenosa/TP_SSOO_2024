package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

var Auxlogger *log.Logger

// *======================================| MAIN |======================================\\
func main() {

	// Configura el logger
	logueano.Logger("IO.log")

	Auxlogger = logueano.InitAuxLog("IO")

	//Toma los parametros pasados por argumento
	nombreInterfaz := os.Args[1]
	configPath := os.Args[2]

	config.Iniciar(configPath, &configInterfaz)

	//----------( LEVANTAMOS ARCHIVOS FS )----------

	if configInterfaz.Type == "DIALFS" {
		levantarFS(configInterfaz)
	}

	var cantBloquesDisponibles int = configInterfaz.Dialfs_Block_Count //?Esto funciona solamente si una vez que se termina la ejecución de la interfaz, se reinicia el sistema. Si hay permanencia, entonces cada vez que levantemos la interfaz deberíamos leer el bitmap.dat y contar la cantidad de bloques libres.

	//----------( INICIAMOS INTERFAZ )----------

	// Envio a Kernel la nueva interfaz
	conectarInterfazIO(nombreInterfaz)

	// Levanta el server de la nuevaInterfazIO
	serverErr := iniciarServidorInterfaz(&cantBloquesDisponibles)
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

func iniciarServidorInterfaz(cantBloquesDisponibles *int) error {

	http.HandleFunc("POST /GENERICA/IO_GEN_SLEEP", handlerIO_GEN_SLEEP)
	http.HandleFunc("POST /STDIN/IO_STDIN_READ", handlerIO_STDIN_READ)
	http.HandleFunc("POST /STDOUT/IO_STDOUT_WRITE", handlerIO_STDOUT_WRITE)
	http.HandleFunc("POST /DIALFS/IO_FS_CREATE", handlerIO_FS_CREATE(cantBloquesDisponibles)) //!Modificar la request desde kernel para que no ponga /TipoDeInstruccion

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

//*---------------( STDOUT )--------------------

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

// *---------------( DIALFS )--------------------

type metadata struct {
	InitialBlock int `json:"initial_block"`
	Size         int `json:"size"`
}

func handlerIO_FS_CREATE(cantBloquesDisponibles *int) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		mx_interfaz.Lock()
		defer mx_interfaz.Unlock()
		//--------- RECIBE ---------
		var instruccionIO structs.RequestEjecutarInstruccionIO
		err := json.NewDecoder(r.Body).Decode(&instruccionIO)
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Decodifica el request (codificado en formato json)
		// err := json.NewDecoder(r.Body).Decode(&instruccionIO)
		// if err != nil {
		// 	fmt.Println(err)
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		//-------- EJECUTA ---------

		nuevoArchivo, err := os.Create(configInterfaz.Dialfs_Path + "/" + instruccionIO.NombreArchivo)
		if err != nil {
			fmt.Println(err)
			return
		}

		defer nuevoArchivo.Close()

		var bloque int = asignarEspacio(cantBloquesDisponibles)

		//Data del archivo
		metadata := metadata{bloque, 0}

		//Se escribe en el archivo
		encoder := json.NewEncoder(nuevoArchivo)
		err = encoder.Encode(metadata)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Envía el status al Kernel
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(":/"))
	}
}

func asignarEspacio(cantBloquesDisponibles *int) int {

	file, err := os.OpenFile(configInterfaz.Dialfs_Path+"/"+"bitmap.dat", os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer file.Close()

	byteCount := 0
	buf := make([]byte, 1)
	for {
		_, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			return -1
		}

		if buf[0] == 0 {
			// Encontramos un byte que es 0
			Auxlogger.Println("Bloque libre: ", byteCount) //!LOG
			*cantBloquesDisponibles--
			Auxlogger.Println("Cantidad de Blq. Libres: ", *cantBloquesDisponibles) //!LOG

			// Escribimos un 1 en el byte que encontramos
			_, err = file.WriteAt([]byte{1}, int64(byteCount))
			if err != nil {
				fmt.Println(err)
				return -1
			}

			return byteCount
		}
		byteCount++
	}
	// No encontramos ningún byte que sea 0
	return -1
}

func levantarFS(configInterfaz config.IO) {

	//-------- BLOQUES.DAT ---------

	bloques, err := os.Create(configInterfaz.Dialfs_Path + "/bloques.dat")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer bloques.Close()

	// Establecer el tamaño del archivo
	err = bloques.Truncate(int64(configInterfaz.Dialfs_Block_Size * configInterfaz.Dialfs_Block_Count))
	if err != nil {
		fmt.Println(err)
		return
	}

	//-------- BITMAP.DAT ---------

	bitmap, err := os.Create(configInterfaz.Dialfs_Path + "/bitmap.dat")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer bitmap.Close()

	// Establecer el tamaño del archivo
	err = bitmap.Truncate(int64(configInterfaz.Dialfs_Block_Count)) //? Cada byte representa un bloque. Debería ser cada bit?
	if err != nil {
		fmt.Println(err)
		return
	}
}
