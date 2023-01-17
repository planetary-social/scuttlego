package commands

type Migrations struct {
	MigrationDeleteGoSSBRepositoryInOldFormat *MigrationHandlerDeleteGoSSBRepositoryInOldFormat
	MigrationImportDataFromGoSSB              *MigrationHandlerImportDataFromGoSSB
}
