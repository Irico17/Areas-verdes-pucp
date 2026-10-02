package errors

import "errors"

var (
	// ErrSectorDuplicado is returned when a capataz sector code already exists.
	ErrSectorDuplicado = errors.New("ya existe un sector de capataz con ese código")
	// ErrSectorNoExiste is returned when the sector code is not in the catalog.
	ErrSectorNoExiste = errors.New("no existe ese sector de capataz")
	// ErrCodigoSector is returned when the sector code is not a short slug.
	ErrCodigoSector = errors.New("el código del sector usa minúsculas, números y guion")
	// ErrColorSector is returned when the map color is not a six-digit hex.
	ErrColorSector = errors.New("el color es un hexadecimal de seis dígitos")
	// ErrLugarDesconocido is returned when a place is not already in the catalog.
	ErrLugarDesconocido = errors.New("ese lugar no está en el catálogo")
	// ErrEdificioDesconocido is returned when the building id is not in the extract.
	ErrEdificioDesconocido = errors.New("el edificio se elige por id del catálogo")
	// ErrArchivoSinFilas is returned when an import file has nothing to write.
	ErrArchivoSinFilas = errors.New("el archivo no trae filas")
	// ErrViaVacia is returned when a roads file has no lines to import.
	ErrViaVacia = errors.New("el archivo no trae vías")
	// ErrGeomVia is returned when a road is not a line inside the campus.
	ErrGeomVia = errors.New("la vía no es una línea dentro del campus")
)
