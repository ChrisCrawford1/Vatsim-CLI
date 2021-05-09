package services

import (
	"MyVatsimCLI/data"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

const URL = "https://data.vatsim.net/v3/vatsim-data.json"

func FetchCurrentData() data.Datafile {
	var dataInBytes []byte = callVatsim()

	var retrievedData data.Datafile

	if err := json.Unmarshal(dataInBytes, &retrievedData); err != nil {
		panic(err)
	}

	return retrievedData
}

func callVatsim() []byte {
	resp, err := http.Get(URL)

	if err != nil {
		panic(err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	dataInByteForm, byteErr := ioutil.ReadAll(resp.Body)

	if byteErr != nil {
		panic(err)
	}

	return dataInByteForm
}
