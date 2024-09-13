package client

import (
	"fmt"
	"log"
	"testing"
)

func TestClientNodes(t *testing.T) {
	t.Log("Start")
	c := NewClient("http://127.0.0.1:8080")
	t.Log("Login")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("GetNodes")
	nodes, err := c.GetNodes()
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range nodes {
		t.Logf("%+v", n)
	}
	t.Log("Done")
}

func TestClientPollings(t *testing.T) {
	t.Log("Start")
	c := NewClient("http://127.0.0.1:8080")
	t.Log("Login")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("GetPollings")
	pollings, err := c.GetPollings()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range pollings.Pollings {
		t.Logf("%+v", p)
	}
	t.Log("Done")
}

func TestClientPollingLogs(t *testing.T) {
	t.Log("Start")
	c := NewClient("http://127.0.0.1:8080")
	t.Log("Login")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("GetPollings")
	pollings, err := c.GetPollings()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range pollings.Pollings {
		if p.LogMode > 0 {
			t.Logf("%+v", p)
			logs, err := c.GetPollingLogs(p.ID, &TimeFilter{})
			if err != nil {
				t.Fatal(err)
			}
			for _, l := range logs {
				t.Logf("%+v", l)
			}
		}
	}
	t.Log("Done")
}

func TestClientEventLog(t *testing.T) {
	t.Log("Start")
	c := NewClient("http://127.0.0.1:8080")
	t.Log("Login")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("GetEventLogs")
	logs, err := c.GetEventLogs(&EventLogFilter{})
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range logs.EventLogs {
		t.Logf("%+v", l)
	}
	t.Log("Done")
}
func TestClientSyslog(t *testing.T) {
	t.Log("Start")
	c := NewClient("http://127.0.0.1:8080")
	t.Log("Login")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("GetSyslogs")
	logs, err := c.GetSyslogs(&SyslogFilter{})
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range logs.Logs {
		t.Logf("%+v", l)
	}
	t.Log("Done")
}

func TestClientSnmpTrap(t *testing.T) {
	t.Log("Start")
	c := NewClient("http://127.0.0.1:8080")
	t.Log("Login")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("GetSnmpTraps")
	logs, err := c.GetSnmpTraps(&SnmpTrapFilter{})
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range logs {
		t.Logf("%+v", l)
	}
	t.Log("Done")
}

func TestClientNetFlow(t *testing.T) {
	t.Log("Start")
	c := NewClient("http://127.0.0.1:8080")
	t.Log("Login")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("GetNetFlow")
	r, err := c.GetNetFlow(&NetflowFilter{})
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range r.Logs {
		t.Logf("%+v", l)
	}
	t.Log("Done")
}

func TestClientSFlow(t *testing.T) {
	t.Log("Start")
	c := NewClient("http://127.0.0.1:8080")
	t.Log("Login")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("GetSFlow")
	logs, err := c.GetSnmpTraps(&SnmpTrapFilter{})
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range logs {
		t.Logf("%+v", l)
	}
	t.Log("Done")
}

func TestClientSFlowCounter(t *testing.T) {
	t.Log("Start")
	c := NewClient("http://127.0.0.1:8080")
	t.Log("Login")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("GetSFlowCounter")
	r, err := c.GetSFlowCounter(&SFlowCounterFilter{})
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range r.Logs {
		t.Logf("%+v", l)
	}
	t.Log("Done")
}

func TestClientIPFIX(t *testing.T) {
	t.Log("Start")
	c := NewClient("http://127.0.0.1:8080")
	t.Log("Login")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("GetIPFIX")
	r, err := c.GetIPFIX(&NetflowFilter{})
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range r.Logs {
		t.Logf("%+v", l)
	}
	t.Log("Done")
}

func TestClientDevices(t *testing.T) {
	t.Log("Start")
	c := NewClient("http://127.0.0.1:8080")
	t.Log("Login")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("GetDevices")
	devices, err := c.GetDevices()
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range devices {
		t.Logf("%+v", d)
	}
	t.Log("Done")
}

func TestClientAIList(t *testing.T) {
	t.Log("Start")
	c := NewClient("http://127.0.0.1:8080")
	t.Log("Login")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("GetAIList")
	aiList, err := c.GetAIList()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range aiList {
		t.Logf("%+v", a)
	}
	t.Log("Done")
}

func TestClientNodeEdit(t *testing.T) {
	t.Log("Start")
	c := NewClient("http://127.0.0.1:8080")
	t.Log("Login")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Add Node")
	err = c.UpdateNode(&NodeEnt{
		Name: "test-api-123",
		X:    100,
		Y:    100,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Update Invalid Node")
	err = c.UpdateNode(&NodeEnt{
		ID:   "invalid id",
		Name: "invalid id node",
		X:    200,
		Y:    200,
	})
	if err == nil {
		t.Fatal("invalid id node update with no error")
	}
	t.Log("GetNodes")
	nodes, err := c.GetNodes()
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range nodes {
		if n.Name == "test-api-123" {
			t.Logf("%+v", n)
			t.Log("Delete Node")
			err = c.DeleteNodes([]string{n.ID})
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Log("Done")
}

func TestClientResetReport(t *testing.T) {
	t.Log("Start")
	c := NewClient("http://127.0.0.1:8080")
	t.Log("Login")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Reset Report")
	err = c.ResetReport("devices")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Done")
}

func TestClientDeleteLog(t *testing.T) {
	t.Log("Start")
	c := NewClient("http://127.0.0.1:8080")
	t.Log("Login")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Delete SNMP Trap")
	err = c.DeleteLog("trap")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Done")
}

func ExampleTWSNMPApi_Login() {
	c := NewClient("http://127.0.0.1:8080")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		log.Fatal(err)
	}
}
func ExampleTWSNMPApi_GetNodes() {
	c := NewClient("http://127.0.0.1:8080")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		log.Fatal(err)
	}
	nodes, err := c.GetNodes()
	if err != nil {
		log.Fatal(err)
	}
	for _, n := range nodes {
		fmt.Printf("%+v\n", n)
	}
}
func ExampleTWSNMPApi_UpdateNode() {
	c := NewClient("http://127.0.0.1:8080")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		log.Fatal(err)
	}
	err = c.UpdateNode(&NodeEnt{
		Name: "test",
		X:    100,
		Y:    100,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleTWSNMPApi_DeleteNodes() {
	c := NewClient("http://127.0.0.1:8080")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		log.Fatal(err)
	}
	err = c.DeleteNodes([]string{"node id"})
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleTWSNMPApi_GetPollingLogs() {
	c := NewClient("http://127.0.0.1:8080")
	err := c.Login("twsnmp", "twsnmp")
	if err != nil {
		log.Fatal(err)
	}
	r, err := c.GetPollings()
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range r.Pollings {
		if p.LogMode > 0 {
			logs, err := c.GetPollingLogs(p.ID, &TimeFilter{})
			if err != nil {
				log.Fatal(err)
			}
			for _, l := range logs {
				fmt.Printf("%+v\n", l)
			}
		}
	}
}
