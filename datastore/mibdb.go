package datastore

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/sleepinggenius2/gosmi/parser"
	gomibdb "github.com/twsnmp/go-mibdb"
)

type MIBTreeEnt struct {
	OID      string        `json:"oid"`
	Name     string        `json:"name"`
	Children []*MIBTreeEnt `json:"children"`
}

var MIBTree = []*MIBTreeEnt{}

func loadMIBDB(f io.ReadCloser) {
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
		MIBDB = mibdb
	} else {
		log.Printf("loadMIBDB err=%v", err)
	}
}

func loadExtMIBs(root string) {
	if MIBDB == nil {
		return
	}
	filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("loadExtMIBs path %q: %v\n", path, err)
				return err
			}
			if info.IsDir() {
				return nil
			}
			loadExtMIB(filepath.Join(root, path))
			return nil
		})
}

func loadExtMIB(path string) {
	var nameList []string
	var mapNameToOID = make(map[string]string)
	for _, name := range MIBDB.GetNameList() {
		mapNameToOID[name] = MIBDB.NameToOID(name)
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
		_ = MIBDB.Add(name, mapNameToOID[name])
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

var (
	mibTreeMAP  = map[string]*MIBTreeEnt{}
	mibTreeRoot *MIBTreeEnt
)

func addToMibTree(oid, name, poid string) {
	n := &MIBTreeEnt{Name: name, OID: oid, Children: []*MIBTreeEnt{}}
	if poid == "" {
		mibTreeRoot = n
	} else {
		p, ok := mibTreeMAP[poid]
		if !ok {
			log.Printf("addToMibTree parentId=%v: not found", poid)
			return
		}
		p.Children = append(p.Children, n)
	}
	mibTreeMAP[oid] = n
}

func makeMibTreeList() {
	oids := []string{}
	for _, n := range MIBDB.GetNameList() {
		oid := MIBDB.NameToOID(n)
		if oid == ".0.0" {
			continue
		}
		oids = append(oids, oid)
	}
	sort.Slice(oids, func(i, j int) bool {
		a := strings.Split(oids[i], ".")
		b := strings.Split(oids[j], ".")
		for k := 0; k < len(a) && k < len(b); k++ {
			l, _ := strconv.Atoi(a[k])
			m, _ := strconv.Atoi(b[k])
			if l == m {
				continue
			}
			if l < m {
				return true
			}
			return false
		}
		return len(a) < len(b)
	})
	addToMibTree(".1.3.6.1", "iso.org.dod.internet", "")
	for _, oid := range oids {
		name := MIBDB.OIDToName(oid)
		if name == "" {
			continue
		}
		lastDot := strings.LastIndex(oid, ".")
		if lastDot < 0 {
			continue
		}
		poid := oid[:lastDot]
		addToMibTree(oid, name, poid)
	}
	if mibTreeRoot != nil {
		MIBTree = append(MIBTree, mibTreeRoot.Children...)
	}
}
