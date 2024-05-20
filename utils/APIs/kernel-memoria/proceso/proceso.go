package proceso

//=====================================| CLIENT SIDE |========================================================\\

//-------------------------- VARIABLES && STRUCTS ------------------------------------

type BodyIniciar struct {
	// Path del archivo que se utilizará como base para ejecutar un nuevo proceso
	PID  uint32 `json:"pid"`
	Path string `json:"path"`
}

//-------------------------- FUNCIONES AUX -------------------------------------------

// Estructura de los PCB
type PCB struct {
	PID     uint32
	PC      uint32
	Quantum uint16
	Estado  string
	RegistrosUsoGeneral
}

// Estructura de los registros de uso general (para tener info del contexto de ejecución de cada PCB)
type RegistrosUsoGeneral struct {
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

//=====================================| SERVER SIDE |========================================================\\

//-------------------------- VARIABLES && STRUCTS ------------------------------------

type Response struct {
	PID    uint32 `json:"pid"`
	Estado string `json:"estado"`
}

// Variable global para llevar la cuenta de los procesos (y así poder nombrarlos de manera correcta)
var Counter int = 0
