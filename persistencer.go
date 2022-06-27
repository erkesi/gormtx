package gormtx

type Persistencer interface {
     DBTxManager() DBTxManager
}
