package datastore

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sleepinggenius2/gosmi/parser"
	gomibdb "github.com/twsnmp/go-mibdb"
)

func (ds *DataStore) loadMIBDB(f io.ReadCloser) {
	if f == nil {
		return
	}
	defer f.Close()
	if s, err := ioutil.ReadAll(f); err == nil {
		mibdb, err := gomibdb.NewMIBDBFromStr(string(s), "")
		if err != nil {
			log.Printf("loadMIBDB err=%v", err)
			return
		}
		ds.MIBDB = mibdb
	} else {
		log.Printf("loadMIBDB err=%v", err)
	}
}

func (ds *DataStore) loadExtMIBs(root string) {
	if ds.MIBDB == nil {
		return
	}
	filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("loadExtMIBs prevent panic by handling failure accessing a path %q: %v\n", path, err)
				return err
			}
			if info.IsDir() {
				return nil
			}
			ds.loadExtMIB(filepath.Join(root, path))
			return nil
		})
}

func (ds *DataStore) loadExtMIB(path string) {
	var nameList []string
	var mapNameToOID = make(map[string]string)
	for _, name := range ds.MIBDB.GetNameList() {
		mapNameToOID[name] = ds.MIBDB.NameToOID(name)
	}
	asn1, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	module, err := parser.Parse(bytes.NewReader(asn1))
	if err != nil || module == nil {
		return
	}
	if module.Body.Identity != nil {
		name := module.Body.Identity.Name.String()
		oid := getOid(&module.Body.Identity.Oid)
		mapNameToOID[name] = oid
		nameList = append(nameList, name)
	}
	for _, n := range module.Body.Nodes {
		if n.Name.String() == "" || n.Oid == nil {
			continue
		}
		name := n.Name.String()
		mapNameToOID[name] = getOid(n.Oid)
		nameList = append(nameList, name)
	}
	for _, name := range nameList {
		oid, ok := mapNameToOID[name]
		if !ok {
			log.Printf("Can not find mib name %s", name)
			continue
		}
		a := strings.SplitN(oid, ".", 2)
		if len(a) < 2 {
			log.Printf("Can not split mib name=%s oid=%s", name, oid)
			continue
		}
		noid, ok := mapNameToOID[a[0]]
		if !ok {
			log.Printf("Can not split mib name=%s oid=%s", name, a[0])
			continue
		}
		mapNameToOID[name] = noid + "." + a[1]
	}
	for _, name := range nameList {
		_ = ds.MIBDB.Add(name, mapNameToOID[name])
	}
}

func getOid(oid *parser.Oid) string {
	ret := ""
	for _, o := range oid.SubIdentifiers {
		if o.Name != nil {
			ret += o.Name.String()
		}
		if o.Number != nil {
			ret += fmt.Sprintf(".%d", int(*o.Number))
		}
	}
	return ret
}
