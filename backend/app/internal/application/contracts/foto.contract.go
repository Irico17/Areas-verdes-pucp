// Package contracts defines interfaces for repositories, services, and adapters.
package contracts

// IFotoDiscoAdapter defines disk access to local inventory photos.
type IFotoDiscoAdapter interface {
	RutaFoto(nombre string) (string, error)
}
