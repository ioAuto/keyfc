package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"github.com/wcharczuk/go-chart"
)

type Record struct {
	Date      time.Time
	Yesterday int
}

func GetYesterday() (int, error) {
	resp, err := http.Get("http://www.keyfc.net/bbs/index.aspx")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("wrong status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, err
	}

	selection := doc.Find("#wrap > div > div.announcement.s_clear > span > em")
	return strconv.Atoi(selection.Get(1).FirstChild.Data)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func timeSeries(records []*Record) chart.TimeSeries {
	ts := chart.TimeSeries{}
	for i := range records {
		ts.XValues = append(ts.XValues, records[i].Date)
		ts.YValues = append(ts.YValues, float64(records[i].Yesterday))
	}
	return ts
}

func DrawChart(path string,records []*Record) error {
	graph := chart.Chart{
		Series: []chart.Series{
			timeSeries(records),
		},
	}
	buf := new(bytes.Buffer)
	err := graph.Render(chart.PNG, buf)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, buf.Bytes(), 0644)
}

func main() {
	path := flag.String("p","record.json","json path")
	chartPath := flag.String("chart","chart.png","chart path")
	flag.Parse()

	records := &[]*Record{}
	if fileExists(*path) {
		b, err := ioutil.ReadFile(*path)
		if err != nil {
			logrus.Fatal(err)
		}
		err = json.Unmarshal(b, records)
		if err != nil {
			logrus.Fatal(err)
		}
	}
	num, err := GetYesterday()
	if err != nil {
		logrus.Fatal(err)
	}
	*records = append(*records,&Record{
		Date:time.Now(),
		Yesterday:num,
	})
	b, err := json.Marshal(records)
	if err != nil {
		logrus.Fatal(err)
	}
	err = ioutil.WriteFile(*path,b,0644)
	if err != nil {
		logrus.Fatal(err)
	}

	if len(*records) >1 {
		if err := DrawChart(*chartPath, *records); err != nil {
			logrus.Fatal(err)
		}
	}
}
