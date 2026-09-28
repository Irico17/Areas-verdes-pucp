package contracts

import "context"

// IArchivoEstaticoAdapter defines access to static reference files (e.g. OSM buildings).
type IArchivoEstaticoAdapter interface {
	LeerEdificios(ctx context.Context) ([]byte, error)
}
