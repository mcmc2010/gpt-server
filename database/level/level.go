package database_level

import (
	"encoding/json"
	"errors"

	"github.com/syndtr/goleveldb/leveldb"
)

type LevelDB struct {
	Filename string
	DB       *leveldb.DB
	Error    error
}

func (I *LevelDB) Set(key string, value any) error {

	text := ""
	data, ok := value.(map[string]any)
	if ok {
		bytes, err := json.Marshal(data)
		if err != nil {
			return err
		}
		text = string(bytes)
	} else {
		text, ok = value.(string)
		if !ok {
			return errors.New("not text format")
		}
	}

	if I.DB == nil {
		return errors.New("db instance is null")
	}

	err := I.DB.Put([]byte(key), []byte(text), nil)
	if err != nil {
		return err
	}

	return nil
}

func (I *LevelDB) GetString(key string) (string, error) {
	if I.DB == nil {
		return "", errors.New("db instance is null")
	}

	data, err := I.DB.Get([]byte(key), nil)
	if err != nil {
		return "", err
	}

	text := string(data)
	return text, nil
}

func (I *LevelDB) Get(key string) (any, error) {
	if I.DB == nil {
		return nil, errors.New("db instance is null")
	}

	data, err := I.DB.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}

	var v any
	err = json.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

var _instance_list []*LevelDB = []*LevelDB{}

func ReleaseAll() {
	for _, v := range _instance_list {
		Release(v)
		v = nil
	}
	_instance_list = []*LevelDB{}
}

func Release(ldb *LevelDB) {
	if ldb == nil {
		return
	}

	ldb.DB.Close()
	ldb.DB = nil

	//
	println("[LevelDB] (work) released ")
}

// Only one instance
func NewAndInitialize(filename string) *LevelDB {

	ldb := &LevelDB{
		Filename: filename,
		DB:       nil,
		Error:    nil,
	}

	// 打开或创建 LevelDB 数据库
	db, err := leveldb.OpenFile(filename, nil)
	if err != nil {
		ldb.Error = err
		println("[LevelDB] (work) error " + err.Error())
		return nil
	}
	ldb.DB = db

	_instance_list = append(_instance_list, ldb)

	println("[LevelDB] (work) starting ")
	return ldb
}
