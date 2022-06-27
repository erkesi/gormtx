type Persistencer interface {
	DBTxManager() GormTxManager
}
