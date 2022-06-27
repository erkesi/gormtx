package gormtx

type Persistencer interface {
	DBTxManager() GormTxManager
}
