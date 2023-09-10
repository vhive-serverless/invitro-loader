package driver

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"strconv"

	"github.com/eth-easl/loader/pkg/common"
	"github.com/eth-easl/loader/pkg/config"
	"github.com/eth-easl/loader/pkg/generator"
	mc "github.com/eth-easl/loader/pkg/metric"
	"github.com/eth-easl/loader/pkg/trace"
	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
)

type DriverConfiguration struct {
	LoaderConfiguration *config.LoaderConfiguration
	IATDistribution     common.IatDistribution
	TraceGranularity    common.TraceGranularity
	TraceDuration       int // in minutes

	YAMLPath string
	TestMode bool

	Functions       []*common.Function
	PromptFunctions []*common.Function
}

type Driver struct {
	Configuration          *DriverConfiguration
	SpecificationGenerator *generator.SpecificationGenerator
}

func NewDriver(driverConfig *DriverConfiguration) *Driver {
	return &Driver{
		Configuration:          driverConfig,
		SpecificationGenerator: generator.NewSpecificationGenerator(driverConfig.LoaderConfiguration.Seed),
	}
}

func (c *DriverConfiguration) WithWarmup() bool {
	if c.LoaderConfiguration.WarmupDuration > 0 {
		return true
	} else {
		return false
	}
}

// ///////////////////////////////////////
// HELPER METHODS
// ///////////////////////////////////////
func (d *Driver) outputFilename(name string) string {
	return fmt.Sprintf("%s_%s_%d_ClientTraining_%s.csv", d.Configuration.LoaderConfiguration.OutputPathPrefix, name, d.Configuration.TraceDuration, d.Configuration.LoaderConfiguration.ClientTraining)
}

func (d *Driver) runCSVWriter(records chan interface{}, filename string, writerDone *sync.WaitGroup) {
	log.Debugf("Starting writer for %s", filename)

	file, err := os.Create(filename)
	common.Check(err)
	defer file.Close()

	writer := gocsv.NewSafeCSVWriter(csv.NewWriter(file))
	if err := gocsv.MarshalChan(records, writer); err != nil {
		log.Fatal(err)
	}

	writerDone.Done()
}

func (d *Driver) runJobLogWriter(records chan interface{}, filename string, writerDone *sync.WaitGroup) {
	log.Debugf("Starting Job Log writer for %s", filename)

	file, err := os.Create(filename)
	common.Check(err)
	defer file.Close()
	for record := range records {
		byteArr, err := json.Marshal(record)
		common.Check(err)

		_, err = file.Write(byteArr)
		common.Check(err)

		_, err = file.WriteString("\n")
		common.Check(err)
	}
	writerDone.Done()
}

/////////////////////////////////////////
// METRICS SCRAPPERS
/////////////////////////////////////////

func (d *Driver) CreateMetricsScrapper(interval time.Duration,
	signalReady *sync.WaitGroup, finishCh chan int, allRecordsWritten *sync.WaitGroup) func() {
	timer := time.NewTicker(interval)

	return func() {
		signalReady.Done()
		clusterUsageRecords := make(chan interface{}, 100)
		knStatRecords := make(chan interface{}, 100)
		writerDone := sync.WaitGroup{}

		clusterUsageFile, err := os.Create(d.outputFilename("cluster_usage"))
		common.Check(err)
		defer clusterUsageFile.Close()

		writerDone.Add(1)
		go d.runCSVWriter(knStatRecords, d.outputFilename("kn_stats"), &writerDone)

		for {
			select {
			case <-timer.C:
				recCluster := mc.ScrapeClusterUsage()
				recCluster.Timestamp = time.Now().UnixMicro()

				byteArr, err := json.Marshal(recCluster)
				common.Check(err)

				_, err = clusterUsageFile.Write(byteArr)
				common.Check(err)

				_, err = clusterUsageFile.WriteString("\n")
				common.Check(err)

				recKnative := mc.ScrapeKnStats()
				recKnative.Timestamp = time.Now().UnixMicro()
				knStatRecords <- recKnative
			case <-finishCh:
				close(clusterUsageRecords)
				close(knStatRecords)

				writerDone.Wait()
				allRecordsWritten.Done()

				return
			}
		}
	}
}

/////////////////////////////////////////
// DRIVER LOGIC
/////////////////////////////////////////

type InvocationMetadata struct {
	Function              *common.Function
	RuntimeSpecifications *common.RuntimeSpecification
	Phase                 common.ExperimentPhase

	MinuteIndex     int
	InvocationIndex int

	SuccessCount        *int64
	FailedCount         *int64
	FailedCountByMinute []int64

	RecordOutputChannel    chan *mc.ExecutionRecord
	JobRecordOutputChannel chan *mc.JobExecutionRecord
	JobSchedOutputChannel  chan *mc.JobSchedRequest
	JobSchedInputChannel   chan *mc.JobSchedReply
	AnnounceDoneWG         *sync.WaitGroup
}

func composeInvocationID(timeGranularity common.TraceGranularity, minuteIndex int, invocationIndex int) string {
	var timePrefix string

	switch timeGranularity {
	case common.MinuteGranularity:
		timePrefix = "min"
	case common.SecondGranularity:
		timePrefix = "sec"
	default:
		log.Fatal("Invalid trace granularity parameter.")
	}

	return fmt.Sprintf("%s%d.inv%d", timePrefix, minuteIndex, invocationIndex)
}

func extractModelName(functionName string) string {

	if strings.Contains(functionName, "llama-7b") {
		return "llama-7b-"
	}
	if strings.Contains(functionName, "llama-13b") {
		return "llama-13b-"
	}
	if strings.Contains(functionName, "gpt2-base") {
		return "gpt2-base-"
	}
	if strings.Contains(functionName, "gpt2-large") {
		return "gpt2-large-"
	}
	return "test"
}
func (d *Driver) invokeFunction(metadata *InvocationMetadata, functions []*common.Function, promptFunctions []*common.Function) {
	defer metadata.AnnounceDoneWG.Done()
	invocationID := extractModelName(metadata.Function.Name) + composeInvocationID(d.Configuration.TraceGranularity, metadata.MinuteIndex, metadata.InvocationIndex)
	success, record, jobRecord := Invoke(metadata.Function, functions, promptFunctions, metadata.RuntimeSpecifications, d.Configuration.LoaderConfiguration, invocationID, metadata.JobSchedOutputChannel, metadata.JobSchedInputChannel)

	record.Phase = int(metadata.Phase)
	record.InvocationID = extractModelName(metadata.Function.Name) + composeInvocationID(d.Configuration.TraceGranularity, metadata.MinuteIndex, metadata.InvocationIndex)

	if success {
		atomic.AddInt64(metadata.SuccessCount, 1)
	} else {
		atomic.AddInt64(metadata.FailedCount, 1)
		atomic.AddInt64(&metadata.FailedCountByMinute[metadata.MinuteIndex], 1)
	}

	metadata.RecordOutputChannel <- record
	metadata.JobRecordOutputChannel <- jobRecord
}

func (d *Driver) individualFunctionDriver(function *common.Function, functions []*common.Function, promptFunctions []*common.Function,
	announceFunctionDone *sync.WaitGroup, totalSuccessful *int64, totalFailed *int64, totalIssued *int64,
	recordOutputChannel chan *mc.ExecutionRecord, jobRecordOutputChannel chan *mc.JobExecutionRecord, jobSchedRequest chan *mc.JobSchedRequest, jobSchedReply chan *mc.JobSchedReply) {
	totalTraceDuration := d.Configuration.TraceDuration
	minuteIndex, invocationIndex := 0, 0

	IAT, runtimeSpecification := function.Specification.IAT, function.Specification.RuntimeSpecification

	var successfulInvocations int64
	var failedInvocations int64
	var failedInvocationByMinute = make([]int64, totalTraceDuration)
	var numberOfIssuedInvocations int64
	var currentPhase = common.ExecutionPhase

	waitForInvocations := sync.WaitGroup{}

	if d.Configuration.WithWarmup() {
		currentPhase = common.WarmupPhase
		// skip the first minute because of profiling
		minuteIndex = 1

		log.Infof("Warmup phase has started.")
	}

	startOfMinute := time.Now()
	var previousIATSum int64
	gpuCount := -1
	if IsStringInList(d.Configuration.LoaderConfiguration.ClientTraining, []string{common.Multi, common.HiveD, common.INFless, common.Elastic}) {
		// IsStringInList(d.Configuration.LoaderConfiguration.ClientTraining); d.Configuration.LoaderConfiguration.ClientTraining == common.Single || d.Configuration.LoaderConfiguration.ClientTraining == common.HiveD {
		parts := strings.Split(function.Name, "-")
		gpuCount, _ = strconv.Atoi(parts[len(parts)-1])
	} else if IsStringInList(d.Configuration.LoaderConfiguration.ClientTraining,
		[]string{common.Caerus, common.BatchPriority, common.PipelineBatchPriority, common.Knative}) {

	} else if IsStringInList(d.Configuration.LoaderConfiguration.ClientTraining, []string{common.ElasticFlow}) {

	} else {
		log.Errorf("Invalid client_training value: %s", d.Configuration.LoaderConfiguration.ClientTraining)
	}

	for {
		if minuteIndex >= totalTraceDuration {
			// Check whether the end of trace has been reached
			break
		} else if function.InvocationStats.Invocations[minuteIndex] == 0 {
			// Sleep for a minute if there are no invocations
			if d.proceedToNextMinute(function, &minuteIndex, &invocationIndex,
				&startOfMinute, true, &currentPhase, failedInvocationByMinute, &previousIATSum) {
				break
			}

			switch d.Configuration.TraceGranularity {
			case common.MinuteGranularity:
				time.Sleep(time.Minute)
			case common.SecondGranularity:
				time.Sleep(time.Second)
			default:
				log.Fatal("Unsupported trace granularity.")
			}

			continue
		}
		numberOfIssuedInvocations++

		invokeFunctionOrNot := true

		if IsStringInList(d.Configuration.LoaderConfiguration.ClientTraining, []string{common.Multi, common.HiveD, common.INFless, common.Elastic}) {
			// log.Infof("numberOfIssuedInvocations %d, length of invocation %d\n", numberOfIssuedInvocations, len(function.BatchStats.Invocations))
			expectedGPUCount := function.BatchStats.Invocations[numberOfIssuedInvocations-1] / common.BszPerDevice
			if gpuCount != expectedGPUCount {
				invokeFunctionOrNot = false
			}
			// log.Infof("d.Configuration.TestMode invokeFunctionOrNot: %v: expectedGPUCount %d, gpuCount %d", invokeFunctionOrNot, expectedGPUCount, gpuCount)
		} else if IsStringInList(d.Configuration.LoaderConfiguration.ClientTraining, []string{common.Caerus, common.BatchPriority, common.PipelineBatchPriority, common.Knative}) {

		} else if IsStringInList(d.Configuration.LoaderConfiguration.ClientTraining, []string{common.ElasticFlow}) {

		} else {
			log.Errorf("Invalid client_training value: %s", d.Configuration.LoaderConfiguration.ClientTraining)
		}

		if (!d.Configuration.TestMode) && invokeFunctionOrNot {
			waitForInvocations.Add(1)
			log.Infof("numberOfIssuedInvocations %v", numberOfIssuedInvocations)
			log.Infof("length %v", len(function.IterationStats.Invocations))
			runtimeSpecification[minuteIndex][invocationIndex].Stats = common.GPTStats{
				Iterations: function.IterationStats.Invocations[numberOfIssuedInvocations-1],
				BatchSize:  function.BatchStats.Invocations[numberOfIssuedInvocations-1],
				Deadline:   function.DeadlineStats.Invocations[numberOfIssuedInvocations-1],
			}
			invoked_function := function
			if IsStringInList(d.Configuration.LoaderConfiguration.ClientTraining, []string{common.ElasticFlow}) {
				invoked_function = functions[numberOfIssuedInvocations%common.ServerfulCopyReplicas]
			}
			go d.invokeFunction(&InvocationMetadata{
				Function:               invoked_function,
				RuntimeSpecifications:  &runtimeSpecification[minuteIndex][invocationIndex],
				Phase:                  currentPhase,
				MinuteIndex:            minuteIndex,
				InvocationIndex:        invocationIndex,
				SuccessCount:           &successfulInvocations,
				FailedCount:            &failedInvocations,
				FailedCountByMinute:    failedInvocationByMinute,
				JobRecordOutputChannel: jobRecordOutputChannel,
				RecordOutputChannel:    recordOutputChannel,
				AnnounceDoneWG:         &waitForInvocations,
				JobSchedOutputChannel:  jobSchedRequest,
				JobSchedInputChannel:   jobSchedReply,
			},
				functions, promptFunctions)
		} else {
			// To be used from within the Golang testing framework
			log.Debugf("Test mode invocation fired.\n")

			recordOutputChannel <- &mc.ExecutionRecord{
				Phase:        int(currentPhase),
				InvocationID: function.Name + composeInvocationID(d.Configuration.TraceGranularity, minuteIndex, invocationIndex),
				StartTime:    time.Now().UnixNano(),
			}
			jobRecordOutputChannel <- &mc.JobExecutionRecord{
				InvocationID:   function.Name + composeInvocationID(d.Configuration.TraceGranularity, minuteIndex, invocationIndex),
				StartTime:      make([]int64, 0),
				Replica:        make([]int, 0),
				GpuCount:       make([]int, 0),
				ComputeTime:    make([]int64, 0),
				ExecutionTime:  make([]int64, 0),
				StartIteration: make([]int, 0),
				EndIteration:   make([]int, 0),
				TotalIteration: make([]int, 0),
				BatchSize:      make([]int, 0),
			}
			successfulInvocations++
		}

		iat := time.Duration(IAT[minuteIndex][invocationIndex]) * time.Microsecond

		currentTime := time.Now()
		schedulingDelay := currentTime.Sub(startOfMinute).Microseconds() - previousIATSum
		sleepFor := iat.Microseconds() - schedulingDelay
		time.Sleep(time.Duration(sleepFor) * time.Microsecond)

		previousIATSum += iat.Microseconds()

		invocationIndex++
		if function.InvocationStats.Invocations[minuteIndex] == invocationIndex || hasMinuteExpired(startOfMinute) {
			readyToBreak := d.proceedToNextMinute(function, &minuteIndex, &invocationIndex, &startOfMinute,
				false, &currentPhase, failedInvocationByMinute, &previousIATSum)

			if readyToBreak {
				break
			}
		}
		if time.Now().Second()%10 == 0 {
			red := "\033[32m"
			reset := "\033[0m"
			message := fmt.Sprintf("\t numberOfIssuedInvocations %d, successfulInvocations %d", numberOfIssuedInvocations, successfulInvocations)
			log.Debugf(red + message + reset)
		}
	}

	waitForInvocations.Wait()

	log.Debugf("All the invocations for function %s have been completed.\n", function.Name)
	announceFunctionDone.Done()

	atomic.AddInt64(totalSuccessful, successfulInvocations)
	atomic.AddInt64(totalFailed, failedInvocations)
	atomic.AddInt64(totalIssued, numberOfIssuedInvocations)
}

func (d *Driver) proceedToNextMinute(function *common.Function, minuteIndex *int, invocationIndex *int, startOfMinute *time.Time,
	skipMinute bool, currentPhase *common.ExperimentPhase, failedInvocationByMinute []int64, previousIATSum *int64) bool {

	if d.Configuration.TraceGranularity == common.MinuteGranularity {
		if !isRequestTargetAchieved(function.InvocationStats.Invocations[*minuteIndex], *invocationIndex, common.RequestedVsIssued) {
			// Not fatal because we want to keep the measurements to be written to the output file
			log.Warnf("Relative difference between requested and issued number of invocations is greater than %.2f%%. Terminating function driver for %s!\n", common.RequestedVsIssuedTerminateThreshold*100, function.Name)

			return true
		}

		for i := 0; i <= *minuteIndex; i++ {
			notFailedCount := function.InvocationStats.Invocations[i] - int(atomic.LoadInt64(&failedInvocationByMinute[i]))
			if !isRequestTargetAchieved(function.InvocationStats.Invocations[i], notFailedCount, common.IssuedVsFailed) {
				// Not fatal because we want to keep the measurements to be written to the output file
				log.Warnf("Percentage of failed request is greater than %.2f%%. Terminating function driver for %s!\n", common.FailedTerminateThreshold*100, function.Name)

				return true
			}
		}
	}

	*minuteIndex++
	*invocationIndex = 0
	*previousIATSum = 0

	if d.Configuration.WithWarmup() && *minuteIndex == (d.Configuration.LoaderConfiguration.WarmupDuration+1) {
		*currentPhase = common.ExecutionPhase
		log.Infof("Warmup phase has finished. Starting the execution phase.")
	}

	if !skipMinute {
		*startOfMinute = time.Now()
	} else {
		switch d.Configuration.TraceGranularity {
		case common.MinuteGranularity:
			*startOfMinute = time.Now().Add(time.Minute)
		case common.SecondGranularity:
			*startOfMinute = time.Now().Add(time.Second)
		default:
			log.Fatal("Unsupported trace granularity.")
		}
	}

	return false
}

func isRequestTargetAchieved(ideal int, real int, assertType common.RuntimeAssertType) bool {
	if ideal == 0 {
		return true
	}

	ratio := float64(ideal-real) / float64(ideal)

	var warnBound float64
	var terminationBound float64
	var warnMessage string

	switch assertType {
	case common.RequestedVsIssued:
		warnBound = common.RequestedVsIssuedWarnThreshold
		terminationBound = common.RequestedVsIssuedTerminateThreshold
		warnMessage = fmt.Sprintf("Relative difference between requested and issued number of invocations has reached %.2f.", ratio)
	case common.IssuedVsFailed:
		warnBound = common.FailedWarnThreshold
		terminationBound = common.FailedTerminateThreshold
		warnMessage = fmt.Sprintf("Percentage of failed invocations within a minute has reached %.2f.", ratio)
	default:
		log.Fatal("Invalid type of assertion at runtime.")
	}

	if ratio < 0 || ratio > 1 {
		log.Fatalf("Invalid arguments provided to runtime assertion.\n")
	} else if ratio >= terminationBound {
		return false
	}

	if ratio >= warnBound && ratio < terminationBound {
		log.Warn(warnMessage)
	}

	return true
}

func hasMinuteExpired(t1 time.Time) bool {
	return time.Since(t1) > time.Minute
}

func (d *Driver) globalTimekeeper(totalTraceDuration int, signalReady *sync.WaitGroup) {
	ticker := time.NewTicker(time.Minute)
	globalTimeCounter := 0

	signalReady.Done()

	for {
		<-ticker.C

		log.Debugf("End of minute %d\n", globalTimeCounter)
		globalTimeCounter++
		if globalTimeCounter >= totalTraceDuration {
			break
		}

		log.Debugf("Start of minute %d\n", globalTimeCounter)
	}

	ticker.Stop()
}

func (d *Driver) createGlobalMetricsCollector(filename string, joblogfilename string, collector chan *mc.ExecutionRecord, joblogCollector chan *mc.JobExecutionRecord,
	signalReady *sync.WaitGroup, signalEverythingWritten *sync.WaitGroup, totalIssuedChannel chan int64) {

	// NOTE: totalNumberOfInvocations is initialized to MaxInt64 not to allow collector to complete before
	// the end signal is received on totalIssuedChannel, which deliver the total number of issued invocations.
	// This number is known once all the individual function drivers finish issuing invocations and
	// when all the invocations return
	var totalNumberOfInvocations int64 = math.MaxInt64
	var currentlyWritten int64
	var currentlyLogWritten int64

	file, err := os.Create(filename)
	common.Check(err)
	defer file.Close()

	joblogfile, joblogerror := os.Create(joblogfilename)
	common.Check(joblogerror)
	defer joblogfile.Close()

	signalReady.Done()

	records := make(chan interface{}, 100)
	jobrecords := make(chan interface{}, 100)
	writerDone := sync.WaitGroup{}
	writerDone.Add(1)
	go d.runCSVWriter(records, filename, &writerDone)

	writerDone.Add(1)
	go d.runJobLogWriter(jobrecords, joblogfilename, &writerDone)

	for {
		select {
		case record := <-collector:
			records <- record

			currentlyWritten++
		case record := <-totalIssuedChannel:
			totalNumberOfInvocations = record

		case record := <-joblogCollector:
			jobrecords <- record
			currentlyLogWritten++
		}
		if currentlyWritten == totalNumberOfInvocations && currentlyLogWritten == totalNumberOfInvocations {
			close(records)
			close(jobrecords)
			writerDone.Wait()
			(*signalEverythingWritten).Done()
			return
		}
	}
}

func (d *Driver) startBackgroundProcesses(allRecordsWritten *sync.WaitGroup) (*sync.WaitGroup, chan *mc.ExecutionRecord, chan *mc.JobExecutionRecord, chan int64, chan int) {
	auxiliaryProcessBarrier := &sync.WaitGroup{}

	finishCh := make(chan int, 1)

	if d.Configuration.LoaderConfiguration.EnableMetricsScrapping {
		auxiliaryProcessBarrier.Add(1)

		allRecordsWritten.Add(1)
		metricsScrapper := d.CreateMetricsScrapper(time.Second*time.Duration(d.Configuration.LoaderConfiguration.MetricScrapingPeriodSeconds), auxiliaryProcessBarrier, finishCh, allRecordsWritten)
		go metricsScrapper()
	}

	auxiliaryProcessBarrier.Add(2)

	globalMetricsCollector := make(chan *mc.ExecutionRecord)
	totalIssuedChannel := make(chan int64)
	joblogsMetricsCollector := make(chan *mc.JobExecutionRecord)

	go d.createGlobalMetricsCollector(d.outputFilename("duration"), d.outputFilename("joblogs"), globalMetricsCollector, joblogsMetricsCollector, auxiliaryProcessBarrier, allRecordsWritten, totalIssuedChannel)

	traceDurationInMinutes := d.Configuration.TraceDuration
	go d.globalTimekeeper(traceDurationInMinutes, auxiliaryProcessBarrier)

	return auxiliaryProcessBarrier, globalMetricsCollector, joblogsMetricsCollector, totalIssuedChannel, finishCh
}

func (d *Driver) internalRun(iatOnly bool, generated bool) {
	var successfulInvocations int64
	var failedInvocations int64
	var invocationsIssued int64

	allIndividualDriversCompleted := sync.WaitGroup{}
	allRecordsWritten := sync.WaitGroup{}
	allRecordsWritten.Add(1)

	backgroundProcessesInitializationBarrier, globalMetricsCollector, joblogMetricsCollector, totalIssuedChannel, scraperFinishCh := d.startBackgroundProcesses(&allRecordsWritten)

	var jobSchedRequest chan *mc.JobSchedRequest = nil
	var jobSchedReply chan *mc.JobSchedReply = nil
	if IsStringInList(d.Configuration.LoaderConfiguration.ClientTraining, []string{common.ElasticFlow, common.Elastic, common.INFless}) {
		jobSchedRequest, jobSchedReply = d.startSchedBackgroundProcesses(&allRecordsWritten)
	}
	if !iatOnly {
		log.Info("Generating IAT and runtime specifications for all the functions")
		for i, function := range d.Configuration.Functions {
			spec := d.SpecificationGenerator.GenerateInvocationData(
				function,
				d.Configuration.IATDistribution,
				d.Configuration.TraceGranularity,
			)

			d.Configuration.Functions[i].Specification = spec
		}
	}

	backgroundProcessesInitializationBarrier.Wait()

	if generated {
		for i := range d.Configuration.Functions {
			var spec common.FunctionSpecification

			iatFile, _ := os.ReadFile("iat" + strconv.Itoa(i) + ".json")
			err := json.Unmarshal(iatFile, &spec)
			if err != nil {
				log.Fatalf("Failed tu unmarshal iat file: %s", err)
			}

			d.Configuration.Functions[i].Specification = &spec
		}
	}

	log.Infof("Starting function invocation driver\n")
	if IsStringInList(d.Configuration.LoaderConfiguration.ClientTraining, []string{common.ElasticFlow}) {
		for func_idx, function := range d.Configuration.Functions {
			if func_idx%common.ServerfulCopyReplicas != 0 {
				continue
			}
			allIndividualDriversCompleted.Add(1)
			fmt.Printf("invoke function %v, length of prompt functions %v\n", function, len(d.Configuration.PromptFunctions))
			key_prefix := strings.Split(function.Name, "-serverful-copy-")[0]
			filter_functions := FilterByKey(d.Configuration.Functions, key_prefix)
			go d.individualFunctionDriver(
				function,
				filter_functions,
				d.Configuration.PromptFunctions,
				&allIndividualDriversCompleted,
				&successfulInvocations,
				&failedInvocations,
				&invocationsIssued,
				globalMetricsCollector,
				joblogMetricsCollector,
				jobSchedRequest,
				jobSchedReply,
			)
		}
	} else {
		for _, function := range d.Configuration.Functions {
			allIndividualDriversCompleted.Add(1)
			fmt.Printf("invoke function %v, length of prompt functions %v\n", function, len(d.Configuration.PromptFunctions))
			if IsStringInList(d.Configuration.LoaderConfiguration.ClientTraining, []string{common.Caerus, common.BatchPriority, common.PipelineBatchPriority, common.Knative, common.Multi}) {
				go d.individualFunctionDriver(
					function,
					d.Configuration.Functions,
					d.Configuration.PromptFunctions,
					&allIndividualDriversCompleted,
					&successfulInvocations,
					&failedInvocations,
					&invocationsIssued,
					globalMetricsCollector,
					joblogMetricsCollector,
					jobSchedRequest,
					jobSchedReply,
				)
			} else if IsStringInList(d.Configuration.LoaderConfiguration.ClientTraining, []string{common.HiveD, common.INFless, common.Elastic}) {
				key_prefix := strings.Split(function.Name, "-gpu-")[0]
				filter_functions := FilterByKey(d.Configuration.Functions, key_prefix)
				go d.individualFunctionDriver(
					function,
					filter_functions,
					d.Configuration.PromptFunctions,
					&allIndividualDriversCompleted,
					&successfulInvocations,
					&failedInvocations,
					&invocationsIssued,
					globalMetricsCollector,
					joblogMetricsCollector,
					jobSchedRequest,
					jobSchedReply,
				)
			} else {
				log.Errorf("Invalid client_training value: %s", d.Configuration.LoaderConfiguration.ClientTraining)
			}

		}
	}

	allIndividualDriversCompleted.Wait()
	if atomic.LoadInt64(&successfulInvocations)+atomic.LoadInt64(&failedInvocations) != 0 {
		log.Debugf("Waiting for all the invocations record to be written.\n")

		totalIssuedChannel <- atomic.LoadInt64(&invocationsIssued)
		scraperFinishCh <- 0 // Ask the scraper to finish metrics collection
		SetFinish()
		allRecordsWritten.Wait()
	}

	log.Infof("Trace has finished executing function invocation driver\n")
	log.Infof("Number of successful invocations: \t%d\n", atomic.LoadInt64(&successfulInvocations))
	log.Infof("Number of failed invocations: \t%d\n", atomic.LoadInt64(&failedInvocations))
}

func (d *Driver) RunExperiment(iatOnly bool, generated bool) {
	if iatOnly {
		log.Info("Generating IAT and runtime specifications for all the functions")
		for i, function := range d.Configuration.Functions {
			spec := d.SpecificationGenerator.GenerateInvocationData(
				function,
				d.Configuration.IATDistribution,
				d.Configuration.TraceGranularity,
			)
			d.Configuration.Functions[i].Specification = spec

			file, _ := json.MarshalIndent(spec, "", " ")
			err := os.WriteFile("iat"+strconv.Itoa(i)+".json", file, 0644)
			if err != nil {
				log.Fatalf("Writing the loader config file failed: %s", err)
			}
		}

		return
	}

	if d.Configuration.WithWarmup() {
		trace.DoStaticTraceProfiling(d.Configuration.Functions)
	}
	if strings.Contains(d.Configuration.LoaderConfiguration.TracePath, "gpt") ||
		strings.Contains(d.Configuration.LoaderConfiguration.TracePath, "gpu") {
		trace.ApplyResourceLimitsForGPU(d.Configuration.Functions)
	} else {
		trace.ApplyResourceLimits(d.Configuration.Functions)
	}

	DeployFunctions(
		d.Configuration.LoaderConfiguration,
		d.Configuration.Functions,
		d.Configuration.YAMLPath,
		d.Configuration.LoaderConfiguration.IsPartiallyPanic,
		d.Configuration.LoaderConfiguration.EndpointPort,
		d.Configuration.LoaderConfiguration.AutoscalingMetric)

	if d.Configuration.LoaderConfiguration.WithPromptBank {
		d.Configuration.PromptFunctions = DeployPromptFunctions(
			d.Configuration.LoaderConfiguration,
			d.Configuration.Functions,
			d.Configuration.LoaderConfiguration.PromptYamlPath, // prompt path
			d.Configuration.LoaderConfiguration.IsPartiallyPanic,
			d.Configuration.LoaderConfiguration.EndpointPort,
			d.Configuration.LoaderConfiguration.AutoscalingMetric)
	}
	d.internalRun(iatOnly, generated)
}
