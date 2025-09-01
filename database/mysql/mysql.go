package mysql

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Data struct {
	gorm.Model `gorm:"embedded"`
	Key        string `gorm:"column:key"`
	Value      []byte `gorm:"column:value"`
}

// Register 连接数据库
func Register(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	//自动迁移
	// if err := db.AutoMigrate(&Data{}); err != nil {
	// 	return nil, err
	// }
	return db, err
}

// Insert 插入
func Insert(db *gorm.DB, datas []*Data) error {
	result := db.Create(datas)
	return result.Error
}

// Init 初始化数据
func Init(db *gorm.DB, datas []*Data) error {
	if result := db.Exec("TRUNCATE TABLE data"); result.Error != nil {
		return result.Error
	}
	return Insert(db, datas)
}

// Select 查找
func Select(db *gorm.DB, key string) ([]Data, error) {
	var rows []Data
	var result *gorm.DB
	if key == "*" {
		result = db.Model(&Data{}).Find(&rows)
	} else {
		result = db.Model(&Data{}).Where("`key`=?", key).Find(&rows)
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return rows, nil
}
