package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	url string = "https://api.fiscaldata.treasury.gov/services/api/fiscal_service/v2/accounting/od/debt_to_penny?sort=-record_date"
)

// DataPointCollection represents a collection of DataPoint objects.
type DataPointCollection struct {
	Items []*DataPoint `json:"data"`
}

type Sum float64

func (s *Sum) UnmarshalJSON(b []byte) error {
	var value string
	if err := json.Unmarshal(b, &value); err != nil {
		return err
	}
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}
	*s = Sum(number)
	return nil
}

// DataPoint represents a data point that is stored in the DataPointCollection.
type DataPoint struct {
	Sum Sum `json:"tot_pub_debt_out_amt"`
}

type context struct {
	Sum  string
	Time string
}

var output string = `
## United States National Debt

### ${{.Sum}} (-)

Last updated: {{.Time}}
`

func main() {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	if err := response.Body.Close(); err != nil {
		log.Fatal(err)
	}
	var collection DataPointCollection
	if err := json.Unmarshal(b, &collection); err != nil {
		log.Fatal(err)
	}
	p := message.NewPrinter(language.English)
	sum := p.Sprintf("%.2f", collection.Items[0].Sum)
	t := template.Must(template.New("output").Parse(output))
	descriptor, err := os.Create("README.md")
	if err != nil {
		log.Fatal(err)
	}
	context := &context{sum, time.Now().Format(time.RFC822)}
	if err := t.Execute(descriptor, context); err != nil {
		log.Fatal(err)
	}
	if err := descriptor.Close(); err != nil {
		log.Fatal(err)
	}
}
