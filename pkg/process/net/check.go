// +build linux

package net

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/ebpf/probe"
)

const (
	checksURL = "http://unix/check"
)

// GetCheck returns the output of the specified check
func (r *RemoteSysProbeUtil) GetCheck(check string) (interface{}, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", checksURL, check), nil)
	if err != nil {
		return nil, err
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("conn request failed: socket %s, url %s, status code: %d", r.path, fmt.Sprintf("%s/%s", checksURL, check), resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if check == "tcp_queue_length" {
		var stats probe.TCPQueueLengthStats
		err = json.Unmarshal(body, &stats)
		if err != nil {
			return nil, err
		}
		return stats, nil
	} else if check == "oom_kill" {
		var stats []probe.OOMKillStats
		err = json.Unmarshal(body, &stats)
		if err != nil {
			return nil, err
		}
		return stats, nil
	}

	return nil, fmt.Errorf("Invalid check name: %s", check)
}
