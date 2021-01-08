package datastore

import (
	"io/ioutil"

	gomibdb "github.com/twsnmp/go-mibdb"
	"go.etcd.io/bbolt"
)

func (ds *DataStore) LoadMIBDB(s string) error {
	mibdb, err := gomibdb.NewMIBDBFromStr(string(s), "")
	if err != nil {
		return err
	}
	ds.MIBDB = mibdb
	return nil
}

func (ds *DataStore) GetMIBModuleList() []string {
	ret := []string{}
	if ds.db == nil {
		return ret
	}
	_ = ds.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("mibdb"))
		if b == nil {
			return nil
		}
		_ = b.ForEach(func(k, v []byte) error {
			ret = append(ret, string(k))
			return nil
		})
		return nil
	})
	return ret
}

func (ds *DataStore) GetMIBModule(m string) []byte {
	ret := []byte{}
	if ds.db == nil {
		return ret
	}
	_ = ds.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("mibdb"))
		if b == nil {
			return nil
		}
		ret = b.Get([]byte(m))
		return nil
	})
	return ret
}

func (ds *DataStore) PutMIBFileToDB(m, path string) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("mibdb"))
		return b.Put([]byte(m), d)
	})
}

func (ds *DataStore) DelMIBModuleFromDB(m string) error {
	if ds.db == nil {
		return ErrDBNotOpen
	}
	return ds.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("mibdb"))
		return b.Delete([]byte(m))
	})
}
