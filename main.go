package main

import (
	"bufio"
	"container/heap"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/santosh/gingo/docs"
	log "github.com/sirupsen/logrus"
)

var (
	domainsAdded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "domains_added",
		Help: "The total number of added domains",
	})
	queryResponseTimes = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "syringe",
			Name:      "query_response_time",
			Help:      "query_response_time",
			Buckets:   []float64{0.1, 0.2, 0.5, 1.0, 1.5, 2, 5},
		})
	queryResponseTtl = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "syringe",
			Name:      "query_response_ttl",
			Help:      "query_response_ttl",
			Buckets:   []float64{0, 5, 10, 30, 60, 300, 600, 900, 1800, 3600, 86400},
		})
	queueSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "syringe",
		Name:      "queue_size",
		Help:      "The current size of the queue",
	})
	// heapPushChan - push channel for pushing to a heap
	heapPushChan = make(chan heapPushChanMsg)
	// heapPopChan - pop channel for popping from a heap
	heapPopChan = make(chan heapPopChanMsg)
	domainList  []string
)

var resolverConfiguration *ResolverConfiguration = &ResolverConfiguration{}
var ginInstance *gin.Engine
var dh *DomainHeap
var resolverStrategies *ResolverStrategies

func init() {
	// Initialize configuration
	err := ParseFlags(resolverConfiguration)
	if err != nil {
		log.Fatal(err)
	}

	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	//log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.Level(resolverConfiguration.LogLevel))
	// Prometheus
	prometheus.Register(domainsAdded)
	prometheus.Register(queryResponseTimes)
	prometheus.Register(queryResponseTtl)
	prometheus.Register(queueSize)

	// Initialize router/variables
	if resolverConfiguration.LogLevel >= uint(log.DebugLevel) {
		log.Debug("Setting gin mode to DebugMode because LogLevel is set to ", resolverConfiguration.LogLevel, ". Using LogLevel<", uint(log.DebugLevel), " will set the mode to release")
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	ginInstance = SetupRouter()
	dh = &DomainHeap{}
	heap.Init(dh)
	// Start the queue serializer (will schedule heap access)
	dh.watchHeapOps()

	// Initialize strategies
	resolverStrategies = &ResolverStrategies{
		ResolveFunctions: []func(config *ResolverConfiguration, domain *Domain) (uint, error){
			TryQueryRegularDomain,
			TryQuerySOADomain,
			TryQueryFlexibleDelayDomain},
	}
}

// @title           Syringe Api Documentation
// @version         1.0
// @description     A lightweight api for the syringe daemon
// @termsOfService  https://github.com/TCMPK/syringe

// @contact.name   Peter Klein
// @contact.url    https://blog.tcmpk.de
// @contact.email  github@tcmpk.de

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8000
// @BasePath  /api/v1
func main() {
	// Register api endpoints
	ginInstance.GET("/docs/*any", DocOverrideHandler)
	ginInstance.StaticFile("/swagger-static/doc.json", "docs/swagger.json")
	ginInstance.GET("/metrics", gin.WrapH(promhttp.Handler()))
	// Run the webserver
	go func() {
		ginInstance.Run(":" + strconv.FormatUint(uint64(resolverConfiguration.ServerListenPort), 10))
	}()

	if resolverConfiguration.LoadDomainsFileOnStart {
		log.Debug("Loading domains from ", resolverConfiguration.DomainsFile)
		rowsRead, err := dh.LoadDomainsFile()
		if err != nil {
			log.Fatal(err)
		}
		log.Info("Read ", rowsRead, " rows from DomainsFile ", resolverConfiguration.DomainsFile)
	}

	// Main loop
	for {
		queueSize.Set(float64(dh.Len()))
		if dh.Len() == 0 {
			time.Sleep(time.Duration(resolverConfiguration.SleepLowTresholdMilliseconds) * time.Millisecond)
			continue
		}

		var cur Domain = HeapPop(dh).(Domain)
		sleep_time := cur.MillisUntilDue()
		if sleep_time > int64(resolverConfiguration.SleepLowTresholdMilliseconds) {
			HeapPush(dh, cur)
			time.Sleep(time.Duration(resolverConfiguration.SleepLowTresholdCheckIntervalMilliseconds) * time.Millisecond)
			continue
		}

		if sleep_time >= 0 && sleep_time <= int64(resolverConfiguration.SleepLowTresholdCheckIntervalMilliseconds) {
			time.Sleep(time.Duration(sleep_time) * time.Millisecond)
		}

		ch := make(chan uint, 1)
		go func() {
			start := time.Now()
			go cur.Query(resolverStrategies, resolverConfiguration, ch)
			var ttl = <-ch
			cur.RefreshInSeconds(ttl)
			HeapPush(dh, cur)
			elapsed := time.Since(start)
			queryResponseTimes.Observe(elapsed.Seconds())
			queryResponseTtl.Observe(float64(ttl))
		}()
	}
}

func ReadDomainsFile(f string) ([]string, error) {
	var domainsRead = []string{}
	readFile, err := os.Open(resolverConfiguration.DomainsFile)
	if err != nil {
		return domainsRead, err
	}
	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)
	i := 0
	for fileScanner.Scan() {
		i++
		line_split := strings.Split(fileScanner.Text(), " ")
		if len(line_split) != 2 {
			log.Error("Malformed line ", i, " in file ", resolverConfiguration.DomainsFile, " syntax='<domain> <rr type>' each per line - example: 'google.de A'")
			continue
		}
		domainsRead = append(domainsRead, fileScanner.Text())
	}

	readFile.Close()
	return domainsRead, nil
}

func (dh *DomainHeap) LoadDomainsFile() (uint, error) {
	var i uint = 0
	domainsRead, err := ReadDomainsFile(resolverConfiguration.DomainsFile)
	if err != nil {
		return i, err
	}

	for _, d := range domainsRead {
		if !slices.Contains(domainList, d) {
			i++
			dh.AddDomain(DomainListEntryToDomain(d))
		}
	}
	return i, nil
}

func (dh *DomainHeap) AppendRandom(delay_seconds uint) {
	ReadDomainsFile(resolverConfiguration.DomainsFile)
	if len(domainList) == 0 {
		return
	}
	domain_index := rand.Intn(len(domainList) - 1)
	domain := DomainListEntryToDomain(domainList[domain_index])
	domain.RefreshInSeconds(delay_seconds)
	dh.AddDomain(domain)
}

func DomainListEntryToDomain(d string) Domain {
	line_split := strings.Split(d, " ")

	domain_name := line_split[0]
	rr_type := line_split[1]

	return Domain{Record_name: domain_name, Record_type: rr_type, Refresh_at: 0}
}

func LoadDomainsBulkWithApproxRateLimit(qps int, d []Domain) {
	dSize := len(d)
	bulk_slot_size := dSize / qps
	if bulk_slot_size == 0 {
		bulk_slot_size = 1
	}
	for i := 0; i < dSize; i++ {
		d[i].RefreshInSeconds(uint(i % dSize))
		dh.AddDomain(d[i])
	}
}
