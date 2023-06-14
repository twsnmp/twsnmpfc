package datastore

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
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
	EnumMap     map[int]string
}

type MIBTreeEnt struct {
	OID      string `json:"oid"`
	Name     string `json:"name"`
	MIBInfo  *MIBInfo
	Children []*MIBTreeEnt `json:"children"`
}

// 読み込んだMIBのリスト
type MIBModuleEnt struct {
	Type  string // int | ext
	File  string
	Name  string
	Error string
}

var MIBTree = []*MIBTreeEnt{}

var MIBInfoMap = make(map[string]MIBInfo)

var MIBModules = []*MIBModuleEnt{}

func FindMIBInfo(name string) *MIBInfo {
	a := strings.SplitN(name, ".", 2)
	if len(a) == 2 {
		name = a[0]
	}
	oid := MIBDB.NameToOID(name)
	if i, ok := MIBInfoMap[oid]; ok {
		return &i
	}
	return nil
}

func loadMIBDB(fs http.FileSystem) {
	// 名前とOIDだけの定義の読み込み（互換性）
	if r, err := os.Open(filepath.Join(dspath, "mib.txt")); err == nil {
		loadMIBDBNameOnly(r)
	} else {
		if r, err := fs.Open("/conf/mib.txt"); err == nil {
			loadMIBDBNameOnly(r)
		}
	}
	// 組み込みのMIB定義を読み込む
	loadMIBsFromFS(fs)
	// 拡張MIBの読み込み
	loadExtMIBs(filepath.Join(dspath, "extmibs"))
	// MIB情報をOIDで検索できるようにする
	checkMIBInfoMap()
	// MIB2の説明を翻訳版に入れ替え
	setMIB2Descr(fs)
	// MIBツリーを作成する
	makeMibTreeList()
}

func loadMIBDBNameOnly(f io.ReadCloser) {
	if f == nil {
		return
	}
	defer f.Close()
	if s, err := io.ReadAll(f); err == nil {
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

var mibs = `AGENTX-MIB.txt
BRIDGE-MIB.txt
DISMAN-EVENT-MIB.txt
DISMAN-SCHEDULE-MIB.txt
DISMAN-SCRIPT-MIB.txt
EtherLike-MIB.txt
HCNUM-TC.txt
HOST-RESOURCES-MIB.txt
HOST-RESOURCES-TYPES.txt
IANA-ADDRESS-FAMILY-NUMBERS-MIB.txt
IANA-LANGUAGE-MIB.txt
IANA-RTPROTO-MIB.txt
IANAifType-MIB.txt
IF-INVERTED-STACK-MIB.txt
IF-MIB.txt
INET-ADDRESS-MIB.txt
IP-FORWARD-MIB.txt
IP-MIB.txt
IPV6-FLOW-LABEL-MIB.txt
IPV6-ICMP-MIB.txt
IPV6-MIB.txt
IPV6-TC.txt
IPV6-TCP-MIB.txt
IPV6-UDP-MIB.txt
NET-SNMP-AGENT-MIB.txt
NET-SNMP-EXAMPLES-MIB.txt
NET-SNMP-EXTEND-MIB.txt
NET-SNMP-MIB.txt
NET-SNMP-PASS-MIB.txt
NET-SNMP-TC.txt
NET-SNMP-VACM-MIB.txt
NOTIFICATION-LOG-MIB.txt
RFC-1215.txt
RFC1155-SMI.txt
RFC1213-MIB.txt
RMON-MIB.txt
RMON2.txt
HC-RMON-MIB.txt
SCTP-MIB.txt
SMUX-MIB.txt
SNMP-COMMUNITY-MIB.txt
SNMP-FRAMEWORK-MIB.txt
SNMP-MPD-MIB.txt
SNMP-NOTIFICATION-MIB.txt
SNMP-PROXY-MIB.txt
SNMP-TARGET-MIB.txt
SNMP-USER-BASED-SM-MIB.txt
SNMP-USM-AES-MIB.txt
SNMP-USM-DH-OBJECTS-MIB.txt
SNMP-VIEW-BASED-ACM-MIB.txt
SNMPv2-CONF.txt
SNMPv2-MIB.txt
SNMPv2-SMI.txt
SNMPv2-TC.txt
SNMPv2-TM.txt
TCP-MIB.txt
TRANSPORT-ADDRESS-MIB.txt
TUNNEL-MIB.txt
UCD-DEMO-MIB.txt
UCD-DISKIO-MIB.txt
UCD-DLMOD-MIB.txt
UCD-IPFWACC-MIB.txt
UCD-SNMP-MIB.txt
UDP-MIB.txt`

func loadMIBsFromFS(fs http.FileSystem) {
	skipList := []string{}
	for _, m := range strings.Split(mibs, "\n") {
		m = strings.TrimSpace(m)
		if m == "" {
			continue
		}
		path := "/conf/mibs/" + m
		log.Printf("load mib path=%s", path)
		if r, err := fs.Open(path); err == nil {
			if asn1, err := io.ReadAll(r); err == nil {
				if !loadExtMIB(asn1, "int", path) {
					skipList = append(skipList, path)
				}
			} else {
				log.Println(err)
			}
		} else {
			log.Println(err)
		}
	}
	for _, path := range skipList {
		log.Printf("retry to load mib path=%s", path)
		if r, err := fs.Open(path); err == nil {
			if asn1, err := io.ReadAll(r); err == nil {
				if loadExtMIB(asn1, "int", path) {
					log.Printf("skip error mib file=%s", path)
				}
			}
		}
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
			log.Printf("load ext mib path=%s", path)
			if asn1, err := os.ReadFile(path); err == nil {
				if loadExtMIB(asn1, "ext", path) {
					skipList = append(skipList, path)
				}
			} else {
				log.Printf("load ext mib err=%v", err)
			}
			return nil
		})
	for _, path := range skipList {
		log.Printf("retry to load ext mib path=%s", path)
		if asn1, err := os.ReadFile(path); err == nil {
			if loadExtMIB(asn1, "ext", path) {
				log.Printf("skip error mib file=%s", path)
			}
		}
	}
}

func loadExtMIB(asn1 []byte, fileType, file string) bool {
	var nameList []string
	var mapNameToOID = make(map[string]string)
	for _, name := range MIBDB.GetNameList() {
		mapNameToOID[name] = MIBDB.NameToOID(name)
	}
	module, err := parser.Parse(bytes.NewReader(asn1))
	if err != nil || module == nil {
		log.Printf("loadExtMIB err=%v", err)
		modErr := err
		mod := "Unknown"
		if module != nil {
			mod = string(module.Name)
		}
		asn1 = rfc2mib(asn1)
		module, err = parser.Parse(bytes.NewReader(asn1))
		if err != nil || module == nil {
			log.Printf("try rfc2mib loadExtMIB err=%v", err)
			if module != nil {
				mod = string(module.Name)
			}
			MIBModules = append(MIBModules, &MIBModuleEnt{
				File:  file,
				Type:  fileType,
				Name:  mod,
				Error: modErr.Error() + "\t" + err.Error(),
			})
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
	if !hasSkip {
		MIBModules = append(MIBModules, &MIBModuleEnt{
			File: file,
			Type: fileType,
			Name: string(module.Name),
		})
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
	enumMap := make(map[int]string)
	for _, e := range n.ObjectType.Syntax.Type.Enum {
		enum = append(enum, fmt.Sprintf("%s:%s ", e.Value, e.Name))
		if i, err := strconv.Atoi(e.Value); err == nil {
			enumMap[i] = string(e.Name)
		}
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
		EnumMap:     enumMap,
	}
}

// checkMIBInfoMap : MIB情報を数値のOIDをキーとしたMAPへ変換する
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

func setMIB2Descr(fs http.FileSystem) {
	r, err := fs.Open("/conf/mib2descr.txt")
	if err != nil {
		return
	}
	rg := regexp.MustCompile(`^#(\S+)`)
	all, err := io.ReadAll(r)
	if err != nil {
		return
	}
	name := ""
	descr := []string{}
	for _, l := range strings.Split(string(all), "\n") {
		m := rg.FindStringSubmatch(l)
		if len(m) > 1 {
			if name != "" && len(descr) > 0 {
				replaceMIBDescr(name, descr)
			}
			name = m[1]
			descr = []string{}
		} else {
			strings.ReplaceAll(l, `"`, "")
			descr = append(descr, l)
		}
	}
	if name != "" && len(descr) > 0 {
		replaceMIBDescr(name, descr)
	}
}

func replaceMIBDescr(name string, descr []string) {
	oid := MIBDB.NameToOID(name)
	if e, ok := MIBInfoMap[oid]; ok {
		e.Description = strings.Join(descr, "\n")
		MIBInfoMap[oid] = e
	}
}

var (
	mibTreeMAP  = map[string]*MIBTreeEnt{}
	mibTreeRoot *MIBTreeEnt
)

// addToMibTree : MIBツリーへ追加
func addToMibTree(oid, name, poid string) {
	n := &MIBTreeEnt{Name: name, OID: oid, Children: []*MIBTreeEnt{}}
	if i, ok := MIBInfoMap[oid]; ok {
		n.MIBInfo = &i
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
