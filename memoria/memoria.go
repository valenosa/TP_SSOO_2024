package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sisoputnfrba/tp-golang/memoria/funciones"
	"github.com/sisoputnfrba/tp-golang/utils/config"
	"github.com/sisoputnfrba/tp-golang/utils/structs"
)

//================================| MAIN |================================\\

//================================| MAIN |================================\\

func main() {

	// Extrae info de config.json
	config.Iniciar("config.json", &funciones.ConfigJson)

	// Crea e inicializa la memoria de instrucciones
	memoriaInstrucciones := make(map[uint32][]string) // Contiene las instrucciones de cada proceso (pid = key). Almacena las instrucciones por separado en un slice de strings.

	espacioUsuario := make([]byte, funciones.ConfigJson.Memory_Size) // Contiene todo lo que se guarda para cada proceso (a excepcion)

	tablasDePaginas := make(map[uint32]structs.Tabla) // Contiene la tabla de cada proceso (pid = key)

	bitMap := make([]bool, funciones.ConfigJson.Memory_Size/funciones.ConfigJson.Page_Size) // TRUE = ocupado, FALSE = libre

	// Variables que no se usan pero se dejan para que no tire error el compilador
	_ = bitMap
	_ = tablasDePaginas
	_ = espacioUsuario

	// Configura el logger
	config.Logger("Memoria.log")

	// ======== HandleFunctions ========
	http.HandleFunc("PUT /process", handlerMemIniciarProceso(memoriaInstrucciones, tablasDePaginas, bitMap))
	http.HandleFunc("GET /instrucciones", handlerEnviarInstruccion(memoriaInstrucciones))
	http.HandleFunc("DELETE /process", handlerFinalizarProcesoMemoria(memoriaInstrucciones, tablasDePaginas, bitMap))
	http.HandleFunc("PUT /memoria/resize", handlerResize(&tablasDePaginas, bitMap))

	//inicio el servidor de Memoria
	config.IniciarServidor(funciones.ConfigJson.Port)
}

//================================| HANDLERS |================================\\

// Wrapper que crea un PCB con el pid recibido.
func handlerMemIniciarProceso(memoriaInstrucciones map[uint32][]string, tablaDePaginas map[uint32]structs.Tabla, bitMap []bool) func(http.ResponseWriter, *http.Request) {

	// Handler para iniciar un proceso.
	return func(w http.ResponseWriter, r *http.Request) {

		//variable que recibirá la request.
		var request structs.BodyIniciarProceso

		// Decodifica en formato JSON la request.
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			fmt.Println(err) //TODO: por el momento se deja para desarrollo, eliminar al terminar el TP.
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Se guardan las instrucciones en un map de memoria.
		funciones.GuardarInstrucciones(request.PID, request.Path, memoriaInstrucciones)

		funciones.AsignarTabla(request.PID, tablaDePaginas)

		// Crea una variable tipo Response (para confeccionar una respuesta)
		var respBody structs.ResponseListarProceso = structs.ResponseListarProceso{PID: request.PID}
		respuesta, err := json.Marshal(respBody)
		if err != nil {
			fmt.Println(err) //TODO: por el momento se deja para desarrollo, eliminar al terminar el TP.
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
		queryParams := r.URL.Query()
		pid, errPid := strconv.ParseUint(queryParams.Get("PID"), 10, 32)
		pc, errPC := strconv.ParseUint(queryParams.Get("PC"), 10, 32)

		if errPid != nil || errPC != nil {
			return
		}

		instruccion := memoriaInstrucciones[uint32(pid)][uint32(pc)]
		fmt.Println(instruccion) //! Borrar despues

		// Esperar un tiempo determinado a tiempo de retardo
		time.Sleep(time.Duration(funciones.ConfigJson.Delay_Response) * time.Millisecond)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(instruccion))
	}
}

// TODO: Crear request que venga a este handler, el endpoint, y probar
func handlerFinalizarProcesoMemoria(memoriaInstrucciones map[uint32][]string, tablaDePaginas map[uint32]structs.Tabla, bitMap []bool) func(http.ResponseWriter, *http.Request) {

	// Recibe el pid y borra las estructuras relacionadas al mismo (instrucciones, tabla de páginas, libera bitmap)
	return func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		pid, errPid := strconv.ParseUint(queryParams.Get("PID"), 10, 32)

		if errPid != nil {
			return
		}

		// Borrar instrucciones
		delete(memoriaInstrucciones, uint32(pid))
		// Desocupar marcos
		funciones.LiberarMarcos(tablaDePaginas[uint32(pid)], bitMap)
		// Borrar tabla de páginas
		delete(tablaDePaginas, uint32(pid)) //?Alcanza o hace falta mandarle un puntero?

		w.WriteHeader(http.StatusOK) //?
		// w.Write([]byte)
	}
}

// TODO: Probar
func handlerResize(tablaDePaginas *map[uint32]structs.Tabla, bitMap []bool) func(http.ResponseWriter, *http.Request) {

	// Recibe el pid y borra las estructuras relacionadas al mismo (instrucciones, tabla de páginas, libera bitmap)
	return func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		pid, errPid := strconv.ParseUint(queryParams.Get("pid"), 10, 32)
		size, errSize := strconv.ParseUint(queryParams.Get("size"), 10, 32)

		if errPid != nil || errSize != nil {
			return
		}

		estado := funciones.ReasignarPaginas(uint32(pid), tablaDePaginas, bitMap, uint32(size))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(estado)) //?Está bien mandar texto así? Es esto lo que hay que mandar?
	}
}
