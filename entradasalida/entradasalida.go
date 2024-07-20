package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
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

var listaArchivos []string

var FS_totalBloquesDisponibles int

// *=====================================| MAIN |=====================================\\
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

	//----------( INICIAMOS INTERFAZ )----------

	// Envio a Kernel la nueva interfaz
	conectarInterfazIO(nombreInterfaz)

	// Levanta el server de la nuevaInterfazIO
	serverErr := iniciarServidorInterfaz()
	if serverErr != nil {
		logueano.Error(Auxlogger, serverErr)
		return
	}
}

//*======================================| CONEXION CON KERNEL |======================================\\

func conectarInterfazIO(nombre string) {

	// Crea Interfaz base
	var nuevaInterfazIO = structs.Interfaz{TipoInterfaz: configInterfaz.Type, PuertoInterfaz: configInterfaz.Port, IpInterfaz: configInterfaz.Ip}

	// Crea y codifica la request de conexion a Kernel
	var requestConectarIO = structs.RequestConectarInterfazIO{NombreInterfaz: nombre, Interfaz: nuevaInterfazIO}
	body, err := json.Marshal(requestConectarIO)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	// Envia la request de conexion a Kernel
	_, err = config.Request(configInterfaz.Port_Kernel, configInterfaz.Ip_Kernel, "POST", "interfazConectada", body)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}
}

func iniciarServidorInterfaz() error {

	//GENERICA
	http.HandleFunc("POST /GENERICA/IO_GEN_SLEEP", handlerIO_GEN_SLEEP)

	//STDIN
	http.HandleFunc("POST /STDIN/IO_STDIN_READ", handlerIO_STDIN_READ)

	//STDOUT
	http.HandleFunc("POST /STDOUT/IO_STDOUT_WRITE", handlerIO_STDOUT_WRITE)

	//DIALFS
	http.HandleFunc("POST /DIALFS/IO_FS_CREATE", handlerIO_FS_CREATE)
	http.HandleFunc("POST /DIALFS/IO_FS_TRUNCATE", handlerIO_FS_TRUNCATE)
	http.HandleFunc("POST /DIALFS/IO_FS_DELETE", handlerIO_FS_DELETE)
	http.HandleFunc("POST /DIALFS/IO_FS_WRITE", handlerIO_FS_WRITE)
	http.HandleFunc("POST /DIALFS/IO_FS_READ", handlerIO_FS_READ)

	var err = config.IniciarServidor(configInterfaz.Port)
	return err
}

//*======================================| INTERFACES |======================================\\

//*======================( GENERICA )======================

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

	//--------- EJECUTA ---------

	//^ log obligatorio (1/6)
	logueano.Operacion(instruccionIO.PidDesalojado, "IO_GEN_SLEEP")

	sleepTime := configInterfaz.Unit_Work_Time * instruccionIO.UnitWorkTime

	fmt.Println(instruccionIO.PidDesalojado, " Zzzzzz...")
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	fmt.Println("Wakey wakey, ", instruccionIO.PidDesalojado, ", its time for school")

	//--------- RESPUESTA ---------

	w.WriteHeader(http.StatusOK)
	mx_interfaz.Unlock()
}

//*======================( STDIN )======================

func handlerIO_STDIN_READ(w http.ResponseWriter, r *http.Request) {

	mx_interfaz.Lock()

	//--------- RECIBE ---------
	var instruccionIO structs.RequestEjecutarInstruccionIO

	// Decodifica el request (codificado en formato json)
	err := json.NewDecoder(r.Body).Decode(&instruccionIO)
	if err != nil {
		logueano.Error(Auxlogger, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//--------- EJECUTA ---------

	//^ log obligatorio (1/6)
	logueano.Operacion(instruccionIO.PidDesalojado, "IO_STDIN_READ")

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
	_, err = config.Request(configInterfaz.Port_Memory, configInterfaz.Ip_Memory, "POST", "memoria/movout", body)
	if err != nil {
		logueano.Error(Auxlogger, err)
		http.Error(w, "INVALID_WRITE", http.StatusBadRequest)
		return
	}

	//--------- RESPUESTA ---------

	// Envía el status al Kernel
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(input))

	mx_interfaz.Unlock()
}

//*======================( STDOUT )======================

func handlerIO_STDOUT_WRITE(w http.ResponseWriter, r *http.Request) {
	mx_interfaz.Lock()

	//--------- RECIBE ---------
	var instruccionIO structs.RequestEjecutarInstruccionIO

	err := json.NewDecoder(r.Body).Decode(&instruccionIO)
	if err != nil {
		logueano.Error(Auxlogger, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//--------- REQUEST A MEMORIA ---------
	//^ log obligatorio (1/6)
	logueano.Operacion(instruccionIO.PidDesalojado, "IO_STDOUT_WRITE")

	// Crea un cliente HTTP
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/memoria/movin", configInterfaz.Ip_Memory, configInterfaz.Port_Memory)

	// Crea una nueva solicitud GET
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logueano.Error(Auxlogger, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		logueano.Error(Auxlogger, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//--------- EJECUTA ---------

	data, err := io.ReadAll(respuesta.Body)
	if err != nil {
		logueano.Error(Auxlogger, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Espera una unidad de trabajo
	time.Sleep(time.Duration(configInterfaz.Unit_Work_Time) * time.Millisecond)

	var inputTruncado = string(data)

	// Muestra por la terminal el dato que se encontraba en la dirección enviada a memoria.
	fmt.Println(inputTruncado) //* No borrar, es parte de STDOUT.

	//--------- RESPUESTA ---------
	// Envía el status al Kernel
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(inputTruncado))

	mx_interfaz.Unlock()
}

//*======================( DIALFS )======================

// Crea/Analiza los archivos necesarios para el FS
func levantarFS(configInterfaz config.IO) {

	//-------- BLOQUES.DAT ---------

	_, err := os.Stat(configInterfaz.Dialfs_Path + "/bloques.dat")
	if os.IsNotExist(err) {
		bloques, err := os.Create(configInterfaz.Dialfs_Path + "/bloques.dat")
		if err != nil {
			logueano.Error(Auxlogger, err)
			return
		}

		// Establecer el tamaño del archivo
		err = bloques.Truncate(int64(configInterfaz.Dialfs_Block_Size * configInterfaz.Dialfs_Block_Count))
		if err != nil {
			logueano.Error(Auxlogger, err)
			return
		}

		bloques.Close()
	}

	//-------- BITMAP.DAT ---------

	_, err = os.Stat(configInterfaz.Dialfs_Path + "/bitmap.dat")
	if err == nil {

		//-------------- Cuento la cantidad de bloques libres

		bitmap, err := os.ReadFile(configInterfaz.Dialfs_Path + "/bitmap.dat")
		if err != nil {
			logueano.Error(Auxlogger, err)
			return
		}

		for _, b := range bitmap {
			if b == 0 {
				FS_totalBloquesDisponibles++
			}
		}

		//-------------- Agrego a la lista de archivos los archivos existentes

		archivos, err := os.ReadDir(configInterfaz.Dialfs_Path)
		if err != nil {
			logueano.Error(Auxlogger, err)
		}

		for _, archivo := range archivos {
			if archivo.Name() == "bitmap.dat" || archivo.Name() == "bloques.dat" {
				continue
			}
			listaArchivos = append(listaArchivos, archivo.Name())
		}

		return
	}

	if os.IsNotExist(err) {

		bitmap, err := os.Create(configInterfaz.Dialfs_Path + "/bitmap.dat")
		if err != nil {
			logueano.Error(Auxlogger, err)
			return
		}

		// Establecer el tamaño del archivo
		err = bitmap.Truncate(int64(configInterfaz.Dialfs_Block_Count))
		if err != nil {
			logueano.Error(Auxlogger, err)
			return
		}

		FS_totalBloquesDisponibles = configInterfaz.Dialfs_Block_Count

		bitmap.Close()

		return
	}
}

func handlerIO_FS_CREATE(w http.ResponseWriter, r *http.Request) {

	mx_interfaz.Lock()
	defer mx_interfaz.Unlock()

	//--------- RECIBE ---------
	var instruccionIO structs.RequestEjecutarInstruccionIO

	err := json.NewDecoder(r.Body).Decode(&instruccionIO)
	if err != nil {
		logueano.Error(Auxlogger, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//--------- EJECUTA ---------

	//^ log obligatorio (1/6)
	logueano.Operacion(instruccionIO.PidDesalojado, "IO_FS_CREATE")

	//Espera una unidad de trabajo
	time.Sleep(time.Duration(configInterfaz.Unit_Work_Time) * time.Millisecond)

	if FS_totalBloquesDisponibles == 0 {
		//No hay espacio en disco
		logueano.Mensaje(Auxlogger, "No hay espacio en disco")
		return
	}

	metadata := structs.MetadataFS{}

	//Asigna su primer bloque
	metadata.InitialBlock = asignarEspacio()

	metadata.Size = int(instruccionIO.Tamaño)

	//Establece a partir de que bloque puede escribir, y el tamaño máximo del archivo.
	actualizarMetadata(instruccionIO.NombreArchivo, metadata)

	//Agrega el nombre del archivo a la listaArchivo
	listaArchivos = append(listaArchivos, instruccionIO.NombreArchivo)

	//^ log obligatorio (2/6)
	logueano.CrearArchivo(instruccionIO.PidDesalojado, instruccionIO.NombreArchivo)

	//--------- RESPUESTA ---------
	// Envía el status al Kernel
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(":*"))
}

func handlerIO_FS_TRUNCATE(w http.ResponseWriter, r *http.Request) {

	mx_interfaz.Lock()
	defer mx_interfaz.Unlock()
	//--------- RECIBE ---------
	var instruccionIO structs.RequestEjecutarInstruccionIO

	err := json.NewDecoder(r.Body).Decode(&instruccionIO)
	if err != nil {
		logueano.Error(Auxlogger, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//-------- EJECUTA ---------

	//^ log obligatorio (1/6)
	logueano.Operacion(instruccionIO.PidDesalojado, "IO_FS_TRUNCATE")

	//^ log obligatorio (4/6)
	logueano.TruncarArchivo(instruccionIO.PidDesalojado, instruccionIO.NombreArchivo, instruccionIO.Tamaño)

	//Espera una unidad de trabajo
	time.Sleep(time.Duration(configInterfaz.Unit_Work_Time) * time.Millisecond)

	//Extrae el tamaño y el bloque inicial del archivo recibido.
	metadata, err := extraerMetadata(instruccionIO.NombreArchivo)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	tamañoEnBloques := calcularTamañoEnBloques(metadata.Size)

	nuevoTamañoEnBloques := calcularTamañoEnBloques(int(instruccionIO.Tamaño))

	if nuevoTamañoEnBloques > tamañoEnBloques {
		agrandarArchivo(nuevoTamañoEnBloques, &metadata, tamañoEnBloques, instruccionIO.NombreArchivo)
	}
	if nuevoTamañoEnBloques < tamañoEnBloques {
		achicarArchivo(nuevoTamañoEnBloques, tamañoEnBloques, metadata)
	}

	metadata.Size = int(instruccionIO.Tamaño)
	actualizarMetadata(instruccionIO.NombreArchivo, metadata)

	//--------- RESPUESTA ---------
	// Envía el status al Kernel
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(":P"))
}

func handlerIO_FS_DELETE(w http.ResponseWriter, r *http.Request) {

	mx_interfaz.Lock()
	defer mx_interfaz.Unlock()

	//--------- RECIBE ---------

	var instruccionIO structs.RequestEjecutarInstruccionIO

	err := json.NewDecoder(r.Body).Decode(&instruccionIO)
	if err != nil {
		logueano.Error(Auxlogger, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//-------- EJECUTA ---------

	//^ log obligatorio (1/6)
	logueano.Operacion(instruccionIO.PidDesalojado, "IO_FS_DELETE")

	//Espera una unidad de trabajo
	time.Sleep(time.Duration(configInterfaz.Unit_Work_Time) * time.Millisecond)

	//Extraigo la metadata
	metadata, err := extraerMetadata(instruccionIO.NombreArchivo)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	//Libero los bloques
	liberarBloques(calcularTamañoEnBloques(metadata.Size), 0, metadata.InitialBlock)

	//Elimino la metadata
	err = os.Remove(configInterfaz.Dialfs_Path + "/" + instruccionIO.NombreArchivo)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	//Elimino el archivo de la lista de archivos
	for i, archivo := range listaArchivos {
		if archivo == instruccionIO.NombreArchivo {
			//^ log obligatorio (2/6)
			logueano.EliminarArchivo(instruccionIO.PidDesalojado, instruccionIO.NombreArchivo)

			listaArchivos = append(listaArchivos[:i], listaArchivos[i+1:]...)
			break
		}
	}

	//--------- RESPUESTA ---------

	// Envía el status al Kernel
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(":/"))
}

func handlerIO_FS_WRITE(w http.ResponseWriter, r *http.Request) {

	mx_interfaz.Lock()
	defer mx_interfaz.Unlock()

	//--------- RECIBE ---------
	var instruccionIO structs.RequestEjecutarInstruccionIO

	err := json.NewDecoder(r.Body).Decode(&instruccionIO)
	if err != nil {
		logueano.Error(Auxlogger, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//--------- REQUEST A MEMORIA ---------
	//^ log obligatorio (1/6)
	logueano.Operacion(instruccionIO.PidDesalojado, "IO_FS_WRITE")

	//^ log obligatorio (6/6)
	logueano.LeerEscribirArchivo(instruccionIO.PidDesalojado, "ESCRIBIR", instruccionIO.NombreArchivo, int(instruccionIO.Tamaño), instruccionIO.PunteroArchivo)

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
		logueano.Error(Auxlogger, err)
		return
	}

	//--------- EJECUTA ---------
	data, err := io.ReadAll(respuesta.Body)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	//Extraigo la metadata
	metadata, err := extraerMetadata(instruccionIO.NombreArchivo)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	//Valido que no se pase del limite del archivo
	if uint32(calcularTamañoEnBloques(metadata.Size)*configInterfaz.Dialfs_Block_Size) < instruccionIO.PunteroArchivo+instruccionIO.Tamaño {
		logueano.Mensaje(Auxlogger, "Error: No se puede escribir en el archivo, sobrepasa el tamaño")
		http.Error(w, "No se puede escribir en el archivo, sobrepasa", http.StatusConflict)
		return
	}

	//Escribo el inputTruncado en la pos de bloques.dat correspondiente + RegistroPunteroArchivo
	fDataBloques, err := os.OpenFile(configInterfaz.Dialfs_Path+"/"+"bloques.dat", os.O_RDWR, 0644)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	defer fDataBloques.Close()

	_, err = fDataBloques.WriteAt(data, int64(uint32(metadata.InitialBlock*configInterfaz.Dialfs_Block_Size)+instruccionIO.PunteroArchivo))
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	//--------- RESPUESTA ---------
	// Envía el status al Kernel
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(":|"))
}

func handlerIO_FS_READ(w http.ResponseWriter, r *http.Request) {

	mx_interfaz.Lock()
	defer mx_interfaz.Unlock()
	//--------- RECIBE ---------
	var instruccionIO structs.RequestEjecutarInstruccionIO

	err := json.NewDecoder(r.Body).Decode(&instruccionIO)
	if err != nil {
		logueano.Error(Auxlogger, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//-------- EJECUTA ---------

	//^ log obligatorio (1/6)
	logueano.Operacion(instruccionIO.PidDesalojado, "IO_FS_READ")

	//^ log obligatorio (6/6)
	logueano.LeerEscribirArchivo(instruccionIO.PidDesalojado, "LEER", instruccionIO.NombreArchivo, int(instruccionIO.Tamaño), instruccionIO.PunteroArchivo)

	//Extraigo la metadata
	metadata, err := extraerMetadata(instruccionIO.NombreArchivo)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	//Valido que no se pase del limite del archivo
	if uint32(calcularTamañoEnBloques(metadata.Size)*configInterfaz.Dialfs_Block_Size) < instruccionIO.PunteroArchivo+instruccionIO.Tamaño {
		logueano.Mensaje(Auxlogger, "Error: No se puede leer del archivo, sobrepasa el tamaño")
		http.Error(w, "No se puede leer del archivo, sobrepasa el tamaño", http.StatusConflict)
		return
	}

	//Abro el archivo de DataBloques
	fDataBloques, err := os.OpenFile(configInterfaz.Dialfs_Path+"/"+"bloques.dat", os.O_RDWR, 0644)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	//Leo el archivo de bloques
	buffer := make([]byte, instruccionIO.Tamaño)
	_, err = fDataBloques.ReadAt(buffer, int64(uint32(metadata.InitialBlock*configInterfaz.Dialfs_Block_Size)+instruccionIO.PunteroArchivo))
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	fDataBloques.Close()
	//--------- REQUEST A MEMORIA ---------

	bodyWriteMemoria := structs.RequestMovOUT{
		Pid:  instruccionIO.PidDesalojado,
		Dir:  instruccionIO.Direccion,
		Data: buffer,
	}

	body, err := json.Marshal(bodyWriteMemoria)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Envía la request a memoria
	_, err = config.Request(configInterfaz.Port_Memory, configInterfaz.Ip_Memory, "POST", "memoria/movout", body)
	if err != nil {
		http.Error(w, "INVALID_WRITE", http.StatusBadRequest)
		return
	}
	//--------- RESPUESTA ---------
	// Envía el status al Kernel
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(":o"))
}

// ------ AUXILIARES DE DIALFS ------

// ------ METADATA ------

func extraerMetadata(nombreArchivo string) (structs.MetadataFS, error) {
	file, err := os.Open(configInterfaz.Dialfs_Path + "/" + nombreArchivo)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return structs.MetadataFS{}, err
	}

	defer file.Close()
	// Lee todo el contenido del archivo
	bytes, err := io.ReadAll(file)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return structs.MetadataFS{}, err
	}

	// Deserializa los bytes a la estructura Metadata
	var metadata structs.MetadataFS
	err = json.Unmarshal(bytes, &metadata)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return structs.MetadataFS{}, err
	}
	return metadata, nil
}

func actualizarMetadata(nombreArchivo string, nuevaMetadata structs.MetadataFS) {

	nuevoArchivo, err := os.OpenFile(configInterfaz.Dialfs_Path+"/"+nombreArchivo, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	defer nuevoArchivo.Close()

	//Se escribe en el archivo
	encoder := json.NewEncoder(nuevoArchivo).Encode(nuevaMetadata)
	err = encoder
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}
}

func calcularTamañoEnBloques(tamañoEnBytes int) int {

	if tamañoEnBytes != 0 {
		return int(math.Ceil(float64(tamañoEnBytes) / float64(configInterfaz.Dialfs_Block_Size)))
	} else {
		return 1
	}
}

// ------ IO_FS_CREATE ------

// En base al bitmap devuelve el primer bloque libre.
func asignarEspacio() int {

	file, err := os.OpenFile(configInterfaz.Dialfs_Path+"/"+"bitmap.dat", os.O_RDWR, 0644)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return -1
	}
	defer file.Close()

	//Busco la primera aparicion de un bloque libre (byte en 0)
	byteCount := 0
	buf := make([]byte, 1)
	for {
		_, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				logueano.Mensaje(Auxlogger, "EOF en asignarEspacio().")
				break
			}
			logueano.Error(Auxlogger, err)
			return -1
		}

		if buf[0] == 0 {
			// Encontramos un byte que es 0
			Auxlogger.Println("Bloque libre: ", byteCount)
			FS_totalBloquesDisponibles--
			Auxlogger.Println("Cantidad de Blq. Libres: ", FS_totalBloquesDisponibles)

			// Escribimos un 1 en el byte que encontramos
			_, err = file.WriteAt([]byte{1}, int64(byteCount))
			if err != nil {
				logueano.Error(Auxlogger, err)
				return -1
			}

			return byteCount
		}
		byteCount++
	}
	// No encontramos ningún byte que sea 0
	return -1
}

// ------ IO_FS_TRUNCATE ------

// ----- Agrandar Archivo
func agrandarArchivo(nuevoTamañoEnBloques int, metadata *structs.MetadataFS, tamañoEnBloques int, nombreArchivo string) {

	//Verifico si hay espacio suficiente en el disco
	if nuevoTamañoEnBloques > FS_totalBloquesDisponibles {
		logueano.Mensaje(Auxlogger, "No hay espacio suficiente en disco")
		//Este caso no se testea, ni se realiza ninguna operacion en especifico
		return
	}

	//Verifico si existe tamaño suficiente contiguo al bloque
	espacioEsContiguo, err := espacioContiguo(nuevoTamañoEnBloques, *metadata)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}

	//En caso de que haya espacio contiguo se reservan los bloques necesarios
	if espacioEsContiguo {
		reservarBloques(nuevoTamañoEnBloques, tamañoEnBloques, *metadata)
		return
	}

	//En caso de que no haya espacio suficiente contiguo al bloque, se reorganizan los bloques (compactación y recolocacion del archivo)
	metadata.InitialBlock = reorganizarBloques(metadata.InitialBlock, tamañoEnBloques, nuevoTamañoEnBloques, nombreArchivo)
}

func espacioContiguo(nuevoTamañoEnBloques int, metadata structs.MetadataFS) (bool, error) {

	bitmap, err := os.Open(configInterfaz.Dialfs_Path + "/" + "bitmap.dat")
	if err != nil {
		logueano.Error(Auxlogger, err)
		return false, err
	}
	defer bitmap.Close()

	// Primer bloque a recorrer para ver si hay espacio contiguo
	primerBloqueARecorrer := metadata.InitialBlock + calcularTamañoEnBloques(metadata.Size)

	buf := make([]byte, 1)

	// Recorre n bloques a partir del primer bloque a recorrer, siendo n la nueva cantidad de bloques
	for i := primerBloqueARecorrer; i < metadata.InitialBlock+nuevoTamañoEnBloques; i++ {
		_, err := bitmap.ReadAt(buf, int64(i))
		if err != nil {
			logueano.Error(Auxlogger, err)
			return false, err
		}
		if buf[0] == 1 {
			return false, nil
		}
	}

	return true, nil
}

func reservarBloques(nuevoTamañoEnBloques int, tamañoEnBloques int, metadata structs.MetadataFS) {

	// Abro bitmap
	file, err := os.OpenFile(configInterfaz.Dialfs_Path+"/"+"bitmap.dat", os.O_RDWR, 0644)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}
	defer file.Close()

	//Asigno (1) los bloques nuevos
	bloquesAgregados := nuevoTamañoEnBloques - tamañoEnBloques

	pos := metadata.InitialBlock + tamañoEnBloques

	for i := bloquesAgregados; i > 0; i-- {
		_, err = file.WriteAt([]byte{1}, int64(pos))
		if err != nil {
			logueano.Error(Auxlogger, err)
			return
		}
		FS_totalBloquesDisponibles--
		pos++
	}
}

func reorganizarBloques(initialBlock int, tamañoEnBloques int, nuevoTamañoEnBloques int, nombreArchivo string) int {

	//Abro el archivo .dat
	fDataBloques, err := os.OpenFile(configInterfaz.Dialfs_Path+"/"+"bloques.dat", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return -1
	}

	//Guardo data del archivo a agrandar en un buffer
	bufferTruncate := make([]byte, tamañoEnBloques*configInterfaz.Dialfs_Block_Size)
	_, err = fDataBloques.ReadAt(bufferTruncate, int64(initialBlock*configInterfaz.Dialfs_Block_Size))
	if err != nil {
		logueano.Error(Auxlogger, err)
		return -1
	}

	//Compacto los archivos en dico (dejo los bloques libres al final del disco)
	nuevaPosInicial := compactar(fDataBloques, nombreArchivo, bufferTruncate)

	fDataBloques.Close()

	//Actualizo bitmap
	bitmap, err := os.OpenFile(configInterfaz.Dialfs_Path+"/"+"bitmap.dat", os.O_RDWR, 0644)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return -1
	}

	//Pongo 1 desdel el inicio del bitmap hasta: nuevaPosInicial + nuevoTamañoEnBloques y ceros en el resto
	for i := 0; i < nuevaPosInicial+nuevoTamañoEnBloques; i++ {
		_, err = bitmap.WriteAt([]byte{1}, int64(i))
		if err != nil {
			logueano.Error(Auxlogger, err)
			return -1
		}
	}

	for i := nuevaPosInicial + nuevoTamañoEnBloques; i < configInterfaz.Dialfs_Block_Count; i++ {
		_, err = bitmap.WriteAt([]byte{0}, int64(i))
		if err != nil {
			logueano.Error(Auxlogger, err)
			return -1
		}
	}

	bitmap.Close()

	FS_totalBloquesDisponibles = configInterfaz.Dialfs_Block_Count - (nuevaPosInicial + nuevoTamañoEnBloques)

	return nuevaPosInicial
}

func compactar(fDataBloques *os.File, nombreArchivo string, bufferTruncate []byte) int {

	//Crear un nuevo archivo temporal
	fTemp, err := os.OpenFile(configInterfaz.Dialfs_Path+"/"+"bloques.dat.tmp", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return -1
	}

	//Tomo todos los archivos y los escribo contiguos en el archivo temporal
	var punteroUltimoBloqueLibre int
	for i := 0; i < len(listaArchivos); i++ {

		//Tomo el archivo i-esimo y lo escribo en el data temporal
		archivo := listaArchivos[i]
		if archivo == nombreArchivo {
			continue
		}

		//Abrir la metadata del archivo
		metadata, err := extraerMetadata(archivo)
		if err != nil {
			logueano.Error(Auxlogger, err)
			return -1
		}

		sizeEnBloques := calcularTamañoEnBloques(metadata.Size)
		tempBuffer := make([]byte, sizeEnBloques*configInterfaz.Dialfs_Block_Size)

		//Leer, de acuerdo a la metadata, los bloques de fDataBloques
		_, err = fDataBloques.ReadAt(tempBuffer, int64(metadata.InitialBlock*configInterfaz.Dialfs_Block_Size))
		if err != nil {
			logueano.Error(Auxlogger, err)
			return -1
		}

		//Actualizar la metadata en base al nuevo bloque inicial
		metadata.InitialBlock = punteroUltimoBloqueLibre
		actualizarMetadata(archivo, metadata)

		punteroUltimoBloqueLibre += sizeEnBloques

		//Escribir en el archivo temporal el tmpBuffer
		_, err = fTemp.Write(tempBuffer)
		if err != nil {
			logueano.Error(Auxlogger, err)
			return -1
		}
	}

	//Escribir en el archivo temporal el bufferTruncate (archivo que se quiere agrandar)
	_, err = fTemp.Write(bufferTruncate)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return -1
	}

	// Establece el tamaño del archivo
	err = fTemp.Truncate(int64(configInterfaz.Dialfs_Block_Size * configInterfaz.Dialfs_Block_Count))
	if err != nil {
		log.Fatal(err)
	}

	fTemp.Close()

	//renombro el archivo temporal con el nombre del archivo original
	err = os.Rename(configInterfaz.Dialfs_Path+"/"+"bloques.dat.tmp", configInterfaz.Dialfs_Path+"/"+"bloques.dat")
	if err != nil {
		logueano.Error(Auxlogger, err)
	}

	return punteroUltimoBloqueLibre
}

// ----- Achicar Archivo
func achicarArchivo(nuevoTamañoEnBloques int, tamañoEnBloques int, metadata structs.MetadataFS) {
	liberarBloques(tamañoEnBloques, nuevoTamañoEnBloques, metadata.InitialBlock)
}

func liberarBloques(tamañoEnBloques int, nuevoTamañoEnBloques int, bloqueInicial int) {

	// Abro bitmap
	file, err := os.OpenFile(configInterfaz.Dialfs_Path+"/"+"bitmap.dat", os.O_RDWR, 0644)
	if err != nil {
		logueano.Error(Auxlogger, err)
		return
	}
	defer file.Close()

	//Libero (0) los bloques extra
	bloquesABorrar := tamañoEnBloques - nuevoTamañoEnBloques

	pos := bloqueInicial + nuevoTamañoEnBloques

	for i := bloquesABorrar; i > 0; i-- {
		_, err = file.WriteAt([]byte{0}, int64(pos))
		if err != nil {
			logueano.Error(Auxlogger, err)
			return
		}
		FS_totalBloquesDisponibles++
		pos++
	}
}
