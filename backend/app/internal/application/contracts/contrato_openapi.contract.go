package contracts

// IContratoOpenAPI defines operations to load and merge the OpenAPI contract.
type IContratoOpenAPI interface {
	ObtenerContrato() ([]byte, error)
}
