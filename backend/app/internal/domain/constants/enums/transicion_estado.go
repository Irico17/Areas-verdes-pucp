package enums

// TransicionesEstado es la tabla mínima de RF-16.
// La clave es el código de origen. El valor son los destinos permitidos.
// Los códigos no se renombran: pendiente es «Por iniciar», cerrada es «Cerrado»,
// cancelada es «Cancelado» y archivada es «Archivado».
// por iniciar → en proceso → ejecutado → cerrado.
// Cancelado y archivado salen de los abiertos.
// bloqueada no es destino: queda en datos, inactiva y provisional, hasta que el cliente decida.
// Si una labor ya está bloqueada, puede volver a en proceso o salir por cancelada o archivada.
var TransicionesEstado = map[string][]string{
	"sin_estado": {"pendiente", "cancelada", "archivada"},
	"pendiente":  {"en_proceso", "cancelada", "archivada"},
	"en_proceso": {"ejecutado", "cancelada", "archivada"},
	"ejecutado":  {"cerrada", "cancelada", "archivada"},
	"bloqueada":  {"en_proceso", "cancelada", "archivada"},
}

// TransicionEstadoPermitida reports whether the declared table allows the change.
// Quedarse en el mismo código no es un salto.
func TransicionEstadoPermitida(desde, hasta string) bool {
	if desde == hasta {
		return true
	}
	for _, destino := range TransicionesEstado[desde] {
		if destino == hasta {
			return true
		}
	}
	return false
}
