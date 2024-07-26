package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sisoputnfrba/tp-golang/memoria/funciones"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/logueano"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

//================================| MAIN |================================\\

func main() {

	// Extrae info de config.json
	configPath := os.Args[1]
	config.Iniciar(configPath, &funciones.ConfigJson)

	// Crea e inicializa la memoria de instrucciones
	memoriaInstrucciones := make(map[uint32][]string) // Contiene las instrucciones de cada proceso (pid = key). Almacena las instrucciones por separado en un slice de strings.

	espacioUsuario := make([]byte, funciones.ConfigJson.Memory_Size) // Contiene todo lo que se guarda para cada proceso (a excepcion)

	tablasDePaginas := make(map[uint32]structs.Tabla) // Contiene la tabla de cada proceso (pid = key)

	bitMap := make([]bool, funciones.ConfigJson.Memory_Size/funciones.ConfigJson.Page_Size) // TRUE = ocupado, FALSE = libre

	// Configura el logger (aux en funciones.go)
	logueano.Logger("memoria.log")

	funciones.Auxlogger = logueano.InitAuxLog("memoria")

	// ======== HandleFunctions ========
	http.HandleFunc("PUT /process", handlerMemIniciarProceso(memoriaInstrucciones, tablasDePaginas))
	http.HandleFunc("GET /instrucciones", handlerEnviarInstruccion(memoriaInstrucciones))
	http.HandleFunc("DELETE /process", handlerFinalizarProcesoMemoria(memoriaInstrucciones, tablasDePaginas, bitMap))
	http.HandleFunc("PUT /memoria/resize", handlerResize(&tablasDePaginas, bitMap))
	http.HandleFunc("GET /memoria/marco", handlerObtenerMarco(tablasDePaginas))
	http.HandleFunc("GET /memoria/movin", handlerMovIn(&espacioUsuario, tablasDePaginas))
	http.HandleFunc("POST /memoria/movout", handlerMovOut(&espacioUsuario, tablasDePaginas))
	http.HandleFunc("GET /memoria/copystr", handlerCopyString(&espacioUsuario, tablasDePaginas))

	// Inicio el servidor de Memoria
	config.IniciarServidor(funciones.ConfigJson.Port)
}

//================================| HANDLERS |================================\\

// Wrapper que crea un PCB con el pid recibido.
func handlerMemIniciarProceso(memoriaInstrucciones map[uint32][]string, tablaDePaginas map[uint32]structs.Tabla) func(http.ResponseWriter, *http.Request) {

	// Handler para iniciar un proceso.
	return func(w http.ResponseWriter, r *http.Request) {

		//--------- REQUEST ---------

		//variable que recibirá la request.
		var request structs.IniciarProceso

		// Decodifica en formato JSON la request.
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			logueano.Error(funciones.Auxlogger, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//--------- EJECUTA ---------

		// Se guardan las instrucciones en un map de memoria.
		funciones.GuardarInstrucciones(request.PID, request.Path, memoriaInstrucciones)

		funciones.AsignarTabla(request.PID, tablaDePaginas)

		//--------- RESPUESTA ---------

		// Crea una variable tipo Response (para confeccionar una respuesta)
		var respBody structs.ResponseListarProceso = structs.ResponseListarProceso{PID: request.PID}
		respuesta, err := json.Marshal(respBody)
		if err != nil {
			logueano.Error(funciones.Auxlogger, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Envía respuesta (con estatus como header) al cliente
		w.WriteHeader(http.StatusOK)
		w.Write(respuesta)
	}
}

// Envía a CPU la instrucción correspondiente al pid y el pc del map de memoria
func handlerEnviarInstruccion(memoriaInstrucciones map[uint32][]string) func(http.ResponseWriter, *http.Request) {

	// Handler para enviar una instruccion
	return func(w http.ResponseWriter, r *http.Request) {

		//--------- REQUEST ---------

		queryParams := r.URL.Query()
		pid, errPid := strconv.ParseUint(queryParams.Get("PID"), 10, 32)
		pc, errPC := strconv.ParseUint(queryParams.Get("PC"), 10, 32)

		if errPid != nil || errPC != nil {
			return
		}

		//--------- EJECUTA ---------

		listaInst, existe := memoriaInstrucciones[uint32(pid)]
		if !existe || uint32(pc) >= uint32(len(listaInst)) {

			// Manejar el error aquí
			logueano.MensajeConFormato(funciones.Auxlogger, "¡ERROR! PID: %d - PC: %d - LEN: %d", pid, pc, len(listaInst))
		}

		instruccion := listaInst[uint32(pc)]

		fetch := structs.Fetch{Page_Size: funciones.ConfigJson.Page_Size, Instruccion: instruccion}

		respuesta, err := json.Marshal(fetch)
		if err != nil {
			logueano.Error(funciones.Auxlogger, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Esperar un tiempo determinado a tiempo de retardo
		time.Sleep(time.Duration(funciones.ConfigJson.Delay_Response) * time.Millisecond)

		//--------- RESPUESTA ---------

		w.WriteHeader(http.StatusOK)
		w.Write(respuesta)
	}
}

func handlerFinalizarProcesoMemoria(memoriaInstrucciones map[uint32][]string, tablaDePaginas map[uint32]structs.Tabla, bitMap []bool) func(http.ResponseWriter, *http.Request) {

	// Recibe el pid y borra las estructuras relacionadas al mismo (instrucciones, tabla de páginas, libera bitmap)
	return func(w http.ResponseWriter, r *http.Request) {

		//--------- REQUEST ---------

		queryParams := r.URL.Query()
		pid, errPid := strconv.ParseUint(queryParams.Get("pid"), 10, 32)
		if errPid != nil {
			return
		}

		//--------- EJECUTA ---------

		// Borrar instrucciones
		delete(memoriaInstrucciones, uint32(pid))
		// Desocupar marcos
		funciones.LiberarMarcos(tablaDePaginas[uint32(pid)], bitMap)

		//^log obligatorio (1/6)
		logueano.OperoConTablaDePaginas(uint32(pid), tablaDePaginas)

		delete(tablaDePaginas, uint32(pid))

		//--------- RESPUESTA ---------

		w.WriteHeader(http.StatusOK)
	}
}

func handlerObtenerMarco(tablaDePaginas map[uint32]structs.Tabla) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		//--------- REQUEST ---------

		// Desglosa los Query Params
		queryParams := r.URL.Query()
		pid, errPid := strconv.ParseUint(queryParams.Get("pid"), 10, 32)
		pagina, errPagina := strconv.ParseUint(queryParams.Get("pagina"), 10, 32)

		// Maneja error en caso de que no se pueda parsear el query
		if errPid != nil || errPagina != nil {
			http.Error(w, "Error al parsear las query params", http.StatusInternalServerError)
			return
		}

		//--------- EJECUTA ---------

		// Busca marco en la tabla de páginas, y en caso de no encontrarlo, devuelve un string vacío
		marco := funciones.BuscarMarco(uint32(pid), uint32(pagina), tablaDePaginas)

		//--------- RESPUESTA ---------

		// Devuelve un status code dependiendo de si se encontró o no el marco
		if marco == "" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		// Envía la respuesta al MMU
		w.Write([]byte(marco))
	}
}

func handlerMovIn(espacioUsuario *[]byte, tablaDePaginas map[uint32]structs.Tabla) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		//--------- REQUEST ---------

		//Obtengo los query params
		queryParams := r.URL.Query()
		pid, errPid := strconv.ParseUint(queryParams.Get("pid"), 10, 32)
		direccionFisica, errDF := strconv.ParseUint(queryParams.Get("dir"), 10, 32)
		byteArraySize, errReg := strconv.Atoi(queryParams.Get("size"))

		// Maneja error en caso de que no se pueda parsear el query
		if errDF != nil || errReg != nil || errPid != nil {
			http.Error(w, "Error al parsear las query params", http.StatusInternalServerError)
			return
		}

		//--------- EJECUTA ---------

		pagina := funciones.ObtenerPagina(uint32(pid), uint32(direccionFisica), tablaDePaginas)

		data, estado := funciones.LeerEnMemoria(uint32(pid), tablaDePaginas, uint32(pagina), uint32(direccionFisica), byteArraySize, espacioUsuario)

		//--------- RESPUESTA ---------

		if estado != "OK" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		w.Write(data)
	}
}

func handlerMovOut(espacioUsuario *[]byte, tablaDePaginas map[uint32]structs.Tabla) func(http.ResponseWriter, *http.Request) { //? Es necesario este parámetro?
	return func(w http.ResponseWriter, r *http.Request) {

		//--------- REQUEST ---------

		// Variable que recibirá la request.
		var request structs.RequestMovOUT

		// Decodifica en formato JSON la request.
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			logueano.Error(funciones.Auxlogger, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//--------- EJECUTA ---------

		pagina := funciones.ObtenerPagina(request.Pid, request.Dir, tablaDePaginas)

		if pagina == -1 {
			logueano.Mensaje(funciones.Auxlogger, "Error: no se encontró la página")
		}

		estado := funciones.EscribirEnMemoria(request.Pid, tablaDePaginas, uint32(pagina), request.Dir, request.Data, espacioUsuario)
		// Devuelve un status code dependiendo de si se encontró o no el marco
		if estado == "OK" {
			w.WriteHeader(http.StatusOK)
		}
		if estado == "OUT_OF_MEMORY" {
			w.WriteHeader(http.StatusNotFound)
		}

		//--------- RESPUESTA ---------
		w.Write([]byte(estado))
	}
}

func handlerResize(tablaDePaginas *map[uint32]structs.Tabla, bitMap []bool) func(http.ResponseWriter, *http.Request) {

	// Recibe el pid y borra las estructuras relacionadas al mismo (instrucciones, tabla de páginas, libera bitmap)
	return func(w http.ResponseWriter, r *http.Request) {

		//--------- REQUEST ---------

		queryParams := r.URL.Query()
		pid, errPid := strconv.ParseUint(queryParams.Get("pid"), 10, 32)
		size, errSize := strconv.ParseUint(queryParams.Get("size"), 10, 32)

		if errPid != nil || errSize != nil {
			return
		}
		//--------- EJECUTA ---------

		estado := funciones.ReasignarPaginas(uint32(pid), tablaDePaginas, bitMap, uint32(size))

		//--------- RESPUESTA ---------

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(estado))
	}
}

func handlerCopyString(espacioUsuario *[]byte, tablaDePaginas map[uint32]structs.Tabla) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		//--------- REQUEST ---------
		//Obtengo los query params
		queryParams := r.URL.Query()
		pid, errPid := strconv.ParseUint(queryParams.Get("pid"), 10, 32)
		direccionEscritura, errDE := strconv.ParseUint(queryParams.Get("write"), 10, 32)
		direccionLectura, errDL := strconv.ParseUint(queryParams.Get("read"), 10, 32)
		byteArraySize, errByte := strconv.Atoi(queryParams.Get("size"))

		// Maneja error en caso de que no se pueda parsear el query
		if errDE != nil || errDL != nil || errByte != nil || errPid != nil {
			http.Error(w, "Error al parsear las query params", http.StatusInternalServerError)
			return
		}

		//--------- EJECUTA ---------

		paginaEscritura := funciones.ObtenerPagina(uint32(pid), uint32(direccionEscritura), tablaDePaginas)

		paginaLectura := funciones.ObtenerPagina(uint32(pid), uint32(direccionLectura), tablaDePaginas)

		if paginaEscritura == -1 || paginaLectura == -1 {
			logueano.Mensaje(funciones.Auxlogger, "ERROR: No se encontró la página")
		}

		//--------- LECTURA ---------

		data, estadoLectura := funciones.LeerEnMemoria(uint32(pid), tablaDePaginas, uint32(paginaLectura), uint32(direccionLectura), byteArraySize, espacioUsuario)

		//--------- ESCRITURA ---------

		estadoEscritura := funciones.EscribirEnMemoria(uint32(pid), tablaDePaginas, uint32(paginaEscritura), uint32(direccionEscritura), data, espacioUsuario)

		//--------- RESPUESTA ---------

		// Devuelve un status code dependiendo de si se encontró o no el marco
		if estadoLectura == "OK" && estadoEscritura == "OK" {
			w.WriteHeader(http.StatusOK)
		} else if estadoLectura == "OUT_OF_MEMORY" || estadoEscritura == "OUT_OF_MEMORY" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Write(data)
	}
}
