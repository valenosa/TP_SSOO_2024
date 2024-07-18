package structs

import "sync"

//*=====================================| PCB |========================================================\\

// Estructura base de un proceso.
type PCB struct {
	PID      uint32
	Quantum  uint16
	Estado   string
	Recursos []string
	RegistrosUsoGeneral
}

// Estructura de los registros de uso general (para tener info del contexto de ejecución de cada PCB)
type RegistrosUsoGeneral struct {
	PC  uint32
	AX  uint8
	BX  uint8
	CX  uint8
	DX  uint8
	EAX uint32
	EBX uint32
	ECX uint32
	EDX uint32
	SI  uint32
	DI  uint32
}

//*=====================================| KERNEL |=====================================\\

// Estructura que contiene el path del archivo que se utilizará como base para ejecutar un nuevo proceso y su PID asociado.
type IniciarProceso struct {
	// Path del archivo que se utilizará como base para ejecutar un nuevo proceso
	PID  uint32 `json:"pid"`
	Path string `json:"path"`
}

type ResponseEstadoProceso struct {
	// Path del archivo que se utilizará como base para ejecutar un nuevo proceso
	State string `json:"state"`
}

// Estructura de respuesta al iniciar un proceso
type ResponseListarProceso struct {
	PID    uint32 `json:"pid"`
	Estado string `json:"estado"`
}

// Estructura de comunicacion al desalojar un proceso
type RespuestaDispatch struct {
	MotivoDeDesalojo string
	PCB              PCB
}

//*====================================| RECURSOS | ====================================\\

type RequestRecurso struct {
	PidSolicitante uint32
	NombreRecurso  string
}

type Recurso struct {
	Instancias int
	ListaBlock ListaSegura
}

//*=====================================| I/O |=====================================\\

// Estructura basica de InterfazIO que se guardara Kernel
type Interfaz struct {
	TipoInterfaz   string
	PuertoInterfaz int
	IpInterfaz     string
}

// Estructura de comunicacion al conectar una interfaz (Contiene su nombre/identificador y lo necesario para validar en Kernel)
type RequestConectarInterfazIO struct {
	NombreInterfaz string
	Interfaz       Interfaz
}

// Estructura de comunicacion entre CPU y Kernel para ejecutar una instruccion de I/O
type RequestEjecutarInstruccionIO struct {
	PidDesalojado  uint32
	NombreInterfaz string
	Instruccion    string
	UnitWorkTime   int
	Direccion      uint32
	Tamaño         uint32
	NombreArchivo  string
	PunteroArchivo uint32
}

// Estructura que comunica Kernel con CPU y CPU con memoria para la instruccion STDIN.
type RequestInputSTDIN struct {
	Pid               uint32
	RegistroDireccion uint32
	TextoUsuario      []byte
	//? Tambien deberia estar el pid?
}

type MetadataFS struct {
	InitialBlock int `json:"initial_block"`
	Size         int `json:"size"`
}

//*=====================================| MEMORIA DE USUARIO |=====================================\\

// Tabla de páginas. Es un slice de marcos que contiene solamente las páginas validadas.
type Tabla []int

type RequestMovOUT struct {
	Pid  uint32
	Dir  uint32
	Data []byte
}

type Fetch struct {
	Page_Size   uint
	Instruccion string
}

//*=====================================| TADs SINCRONIZACION |=====================================|

// ----------------------( LISTA )----------------------\\
type ListaSegura struct {
	Mx   sync.Mutex
	List []uint32
}

func (sList *ListaSegura) Append(value uint32) {
	sList.Mx.Lock()
	sList.List = append(sList.List, value)
	sList.Mx.Unlock()
}

func (sList *ListaSegura) Dequeue() uint32 {
	sList.Mx.Lock()
	var pcb = sList.List[0]
	sList.List = sList.List[1:]
	sList.Mx.Unlock()

	return pcb
}
