package measure

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
	"github.com/ttacon/chalk"
)

type TestCase struct {
	id         int64
	query      string
	concurrent int64
}

type Measurement struct {
	ApiAddresses  []string
	InputFilePath string
	OutputPath    string

	testCases []TestCase
}

func NewMeasurement(apiAddresses []string, inputFilePath string, outputPath string) Measurement {
	return Measurement{
		ApiAddresses:  apiAddresses,
		InputFilePath: inputFilePath,
		OutputPath:    outputPath,
	}
}

func (inst *Measurement) Run() error {
	err := inst.readInputFile()
	if err != nil {
		return err
	}

	err = inst.measure()
	if err != nil {
		return err
	}
	return nil
}

func (inst *Measurement) readInputFile() error {
	file, err := os.Open(inst.InputFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for index, record := range records {
		concurrent, err := strconv.ParseInt(record[1], 10, 64)
		if err != nil {
			return err
		}
		inst.testCases = append(inst.testCases, TestCase{
			id:         int64(index),
			query:      record[0],
			concurrent: concurrent,
		})
	}
	return nil
}

func (inst *Measurement) writeResultFile(name string, data [][]string) error {
	csvFile, err := os.Create(fmt.Sprintf("%s/%s.csv", inst.OutputPath, name))
	if err != nil {
		return err
	}
	defer csvFile.Close()

	w := csv.NewWriter(csvFile)
	defer w.Flush()

	w.WriteAll(data)

	return nil
}

func (inst *Measurement) measure() error {
	for _, api := range inst.ApiAddresses {
		for _, testcase := range inst.testCases {
			err := inst.measureAnApi(api, testcase)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (inst *Measurement) measureAnApi(api string, testcase TestCase) error {
	var bodyResponses, latencies [][]string

	duration := 100 * time.Millisecond
	rate := vegeta.Rate{Freq: int(testcase.concurrent), Per: duration}
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    fmt.Sprintf("%s?%s", api, testcase.query),
	})
	attacker := vegeta.NewAttacker(
		vegeta.KeepAlive(false),
		vegeta.Workers(uint64(testcase.concurrent)),
	)

	var metrics vegeta.Metrics
	var idx = 1
	for res := range attacker.Attack(targeter, rate, duration, "API Benchmark!") {
		name := fmt.Sprintf("%s_%d", testcase.query, idx)
		metrics.Add(res)

		dst := &bytes.Buffer{}
		err := json.Compact(dst, res.Body)
		if err != nil {
			return err
		}
		bodyResponses = append(bodyResponses, []string{name, dst.String()})

		latencies = append(latencies, []string{name, strconv.FormatFloat(res.Latency.Seconds(), 'f', -1, 32)})

		idx++
	}
	metrics.Close()

	fmt.Println(chalk.Blue, fmt.Sprintf("\nResult for query: %s, concurrency: %d", testcase.query, testcase.concurrent))
	fmt.Printf("Success percentage: %v%% (%d requests total)\n", metrics.Success*100, metrics.Requests)
	fmt.Printf("Success request per second: %v requests\n", metrics.Throughput)
	fmt.Printf("Latency Min: %v\n", metrics.Latencies.Min)
	fmt.Printf("Latency Max: %v\n", metrics.Latencies.Max)
	fmt.Printf("Latency Mean: %v\n", metrics.Latencies.Mean)
	fmt.Printf("Latency P95: %v\n", metrics.Latencies.P95)
	fmt.Println("=========================================")

	err := inst.writeResultFile(fmt.Sprintf("%s.response", testcase.query), bodyResponses)
	if err != nil {
		return err
	}

	err = inst.writeResultFile(fmt.Sprintf("%s.latency", testcase.query), latencies)
	if err != nil {
		return err
	}

	return nil
}
