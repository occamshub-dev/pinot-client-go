package pinot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// jsonAsyncHTTPClientTransport is the impl of clientTransport
type jsonAsyncHTTPClientTransport struct {
	client http.Client
	header map[string]string
}

func (t jsonAsyncHTTPClientTransport) execute(brokerAddress string, query *Request) (*BrokerResponse, error) {
	var url string
	if query.queryFormat == "sql" {
		url = fmt.Sprintf("http://%s/query/sql", brokerAddress)
	} else {
		url = fmt.Sprintf("http://%s/query", brokerAddress)
	}
	requestJSON := map[string]string{}
	requestJSON[query.queryFormat] = query.query
	if query.queryFormat == "sql" {
		requestJSON["queryOptions"] = "groupByMode=sql;responseFormat=sql"
	}
	jsonValue, _ := json.Marshal(requestJSON)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Error("Invalid HTTP Request", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := t.client.Do(req)
	if err != nil {
		log.Error("Got exceptions during sending request", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error("Unable to read Pinot response", err)
		}
		var brokerResponse BrokerResponse
		err = json.Unmarshal(bodyBytes, &brokerResponse)
		if err != nil {
			log.Error("Unable to unmarshal json response to a brokerResponse structure.", err)
			return nil, err
		}
		return &brokerResponse, nil
	}
	return nil, fmt.Errorf("Caught http exception when querying Pinot: %v", resp.Status)
}
