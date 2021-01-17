package datastore

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func (ds *DataStore) loadServiceMap(f io.ReadCloser) {
	if f == nil {
		return
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		l := strings.TrimSpace(s.Text())
		if len(l) < 1 || strings.HasPrefix(l, "#") {
			continue
		}
		f := strings.Fields(l)
		if len(f) < 2 {
			continue
		}
		sn := f[0]
		a := strings.Split(f[1], "/")
		if len(a) > 1 {
			sn += "/" + a[1]
		}
		ds.serviceMap[f[1]] = sn
	}
}

func (ds *DataStore) GetServiceName(prot, port int) (string, bool) {
	if p, ok := ds.protMap[prot]; ok {
		k := fmt.Sprintf("%d/%s", port, p)
		if s, ok := ds.serviceMap[k]; ok {
			return s, true
		}
		return p, false
	}
	return fmt.Sprintf("prot(%d)", prot), false
}
