//   Copyright 2016 DigitalOcean
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package collectors

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/valyala/fasthttp"
)

const (
	pushNamespace = "push"
)

// PoolUsageCollector displays statistics about each pool we have created
// in the ceph cluster.
type PushInfoCollector struct {

	// UsedBytes tracks the amount of bytes currently allocated for the pool. This
	// does not factor in the overcommitment made for individual images.
	Online *prometheus.GaugeVec

	// RawUsedBytes tracks the amount of raw bytes currently used for the pool. This
	// factors in the replication factor (size) of the pool.
	UserAmount *prometheus.GaugeVec

	// MaxAvail tracks the amount of bytes currently free for the pool,
	// which depends on the replication settings for the pool in question.
	MSGID *prometheus.GaugeVec
}

// NewPoolUsageCollector creates a new instance of PoolUsageCollector and returns
// its reference.
func NewPushInfoCollector() *PushInfoCollector {
	var (
		//		subSystem = "push"
		pushLabel = []string{"productid"}
	)

	//labels := make(prometheus.Labels)
	//	labels["cluster"] = cluster

	return &PushInfoCollector{

		Online: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: pushNamespace,
				Name:      "online",
				Help:      "push online",
				//			ConstLabels: labels,
			},
			[]string{"productid", "appname"},
		),
		UserAmount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: pushNamespace,
				Name:      "user_amount",
				Help:      "push user amount",
				//				ConstLabels: labels,
			},
			pushLabel,
		),
		MSGID: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: pushNamespace,
				Name:      "msgid",
				Help:      "push msgid",
				//		ConstLabels: labels,
			},
			pushLabel,
		),
	}
}

func (p *PushInfoCollector) collectorList() []prometheus.Collector {
	return []prometheus.Collector{
		p.Online,
		p.UserAmount,
		p.MSGID,
	}
}

type Pushinfo struct {
	Data []struct {
		Amount  int    `json:"amount"`
		ID      string `json:"productId"`
		Appname string `json:"appname"`
	} `json:"data"`
}
type MsgIDinfo struct {
	Data []struct {
		Amount int    `json:"amount"`
		ID     string `json:"ID"`
	} `json:"data"`
}

func (p *PushInfoCollector) collect() error {
	/*
		http://10.143.184.153:8090/online.php
		http://10.143.184.153:8090/user_amount.php
		http://10.143.184.153:8090/msgid.php
	*/
	//onlineBody := doRequest("http://10.143.184.153:8090/online.php")

	var uadata Pushinfo
	userAmountBody := doRequest("http://10.143.184.153:8090/user_amount.php")
	if err := json.Unmarshal(userAmountBody, &uadata); err != nil {
		fmt.Println(err)
	}

	for _, ua := range uadata.Data {
		p.UserAmount.WithLabelValues(ua.ID).Set(float64(ua.Amount))
	}

	var msgdata MsgIDinfo
	msgIDBody := doRequest("http://10.143.184.153:8090/msgid.php")
	if err := json.Unmarshal(msgIDBody, &msgdata); err != nil {
		fmt.Println(err)
	}
	for _, ua := range msgdata.Data {
		p.MSGID.WithLabelValues(ua.ID).Set(float64(ua.Amount))
	}
	var online Pushinfo
	onlineBody := doRequest("http://10.143.184.153:8090/online.php")
	fmt.Println(string(onlineBody))
	if err := json.Unmarshal(onlineBody, &online); err != nil {
		fmt.Println(err)
	}
	for _, ua := range online.Data {

		p.Online.WithLabelValues(ua.ID, ua.Appname).Set(float64(ua.Amount))
	}
	return nil

}

// Describe fulfills the prometheus.Collector's interface and sends the descriptors
// of pool's metrics to the given channel.
func (p *PushInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range p.collectorList() {
		metric.Describe(ch)
	}
}

// Collect extracts the current values of all the metrics and sends them to the
// prometheus channel.
func (p *PushInfoCollector) Collect(ch chan<- prometheus.Metric) {
	if err := p.collect(); err != nil {
		log.Println("[ERROR] failed collecting pool usage metrics:", err)
		return
	}

	for _, metric := range p.collectorList() {
		metric.Collect(ch)
	}
}

func doRequest(url string) []byte {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)

	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	client.Do(req, resp)

	bodyBytes := resp.Body()
	return bodyBytes
	// User-Agent: fasthttp
	// Body:
}
