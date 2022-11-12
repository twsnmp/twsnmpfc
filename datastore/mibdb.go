package datastore

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/sleepinggenius2/gosmi/parser"
	gomibdb "github.com/twsnmp/go-mibdb"
)

type MIBInfo struct {
	OID         string
	Status      string
	Type        string
	Enum        string
	Defval      string
	Units       string
	Index       string
	Description string
}

type MIBTreeEnt struct {
	OID      string `json:"oid"`
	Name     string `json:"name"`
	MIBInfo  *MIBInfo
	Children []*MIBTreeEnt `json:"children"`
}

var MIBTree = []*MIBTreeEnt{}

var MIBInfoMap = make(map[string]MIBInfo)

func loadMIBDB(f io.ReadCloser) {
	if f == nil {
		return
	}
	defer f.Close()
	if s, err := ioutil.ReadAll(f); err == nil {
		mibdb, err := gomibdb.NewMIBDBFromStr(string(s), "")
		if err != nil {
			log.Printf("load mibdb err=%v", err)
			return
		}
		MIBDB = mibdb
	} else {
		log.Printf("load mibdb err=%v", err)
	}
}

func loadExtMIBs(root string) {
	if MIBDB == nil {
		return
	}
	skipList := []string{}
	filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if loadExtMIB(path) {
				skipList = append(skipList, path)
			}
			return nil
		})
	for _, path := range skipList {
		if loadExtMIB(path) {
			log.Printf("skip error mib file=%s", path)
		}
	}
	checkMIBInfoMap()
}

func loadExtMIB(path string) bool {
	log.Printf("load ext mib path=%s", path)
	var nameList []string
	var mapNameToOID = make(map[string]string)
	for _, name := range MIBDB.GetNameList() {
		mapNameToOID[name] = MIBDB.NameToOID(name)
	}
	asn1, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		return false
	}
	module, err := parser.Parse(bytes.NewReader(asn1))
	if err != nil || module == nil {
		log.Printf("loadExtMIB err=%v", err)
		asn1 = rfc2mib(asn1)
		module, err = parser.Parse(bytes.NewReader(asn1))
		if err != nil || module == nil {
			log.Printf("try rfc2mib loadExtMIB err=%v", err)
			return false
		}
	}
	if module.Body.Identity != nil {
		name := module.Body.Identity.Name.String()
		oid := getOid(&module.Body.Identity.Oid)
		mapNameToOID[name] = oid
		nameList = append(nameList, name)
		log.Printf("module %s=%s", name, oid)
	}
	for _, n := range module.Body.Nodes {
		if n.Name.String() == "" || n.Oid == nil {
			continue
		}
		name := n.Name.String()
		mapNameToOID[name] = getOid(n.Oid)
		nameList = append(nameList, name)
		setMIBInfo(mapNameToOID[name], &n)
	}
	for _, name := range nameList {
		oid, ok := mapNameToOID[name]
		if !ok {
			log.Printf("no mib name %s", name)
			continue
		}
		a := strings.SplitN(oid, ".", 2)
		if len(a) < 2 {
			continue
		}
		noid, ok := mapNameToOID[a[0]]
		if !ok {
			continue
		}
		mapNameToOID[name] = noid + "." + a[1]
	}
	hasSkip := false
	oidReg := regexp.MustCompile(`^[.0-9]+$`)
	for _, name := range nameList {
		oid := mapNameToOID[name]
		if !oidReg.MatchString(oid) {
			oid = MIBDB.NameToOID(oid)
			if oid == ".0.0" {
				hasSkip = true
				continue
			}
		}
		_ = MIBDB.Add(name, oid)
	}
	return hasSkip
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

func setMIBInfo(oid string, n *parser.Node) {
	if n == nil {
		return
	}
	if n.NotificationType != nil {
		MIBInfoMap[oid] = MIBInfo{
			OID:         oid,
			Status:      n.NotificationType.Status.ToSmi().String(),
			Type:        "Notification",
			Description: n.NotificationType.Description,
		}
		return
	}
	if n.TrapType != nil {
		MIBInfoMap[oid] = MIBInfo{
			OID:         oid,
			Status:      "current",
			Type:        "Notification",
			Description: n.TrapType.Description,
		}
	}
	if n.ObjectType == nil || n.ObjectType.Syntax.Type == nil {
		return
	}
	enum := []string{}
	for _, e := range n.ObjectType.Syntax.Type.Enum {
		enum = append(enum, fmt.Sprintf("%s:%s ", e.Value, e.Name))
	}
	defval := ""
	if n.ObjectType.Defval != nil {
		defval = *n.ObjectType.Defval
	}
	index := []string{}
	for _, i := range n.ObjectType.Index {
		index = append(index, i.Name.String())
	}

	MIBInfoMap[oid] = MIBInfo{
		OID:         oid,
		Status:      n.ObjectType.Status.ToSmi().String(),
		Type:        n.ObjectType.Syntax.Type.Name.String(),
		Enum:        strings.Join(enum, ","),
		Defval:      defval,
		Units:       n.ObjectType.Units,
		Index:       strings.Join(index, ","),
		Description: n.ObjectType.Description,
	}
}

func checkMIBInfoMap() {
	delList := []string{}
	addList := []MIBInfo{}
	for oid, info := range MIBInfoMap {
		noid := MIBDB.NameToOID(oid)
		if noid != oid {
			delList = append(delList, oid)
			info.OID = noid
			addList = append(addList, info)
		}
	}
	for _, d := range delList {
		delete(MIBInfoMap, d)
	}
	for _, a := range addList {
		MIBInfoMap[a.OID] = a
	}
}

var (
	mibTreeMAP  = map[string]*MIBTreeEnt{}
	mibTreeRoot *MIBTreeEnt
)

func addToMibTree(oid, name, poid string) {
	n := &MIBTreeEnt{Name: name, OID: oid, Children: []*MIBTreeEnt{}}
	if i, ok := MIBInfoMap[oid]; ok {
		n.MIBInfo = &i
		log.Printf("addMinTree MIBInfo=%v", n.MIBInfo)
	}
	if poid == "" {
		mibTreeRoot = n
	} else {
		p, ok := mibTreeMAP[poid]
		if !ok {
			log.Printf("no parent name=%s oid=%s poid=%s", name, oid, poid)
			return
		}
		p.Children = append(p.Children, n)
	}
	mibTreeMAP[oid] = n
}

func makeMibTreeList() {
	oids := []string{}
	minLen := len(".1.3.6.1")
	for _, n := range MIBDB.GetNameList() {
		oid := MIBDB.NameToOID(n)
		if len(oid) <= minLen {
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

func rfc2mib(b []byte) []byte {
	// Remove all carriage returns from the input data
	rp := strings.NewReplacer("\r", "")
	all := rp.Replace(string(b))

	// Extract headers/footers from the document
	regPageBreak := regexp.MustCompile(`[^\n]*\n+\f\n+[^\n]*`)
	all = regPageBreak.ReplaceAllString(all, "\n\n")

	// Replace all occurances of 3 or more newlines with two newlines
	// (ie., at most one blank line between paragraphs/sections/etc.)
	regOver3nl := regexp.MustCompile(`\n{3,}`)
	all = regOver3nl.ReplaceAllString(all, "\n\n")

	regMODULEStart := regexp.MustCompile(`\s*([A-Z]+[-A-Za-z0-9]+)+\s+DEFINITIONS\s+\w*\s*::=\s+BEGIN\s*`)
	regMACROStart := regexp.MustCompile(`\s*([A-Z]+[-A-Za-z0-9]+)+\s+MACRO\s+::=\s+BEGIN\s*`)
	regEnd := regexp.MustCompile(`\s*END\s*`)
	regComment := regexp.MustCompile(`\s*(--\s+.*)$`)
	lines := strings.Split(all, "\n")
	depth := 0
	quoted := false
	mibLines := []string{}
	for _, l := range lines {
		if depth == 0 {
			if regMODULEStart.MatchString(l) {
				mibLines = append(mibLines, l)
				depth = 1
			}
			continue
		}
		mibLines = append(mibLines, l)
		if quoted {
			a := strings.Split(l, `"`)
			if len(a) == 1 {
				continue
			}
			if len(a)%2 == 0 {
				quoted = false
			}
			continue
		} else {
			a := strings.Split(l, `"`)
			if len(a) == 2 {
				quoted = true
				continue
			}
			if len(a) != 1 {
				continue
			}
		}
		// Remove Comment
		l = regComment.ReplaceAllString(l, "")
		if regMACROStart.MatchString(l) {
			depth++
			continue
		}
		if regEnd.MatchString(l) {
			depth--
			if depth == 0 {
				return []byte(strings.Join(mibLines, "\n") + "\n")
			}
		}
	}
	return []byte("")
}
