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

	var cantBloquesDisponiblesTotal int = configInterfaz.Dialfs_Block_Count //?Esto funciona solamente si una vez que se termina la ejecución de la interfaz, se reinicia el sistema. Si hay permanencia, entonces cada vez que levantemos la interfaz deberíamos leer el bitmap.dat y contar la cantidad de bloques libres.

	//----------( INICIAMOS INTERFAZ )----------

	// Envio a Kernel la nueva interfaz
	conectarInterfazIO(nombreInterfaz)

	// Levanta el server de la nuevaInterfazIO
	serverErr := iniciarServidorInterfaz(&cantBloquesDisponiblesTotal)
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

func iniciarServidorInterfaz(cantBloquesDisponiblesTotal *int) error {

	http.HandleFunc("POST /GENERICA/IO_GEN_SLEEP", handlerIO_GEN_SLEEP)
	http.HandleFunc("POST /STDIN/IO_STDIN_READ", handlerIO_STDIN_READ)
	http.HandleFunc("POST /STDOUT/IO_STDOUT_WRITE", handlerIO_STDOUT_WRITE)

	http.HandleFunc("POST /DIALFS/IO_FS_CREATE", handlerIO_FS_CREATE(cantBloquesDisponiblesTotal)) //!Modificar la request desde kernel para que no ponga /TipoDeInstruccion
	http.HandleFunc("POST /DIALFS/handlerIO_FS_TRUNCATE", handlerIO_FS_TRUNCATE(cantBloquesDisponiblesTotal))

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

//*======================( STDIN )======================

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

//*======================( STDOUT )======================

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

//*======================( DIALFS )======================

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

func handlerIO_FS_CREATE(cantBloquesDisponiblesTotal *int) func(http.ResponseWriter, *http.Request) {

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

		//-------- EJECUTA ---------

		metadata := structs.MetadataFS{}

		//Asigna su primer bloque
		metadata.InitialBlock = asignarEspacio(cantBloquesDisponiblesTotal)

		metadata.Size = int(instruccionIO.Tamaño)

		actualizarMetadata(instruccionIO.NombreArchivo, metadata)

		//agrego el nombre del archivo a la listaArchivo
		listaArchivos = append(listaArchivos, instruccionIO.NombreArchivo)

		//--------- RESPUESTA ---------
		// Envía el status al Kernel
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(":/"))
	}
}

// TODO: ACTUALIZAR EL CANTIDAD DE BLOQUES DISPONIBLES EN EL BITMAP.DAT
func handlerIO_FS_TRUNCATE(cantBloquesDisponiblesTotal *int) func(http.ResponseWriter, *http.Request) {

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

		//-------- EJECUTA ---------

		metadata, err := extraerMetadata(instruccionIO.NombreArchivo)
		if err != nil {
			fmt.Println(err)
			return
		}

		tamañoEnBloques := int(math.Ceil(float64(metadata.Size) / float64(configInterfaz.Dialfs_Block_Size)))

		nuevoTamañoEnBloques := int(math.Ceil(float64(instruccionIO.Tamaño) / float64(configInterfaz.Dialfs_Block_Size)))

		if nuevoTamañoEnBloques > tamañoEnBloques {
			agrandarArchivo(nuevoTamañoEnBloques, &metadata, tamañoEnBloques, cantBloquesDisponiblesTotal, instruccionIO.NombreArchivo)
		}
		if nuevoTamañoEnBloques < tamañoEnBloques {
			achicarArchivo(nuevoTamañoEnBloques, tamañoEnBloques, metadata, cantBloquesDisponiblesTotal)
		}

		metadata.Size = int(instruccionIO.Tamaño)
		actualizarMetadata(instruccionIO.NombreArchivo, metadata)

		//--------- RESPUESTA ---------
		// Envía el status al Kernel
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(":/"))
	}
}

//--------------- FUNCIONES DIALFS ---------------

// --------------- METADATA

func extraerMetadata(nombreArchivo string) (structs.MetadataFS, error) {
	file, err := os.Open(configInterfaz.Dialfs_Path + "/" + nombreArchivo)
	if err != nil {
		fmt.Println(err)
		return structs.MetadataFS{}, err
	}

	defer file.Close()
	// Lee todo el contenido del archivo
	bytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return structs.MetadataFS{}, err
	}

	// Deserializa los bytes a la estructura Metadata
	var metadata structs.MetadataFS
	err = json.Unmarshal(bytes, &metadata)
	if err != nil {
		fmt.Println(err)
		return structs.MetadataFS{}, err
	}
	return metadata, nil
}

func actualizarMetadata(nombreArchivo string, nuevaMetadata structs.MetadataFS) {

	nuevoArchivo, err := os.OpenFile(configInterfaz.Dialfs_Path+"/"+nombreArchivo, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer nuevoArchivo.Close()

	//Se escribe en el archivo
	encoder := json.NewEncoder(nuevoArchivo)
	err = encoder.Encode(nuevaMetadata)
	if err != nil {
		fmt.Println(err)
		return
	}
}

//--------------- IO_FS_CREATE

func asignarEspacio(cantBloquesDisponiblesTotal *int) int {

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
			*cantBloquesDisponiblesTotal--
			Auxlogger.Println("Cantidad de Blq. Libres: ", *cantBloquesDisponiblesTotal) //!LOG

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

// --------------- IO_FS_TRUNCATE

// ----- Agrandar Archivo
func agrandarArchivo(nuevoTamañoEnBloques int, metadata *structs.MetadataFS, tamañoEnBloques int, cantBloquesDisponiblesTotal *int, nombreArchivo string) {

	//Verifico si hay espacio suficiente en el disco
	if nuevoTamañoEnBloques < *cantBloquesDisponiblesTotal {
		fmt.Println("No hay suficiente espacio en el disco")
		//Este caso no se testea, ni se realiza ninguna operacion en especifico
		return
	}

	//Verifico si existe tamaño suficiente contiguo al bloque
	espacioEsContiguo, err := espacioContiguo(tamañoEnBloques, *metadata)
	if err != nil {
		fmt.Println(err)
		return
	}

	//En caso de que haya espacio contiguo se reservan los bloques necesarios
	if espacioEsContiguo {
		reservarBloques(nuevoTamañoEnBloques, tamañoEnBloques, *metadata, cantBloquesDisponiblesTotal)
		return
	}

	//En caso de que no haya espacio suficiente contiguo al bloque, se reorganizan los bloques (compactación y recolocacion del archivo)
	metadata.InitialBlock = reorganizarBloques(metadata.InitialBlock, tamañoEnBloques, nuevoTamañoEnBloques, cantBloquesDisponiblesTotal, nombreArchivo)

}

func espacioContiguo(tamaño int, metadata structs.MetadataFS) (bool, error) {

	bitmap, err := os.Open(configInterfaz.Dialfs_Path + "/" + "bitmap.dat")
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	defer bitmap.Close()

	// Primer bloque a recorrer para ver si hay espacio contiguo
	primerBloqueARecorrer := metadata.InitialBlock + int(math.Ceil(float64(metadata.Size)/float64(configInterfaz.Dialfs_Block_Size)))

	buf := make([]byte, 1)

	// Recorre n bloques a partir del primer bloque a recorrer, siendo n la variable "tamaño"
	for i := primerBloqueARecorrer; i < primerBloqueARecorrer+tamaño; i++ {
		_, err := bitmap.ReadAt(buf, int64(i))
		if err != nil {
			fmt.Println(err)
			return false, err
		}
		if buf[0] == 1 {
			return false, nil
		}
	}

	return true, nil
}

func reservarBloques(nuevoTamañoEnBloques int, tamañoEnBloques int, metadata structs.MetadataFS, cantBloquesDisponiblesTotal *int) {

	// Abro bitmap
	file, err := os.OpenFile(configInterfaz.Dialfs_Path+"/"+"bitmap.dat", os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	//Asigno (1) los bloques nuevos
	bloquesAgregados := nuevoTamañoEnBloques - tamañoEnBloques

	pos := metadata.InitialBlock + tamañoEnBloques + 1

	for i := bloquesAgregados; i > 0; i-- {
		_, err = file.WriteAt([]byte{1}, int64(pos))
		if err != nil {
			fmt.Println(err)
			return
		}
		*cantBloquesDisponiblesTotal--
		pos++
	}
}

func reorganizarBloques(initialBlock int, tamañoEnBloques int, nuevoTamañoEnBloques int, cantBloquesDisponiblesTotal *int, nombreArchivo string) int {

	//Abro el archivo .dat
	fDataBloques, err := os.OpenFile(configInterfaz.Dialfs_Path+"/"+"bloques.dat", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	//Guardo data del archivo a agrandar en un buffer
	bufferTruncate := make([]byte, tamañoEnBloques*configInterfaz.Dialfs_Block_Size)
	_, err = fDataBloques.ReadAt(bufferTruncate, int64(initialBlock*configInterfaz.Dialfs_Block_Size))
	if err != nil {
		fmt.Println(err)
		return -1
	}

	//Compacto los archivos en dico (dejo los bloques libres al final del disco)
	nuevaPosInicial := compactar(fDataBloques, nombreArchivo, bufferTruncate)

	fDataBloques.Close()

	//Actualizo bitmap
	bitmap, err := os.OpenFile(configInterfaz.Dialfs_Path+"/"+"bitmap.dat", os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	//Pongo 1 desdel el inicio del bitmap hasta: nuevaPosInicial + nuevoTamañoEnBloques y ceros en el resto
	for i := 0; i < nuevaPosInicial+nuevoTamañoEnBloques; i++ {
		_, err = bitmap.WriteAt([]byte{1}, int64(i))
		if err != nil {
			fmt.Println(err)
			return -1
		}
	}

	for i := nuevaPosInicial + nuevoTamañoEnBloques; i < configInterfaz.Dialfs_Block_Count; i++ {
		_, err = bitmap.WriteAt([]byte{0}, int64(i))
		if err != nil {
			fmt.Println(err)
			return -1
		}
	}

	bitmap.Close()

	//? Verificar este calculo
	*cantBloquesDisponiblesTotal = configInterfaz.Dialfs_Block_Count - (nuevaPosInicial + nuevoTamañoEnBloques)

	return nuevaPosInicial
}

func compactar(fDataBloques *os.File, nombreArchivo string, bufferTruncate []byte) int {

	//Crear un nuevo archivo temporal
	fTemp, err := os.OpenFile(configInterfaz.Dialfs_Path+"/"+"bloques.dat.tmp", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println(err)
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
			fmt.Println(err)
			return -1
		}

		sizeEnBloques := int(math.Ceil(float64(metadata.Size) / float64(configInterfaz.Dialfs_Block_Size)))
		tempBuffer := make([]byte, sizeEnBloques*configInterfaz.Dialfs_Block_Size)

		//Leer, de acuerdo a la metadata, los bloques de fDataBloques
		_, err = fDataBloques.ReadAt(tempBuffer, int64(metadata.InitialBlock*configInterfaz.Dialfs_Block_Size))
		if err != nil {
			fmt.Println(err)
			return -1
		}

		punteroUltimoBloqueLibre += sizeEnBloques

		//Actualizar la metadata en base al nuevo bloque inicial
		metadata.InitialBlock = punteroUltimoBloqueLibre
		actualizarMetadata(archivo, metadata)

		//Escribir en el archivo temporal el tmpBuffer
		_, err = fTemp.Write(tempBuffer)
		if err != nil {
			fmt.Println(err)
			return -1
		}
	}

	//Escribir en el archivo temporal el bufferTruncate (archivo que se quiere agrandar)
	_, err = fTemp.Write(bufferTruncate)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	fTemp.Close()

	//renombro el archivo temporal con el nombre del archivo original
	err = os.Rename(configInterfaz.Dialfs_Path+"/"+"bloques.dat.tmp", configInterfaz.Dialfs_Path+"/"+"bloques.dat")
	if err != nil {
		fmt.Println(err)
	}

	return punteroUltimoBloqueLibre
}

//----- Achicar Archivo

func achicarArchivo(nuevoTamañoEnBloques int, tamañoEnBloques int, metadata structs.MetadataFS, cantBloquesDisponiblesTotal *int) {
	liberarBloques(tamañoEnBloques, nuevoTamañoEnBloques, metadata.InitialBlock, cantBloquesDisponiblesTotal)
}

func liberarBloques(tamañoEnBloques int, nuevoTamañoEnBloques int, bloqueInicial int, cantBloquesDisponiblesTotal *int) {

	// Abro bitmap
	file, err := os.OpenFile(configInterfaz.Dialfs_Path+"/"+"bitmap.dat", os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	//Libero (0) los bloques extra
	bloquesABorrar := tamañoEnBloques - nuevoTamañoEnBloques

	pos := bloqueInicial + nuevoTamañoEnBloques

	for i := bloquesABorrar; i > 0; i-- {
		_, err = file.WriteAt([]byte{0}, int64(pos))
		if err != nil {
			fmt.Println(err)
			return
		}
		*cantBloquesDisponiblesTotal++
		pos++
	}
}
