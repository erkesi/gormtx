package gormtx

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"

	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type NameDB struct {
	Name string
	DB   *gorm.DB
}

type ctxTxId struct {
}

type ctxDBReadOnly struct {
}

var txidCtxKey = &ctxTxId{}

var dbReadOnlyCtxKey = &ctxDBReadOnly{}

const (
	main   = "main"
	backup = "backup"
)

type dbtx struct {
	db, tx *gorm.DB
}

// GormTxManager  db 事务管理器
// mainDB, backupDB *gorm.DB // 默认主从库实例
// tid2Tx            sync.Map // tid（事务ID） 对应的 事务
type GormTxManager struct {
	i                uint64
	mainDB, backupDB *gorm.DB
	db2Name          map[*gorm.DB]string
	tid2Tx           sync.Map
	m                sync.Mutex
}

// NewGormTxManagerWithClauses 使用 DBResolver 读写分离
func NewGormTxManagerWithClauses(db *gorm.DB) *GormTxManager {
	mainDB, backupDB := db.Clauses(dbresolver.Write), db.Clauses(dbresolver.Read)
	return NewGormTxManager(mainDB, backupDB)
}

func NewGormTxManager(mainDB, backupDB *gorm.DB) *GormTxManager {
	db2Name := map[*gorm.DB]string{mainDB: main, backupDB: backup}
	return &GormTxManager{
		mainDB:   mainDB,
		backupDB: backupDB,
		db2Name:  db2Name,
	}
}

// ReadOnly 不使用事务来查询
func (s *GormTxManager) ReadOnly(ctx context.Context) context.Context {
	return context.WithValue(ctx, dbReadOnlyCtxKey, struct{}{})
}

// OpenMainTx 开启 main库 事务
// return 新的context 与 事务ID
func (s *GormTxManager) OpenMainTx(ctx context.Context, opts ...Option) (context.Context, uint64) {
	return s.addTx(ctx, s.mainDB, opts...)
}

// CloseMainTx 关闭 main库 事务
// 参数：
// ctx:	开启 main库 事务返回的新的context
// txid: 开启 main库 事务返回的事务ID
// err: 判断是提交事务还是回滚事务
func (s *GormTxManager) CloseMainTx(ctx context.Context, txid uint64, err *error) {
	isRecover := false
	tmpErr := *err
	if tmpErr == nil {
		r := recover()
		if r != nil {
			isRecover = true
			tmpErr = fmt.Errorf("%v", r)
		}
	}
	*err = s.closeTx(ctx, s.mainDB, txid, tmpErr)
	if isRecover {
		panic(tmpErr)
	}
}

// MainDB 获取 main db，如果已经开启 main db tx，则返回 main db tx
func (s *GormTxManager) MainDB(ctx context.Context) *gorm.DB {
	db, _ := s.tx(ctx, s.mainDB)
	return db
}

// MustMainTx 获取 main db tx，如果已经开启 main db tx，则返回 main db tx，否则 panic
func (s *GormTxManager) MustMainTx(ctx context.Context) *gorm.DB {
	db, ok := s.tx(ctx, s.mainDB)
	if !ok {
		panic(errors.New("not open database transaction"))
	}
	return db
}

// AutoDB 获取 db，如果已经开启 main db tx，则返回 db tx，否则 返回 backup db
func (s *GormTxManager) AutoDB(ctx context.Context) *gorm.DB {
	if db, ok := s.tx(ctx, s.mainDB); ok {
		return db
	}
	return s.BackupDB()
}

// BackupDB 获取 Backup db
func (s *GormTxManager) BackupDB() *gorm.DB {
	return s.backupDB
}

func (s *GormTxManager) tx(ctx context.Context, db *gorm.DB) (*gorm.DB, bool) {
	readonly := ctx.Value(dbReadOnlyCtxKey)
	if readonly != nil {
		return db, false
	}
	curTids := ctx.Value(txidCtxKey)
	if curTids == nil {
		return db, false
	}
	tids := curTids.([]uint64)
	for _, tid := range tids {
		dt, ok := s.tid2Tx.Load(tid)
		if !ok {
			continue
		}
		dbtx := dt.(*dbtx)
		if dbtx.db == db {
			return dbtx.tx, true
		}
	}
	return db, false
}

func (s *GormTxManager) addTx(ctx context.Context, db *gorm.DB, opts ...Option) (context.Context, uint64) {
	option := &options{}
	for _, opt := range opts {
		opt(option)
	}
	txids := s.txids(ctx)
	if !option.StartupNewTx {
		for _, tid := range txids {
			dt, ok := s.tid2Tx.Load(tid)
			if !ok {
				continue
			}
			dbtx := dt.(*dbtx)
			if dbtx.db == db {
				return ctx, 0
			}
		}
	}
	tid := s.increaseTid()
	newTids := make([]uint64, 0, 16)
	newTids = append(newTids, tid)
	newTids = append(newTids, txids...)
	ctx = context.WithValue(ctx, txidCtxKey, newTids)
	s.tid2Tx.Store(tid, &dbtx{
		db: db,
		tx: db.Begin(),
	})
	return ctx, tid
}

func (s *GormTxManager) increaseTid() uint64 {
	if s.i > math.MaxInt64 {
		s.m.Lock()
		if s.i > math.MaxInt64 {
			s.i = 1
		}
		s.m.Unlock()
	}
	return atomic.AddUint64(&s.i, 1)
}

func (s *GormTxManager) closeTx(ctx context.Context, db *gorm.DB, tid uint64, err error) error {
	if tid == 0 { // 表示没有开启新的事务
		return err
	}
	dbName := s.db2Name[db]
	dt, ok := s.tid2Tx.Load(tid)
	if !ok {
		return fmt.Errorf("%s database transaction closed",dbName)
	}
	s.tid2Tx.Delete(tid)
	tx := dt.(*dbtx).tx
	if err != nil {
		dbErr := tx.Rollback().Error
		if dbErr != nil {
			return fmt.Errorf("%s database transaction rollback error:%v, warp:%w", dbName, dbErr, err)
		}
		return err
	}
	dbErr := tx.Commit().Error
	if dbErr != nil {
		return fmt.Errorf("%s database transaction commit error:%w", dbName, dbErr)
	}
	return err
}

func (s *GormTxManager) txids(ctx context.Context) []uint64 {
	curTxids := ctx.Value(txidCtxKey)
	txids := make([]uint64, 0, 16)
	if curTxids != nil {
		txids = curTxids.([]uint64)
	}
	return txids
}
