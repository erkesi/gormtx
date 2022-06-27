package gormtx

import (
	"context"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

type User struct {
	gorm.Model
	Name string
}

func TestGormTxManager_Tx(t *testing.T) {
	mainDB, err := gorm.Open(sqlite.Open("main.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	backupDB, err := gorm.Open(sqlite.Open("backup.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// 迁移 schema
	mainDB.AutoMigrate(&Product{})
	mainDB.AutoMigrate(&User{})
	backupDB.AutoMigrate(&Product{})
	backupDB.AutoMigrate(&User{})

	ctx := context.TODO()
	txManager := NewGormTxManager(mainDB, backupDB)
	// open && close DB
	ctx, txid := txManager.OpenMainTx(ctx)
	defer txManager.CloseMainTx(ctx, txid, &err)
	// test TX
	err = testTx(ctx, txManager)
	// read DB
	var product Product
	txManager.MainDB(ctx).First(&product, "code = ?", "D43") // 查找 code 字段值为 D42 的记录
	t.Log(product)
	// err
	// err = errors.New("db tx rollback")
}

func testTx(ctx context.Context, txManager DBTxManager) error {
	// Create
	err := txManager.MainDB(ctx).Create(&Product{Code: "D43", Price: 100}).Error
	if err != nil {
		return err
	}
	// Create
	err = txManager.MainDB(ctx).Create(&User{Name: "D42"}).Error
	return err
}
