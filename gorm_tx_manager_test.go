package gormtx

import (
	"context"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
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

var txManager DBTxManager
var mainDB, backupDB *gorm.DB

func TestGormTxManager_rollback(t *testing.T) {
	var err error
	ctx := context.TODO()
	initDataFail(ctx, txManager)

	if txManager.MainDB(ctx) != mainDB {
		t.Fatal(txManager.MainDB(ctx))
		return
	}

	if txManager.BackupDB() != backupDB {
		t.Fatal(txManager.BackupDB())
		return
	}

	if txManager.AutoDB(ctx) != backupDB {
		t.Fatal(txManager.AutoDB(ctx))
		return
	}

	var count int64
	err = txManager.MainDB(ctx).Model(&User{}).Count(&count).Error
	if err != nil {
		t.Fatal(err)
		return
	}
	if count != 0 {
		t.Fatal(count)
		return
	}

	err = txManager.MainDB(ctx).Model(&Product{}).Count(&count).Error
	if err != nil {
		t.Fatal(err)
		return
	}
	if count != 0 {
		t.Fatal(count)
		return
	}
}

func TestGormTxManager_Tx(t *testing.T) {
	var err error
	ctx := context.TODO()

	// open && close DB
	ctx, txid := txManager.OpenTx(ctx)
	defer txManager.CloseTx(ctx, txid, &err)

	err = initData(ctx, txManager)
	if err != nil {
		t.Fatal(err)
		return
	}

	// read DB
	var product Product
	txManager.MainDB(ctx).First(&product, "code = ?", "D43") // 查找 code 字段值为 D42 的记录
	if product.Price != 100 {
		t.Fatal(product.Price)
		return
	}
	var user User
	txManager.MainDB(ctx).First(&user, "name = ?", "D42") // 查找 code 字段值为 D42 的记录
	if user.Name != "D42" {
		t.Fatal(user.Name)
		return
	}
}

func initData(ctx context.Context, txManager DBTxManager) error {
	dbtx := txManager.MustMainTx(ctx)

	if txManager.AutoDB(ctx) != dbtx {
		panic(txManager.AutoDB(ctx))
	}

	// Create
	err := dbtx.WithContext(ctx).Create(&Product{Code: "D43", Price: 100}).Error
	if err != nil {
		return err
	}
	// Create
	err = dbtx.WithContext(ctx).Create(&User{Name: "D42"}).Error
	return err
}

func initDataFail(ctx context.Context, txManager DBTxManager) {
	var err error
	defer func() {
		if err == nil || err.Error() != "UNIQUE constraint failed: users.id" {
			panic(err)
		}
	}()
	// open && close DB
	ctx, txid := txManager.OpenTx(ctx)
	defer txManager.CloseTx(ctx, txid, &err)

	dbtx := txManager.MustMainTx(ctx)

	err = dbtx.WithContext(ctx).Create(&Product{Code: "D43", Price: 100}).Error
	if err != nil {
		return
	}

	err = dbtx.WithContext(ctx).Create(&User{Model: gorm.Model{
		ID: 1,
	}, Name: "D42"}).Error
	if err != nil {
		return
	}

	err = dbtx.WithContext(ctx).Create(&User{Model: gorm.Model{
		ID: 1,
	}, Name: "D42"}).Error

}

func setup() {
	_ = os.Remove("testdata/main.db")
	_ = os.Remove("testdata/backup.db")
	var err error
	mainDB, err = gorm.Open(sqlite.Open("testdata/main.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	backupDB, err = gorm.Open(sqlite.Open("testdata/backup.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	// 迁移 schema
	err = mainDB.AutoMigrate(&Product{}, &User{})
	if err != nil {
		panic(err)
	}
	err = backupDB.AutoMigrate(&Product{}, &User{})
	if err != nil {
		panic(err)
	}
	txManager = NewGormTxManager(mainDB, backupDB)
}

func TestMain(m *testing.M) {
	setup()
	_ = m.Run()
}
