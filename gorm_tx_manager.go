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

var ctxTxidMask = &ctxTxId{}

const (
	main   = "main"
	backup = "backup"
)

var (
	notOpenTx = errors.New("not open database transaction")
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

// OpenMainTx 开启 main库 事务
// return 新的context 与 事务ID
func (s *GormTxManager) OpenMainTx(ctx context.Context, opts ...DBTxOpt) (context.Context, uint64) {
	return s.addTx(ctx, s.mainDB, opts...)
}

// CloseMainTx 关闭 main库 事务
// 参数：
// ctx:	开启 main库 事务返回的新的context
// txid: 开启 main库 事务返回的事务ID
// err: 判断是提交事务还是回滚事务
func (s *GormTxManager) CloseMainTx(ctx context.Context, txid uint64, err *error) {
	isRecover := false
	errTmp := *err
	if errTmp == nil {
		r := recover()
		if r != nil {
			isRecover = true
			errTmp = fmt.Errorf("%v", r)
		}
	}
	*err = s.closeTx(ctx, s.mainDB, txid, errTmp)
	if isRecover {
		panic(errTmp)
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
		panic(notOpenTx)
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
	curTids := ctx.Value(ctxTxidMask)
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

func (s *GormTxManager) addTx(ctx context.Context, db *gorm.DB, opts ...DBTxOpt) (context.Context, uint64) {
	option := &Option{}
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
	ctx = context.WithValue(ctx, ctxTxidMask, newTids)
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
	dt, ok := s.tid2Tx.Load(tid)
	if !ok {
		return errors.New(fmt.Sprintf("db(%s) tx closed", s.db2Name[db]))
	}
	s.tid2Tx.Delete(tid)
	tx := dt.(*dbtx).tx
	if err != nil {
		err2 := tx.Rollback().Error
		if err2 != nil {
			return fmt.Errorf("tx rollback error:%s, warp:%w", err2.Error(), err)
		}
		return err
	}
	err2 := tx.Commit().Error
	if err2 != nil {
		return fmt.Errorf("tx commit error:%w", err2)
	}
	return err
}

func (s *GormTxManager) txids(ctx context.Context) []uint64 {
	curTxids := ctx.Value(ctxTxidMask)
	txids := make([]uint64, 0, 16)
	if curTxids != nil {
		txids = curTxids.([]uint64)
	}
	return txids
}
