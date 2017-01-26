package mppuma

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"

	mp "github.com/mackerelio/go-mackerel-plugin-helper"
)

// PumaPlugin mackerel plugin for puma
type PumaPlugin struct {
	StateFile string
	Token     string
	Prefix    string
}

// PumaState ...
type PumaState struct {
	ControlURL       string `yaml:"control_url"`
	ControlAuthToken string `yaml:"control_auth_token"`
	PID              uint32 `yaml:"pid"`
}

/*
{
  "workers": 2,
  "phase": 0,
  "booted_workers": 2,
  "old_workers": 0,
  "worker_status": [
    {
      "pid": 780,
      "index": 0,
      "phase": 0,
      "booted": true,
      "last_checkin": "2017-01-25T11:32:26Z",
      "last_status": {
        "backlog": 0,
        "running": 5
      }
    },
    {
      "pid": 784,
      "index": 1,
      "phase": 0,
      "booted": true,
      "last_checkin": "2017-01-25T11:32:26Z",
      "last_status": {
        "backlog": 0,
        "running": 5
      }
    }
  ]
}
*/

// PumaStatus ...
type PumaStatus struct {
	Workers       uint32             `json:"workers"`
	Phase         uint32             `json:"phase"`
	BootedWorkers uint32             `json:"booted_workers"`
	OldWorkers    uint32             `json:"old_workers"`
	WorkerStatus  []PumaWorkerStatus `json:"worker_status"`
}

// PumaWorkerStatus ...
type PumaWorkerStatus struct {
	Pid         uint32               `json:"pid"`
	Index       uint32               `json:"index"`
	Phase       uint32               `json:"phase"`
	LastCheckin string               `json:"last_checkin"`
	LastStatus  PumaWorkerLastStatus `json:"last_status"`
}

// PumaWorkerLastStatus ...
type PumaWorkerLastStatus struct {
	Running uint32 `json:"running"`
	Backlog uint32 `json:"backlog"`
}

// GraphDefinition interface for mackerelplugin
func (m PumaPlugin) GraphDefinition() map[string]mp.Graphs {
	labelPrefix := strings.Title(m.Prefix)
	return map[string]mp.Graphs{
		(m.Prefix + ".status"): {
			Label: (labelPrefix + " Status"),
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "phase", Label: "Worker Phase", Type: "uint32", Stacked: false, Diff: false},
				{Name: "total", Label: "Worker Num", Type: "uint32", Stacked: false, Diff: false},
				{Name: "booted", Label: "Booted worker Num", Type: "uint32", Stacked: true, Diff: false},
				{Name: "old", Label: "Old worker Num", Type: "uint32", Stacked: true, Diff: false},
			},
		},
		(m.Prefix + ".worker_status.#"): {
			Label: (labelPrefix + " Worker Status"),
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "backlog", Label: "Backlog", Stacked: true, Diff: false},
				{Name: "running", Label: "Running", Stacked: true, Diff: false},
			},
		},
	}
}

// FetchMetrics interface for mackerelplugin
func (m PumaPlugin) FetchMetrics() (map[string]interface{}, error) {
	state := PumaState{}

	data, err := ioutil.ReadFile(m.StateFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal([]byte(data), &state)
	if err != nil {
		return nil, err
	}

	transport := http.Transport{
		Dial: func(proto, addr string) (conn net.Conn, err error) {
			controlURL := strings.Replace(state.ControlURL, "unix://", "", 1)
			return net.Dial("unix", controlURL)
		},
	}
	client := &http.Client{Transport: &transport}
	resp, err := client.Get("http://puma/stats?token=" + state.ControlAuthToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return m.parseStats(resp.Body)
}

func (m PumaPlugin) parseStats(body io.Reader) (map[string]interface{}, error) {
	stat := make(map[string]interface{})
	decoder := json.NewDecoder(body)

	s := PumaStatus{}
	err := decoder.Decode(&s)
	if err != nil {
		return nil, err
	}

	stat["phase"] = s.Phase
	stat["total"] = s.Workers
	stat["booted"] = s.BootedWorkers
	stat["old"] = s.OldWorkers

	for _, w := range s.WorkerStatus {
		index := strconv.FormatUint(uint64(w.Index), 10)
		stat[m.Prefix+".worker_status."+index+".running"] = w.LastStatus.Running
		stat[m.Prefix+".worker_status."+index+".backlog"] = w.LastStatus.Backlog
	}

	return stat, nil
}

// Do the plugin
func Do() {
	optSocket := flag.String("state", "", "State file")
	optToken := flag.String("token", "", "Control token")
	optPrefix := flag.String("metric-key-prefix", "puma", "Metric key prefix")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	pumasrv := PumaPlugin{
		StateFile: *optSocket,
		Token:     *optToken,
		Prefix:    *optPrefix,
	}

	helper := mp.NewMackerelPlugin(pumasrv)
	helper.Tempfile = *optTempfile

	helper.Run()
}
