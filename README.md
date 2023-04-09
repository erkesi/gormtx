# gormtx

> gorm 事务管理

## 接口：[db_tx_manager.go](db_tx_manager.go)
```go
type DBTxManager interface {
	// OpenMainTx 开启 main库 事物
	OpenMainTx(ctx context.Context, opts ...Option) (context.Context, uint64)
	// CloseMainTx 关闭 main库 事务
	CloseMainTx(ctx context.Context, txid uint64, err *error)
	// MainDB 获取 main db，如果已经开启 main db tx，则返回 main db tx
	MainDB(ctx context.Context) *gorm.DB
	// BackupDB 获取 Backup db
	BackupDB() *gorm.DB
	// AutoDB 获取 db，如果已经开启 main db tx，则返回 db tx，否则 返回 backup db
	AutoDB(ctx context.Context) *gorm.DB
	// MustMainTx 获取 main db tx，如果已经开启 main db tx，则返回 main db tx，否则 panic
	MustMainTx(ctx context.Context) *gorm.DB
}
```

## 实现：[gorm_tx_manager.go](gorm_tx_manager.go)

## 测试：[gorm_tx_manager_test.go](gorm_tx_manager_test.go)

