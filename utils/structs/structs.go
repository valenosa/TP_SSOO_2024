package structs

// =====================================| ESTRUCTURAS BASE DE UN PROCESO |========================================================\\

//	Kernel -> Cliente
//
// Estructura que contiene el path del archivo que se utilizará como base para ejecutar un nuevo proceso y su PID asociado.
type BodyIniciarProceso struct {
	// Path del archivo que se utilizará como base para ejecutar un nuevo proceso
	PID  uint32 `json:"pid"`
	Path string `json:"path"`
}

type RequestIniciarProceso struct {
	Path string `json:"path"`
}

// Kernel -> Cliente
type ResponseEstadoProceso struct {
	// Path del archivo que se utilizará como base para ejecutar un nuevo proceso
	State string `json:"state"`
}

// CPU, Kernel.
// Estructura base de un proceso.
type PCB struct {
	PID     uint32
	Quantum uint16
	Estado  string
	RegistrosUsoGeneral
}

// CPU
// Estructura de los registros de uso general (para tener info del contexto de ejecución de cada PCB)
type RegistrosUsoGeneral struct {
	PC  uint32
	AX  uint8
	BX  uint8
	CX  uint8
	DX  uint8
	EAX uint16
	EBX uint16
	ECX uint16
	EDX uint16
	SI  uint32
	DI  uint32
}

//=====================================|  |========================================================\\

// Kernel, Memoria
// Estructura de respuesta al iniciar un proceso
type ResponseListarProceso struct {
	PID    uint32 `json:"pid"`
	Estado string `json:"estado"`
}

// Memoria
// Variable global para llevar la cuenta de los procesos (y así poder nombrarlos de manera correcta)
var Counter int = 0

// CPU, Kernel
// TODO: Completar
type IO_GEN_SLEEP struct {
	Instruccion       string
	NombreInterfaz    string
	UnidadesDeTrabajo int
}

// CPU, Kernel
// Estructura de comunicacion al desalojar un proceso
type RespuestaDispatch struct {
	MotivoDeDesalojo string
	PCB              PCB
}

// CPU, Kernel
// Estructura de comunicacion al ejecutar una instrucción
type InstruccionIO struct {
	PidDesalojado  uint32
	NombreInterfaz string
	Instruccion    string
	UnitWorkTime   int
}

// Kernel, I/O
type RequestInterfaz struct {
	NombreInterfaz string
	Interfaz       Interfaz
}

// Kernel, I/O
type Interfaz struct {
	TipoInterfaz   string
	PuertoInterfaz int
	QueueBlock     []uint32
}

// Kernel, I/O
type NuevaInterfaz struct {
	Nombre string
	Tipo   string
	Puerto int
}
