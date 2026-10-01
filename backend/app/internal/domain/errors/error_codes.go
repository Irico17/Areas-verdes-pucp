package errors

const (
	SrvInternalServer = "SRV-1000:SRV_INTERNAL_ERROR"

	// ===================================================
	// DB (Database Errors) - Range: 7000-7999
	// ===================================================
	DBDatabaseConnection = "DB-7000:DB_CONNECTION_FAILED"
	DBDatabaseQuery      = "DB-7001:DB_QUERY_FAILED"
	DBDatabaseError      = "DB-7002:DB_GENERIC_ERROR"
)
