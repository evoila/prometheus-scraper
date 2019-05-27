package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/evoila/scrape-elasticsearch/config"
	"github.com/evoila/scrape-elasticsearch/elasticsearch"
	"github.com/evoila/scrape-elasticsearch/mongodb"
	"github.com/evoila/scrape-elasticsearch/prometheus"
	"github.com/hashicorp/logutils"
	dto "github.com/prometheus/client_model/go"
)

// Exit codes are int values that represent an exit code for a particular error.
const (
	ExitCodeOK    int = 0
	ExitCodeError int = 1 + iota
)

const (
	// DefaultCfgPath is default config file path
	DefaultCfgPath = "example/scrape-elasticsearch.toml"

	// DefaultStatsInterval is default interval of displaying
	// stats info to console
	DefaultStatsInterval = 10 * time.Second

	// DefaultUsername to grant access token for firehose
	DefaultUsername = "admin"

	// DefaultIdleTimeout is the default timeout for receiving a single message
	// from the firehose
	DefaultIdleTimeout = 60 * time.Second
)

// godocFile is file name for godoc
const (
	godocFile = "doc.go"
)

// CLI is the command line object
type CLI struct {
	// outStream and errStream are the stdout and stderr
	// to write message from the CLI.
	OutStream, ErrStream io.Writer
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var (
		cfgPath  string
		username string
		password string
		logLevel string

		worker int

		debug    bool
		version  bool
		genGodoc bool
	)

	// Define option flag parsing
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(cli.ErrStream)
	flags.Usage = func() {
		fmt.Fprintf(cli.ErrStream, helpText)
	}

	flags.StringVar(&cfgPath, "config", DefaultCfgPath, "")
	flags.StringVar(&username, "username", "", "")
	flags.StringVar(&password, "password", "", "")
	flags.StringVar(&logLevel, "log-level", "INFO", "")
	flags.IntVar(&worker, "worker", runtime.NumCPU(), "")
	flags.BoolVar(&debug, "debug", false, "")
	flags.BoolVar(&version, "version", false, "")

	// -gen-godoc flag is only for developers of this nozzle.
	// It generates godoc.
	flags.BoolVar(&genGodoc, "gen-godoc", false, "")

	// Parse commandline flag
	if err := flags.Parse(args[1:]); err != nil {
		return ExitCodeError
	}

	// Generate godoc
	if genGodoc {
		if err := godoc(); err != nil {
			fmt.Fprintf(cli.ErrStream, "Faild to generate godoc %s\n", err)
			return ExitCodeError
		}

		fmt.Fprintf(cli.OutStream, "Successfully generated godoc\n")
		return ExitCodeOK
	}

	// Show version
	if version {
		fmt.Fprintf(cli.ErrStream, "%s version %s\n", Name, Version)
		return ExitCodeOK
	}

	// Setup logger with level Filtering
	logger := log.New(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "ERROR"},
		MinLevel: (logutils.LogLevel)(strings.ToUpper(logLevel)),
		Writer:   cli.OutStream,
	}, "", log.LstdFlags)
	logger.Printf("[INFO] LogLevel: %s", logLevel)

	// Show basic infomation
	logger.Printf("[INFO] %s version: %s", Name, Version)
	logger.Printf("[INFO] Go version: %s (%s/%s)",
		runtime.Version(), runtime.GOOS, runtime.GOARCH)
	logger.Printf("[INFO] Num of CPU: %d", runtime.NumCPU())

	// Load configuration
	config, err := config.LoadConfig(cfgPath)
	if err != nil {
		logger.Printf("[ERROR] Failed to load configuration file: %s", err)
		return ExitCodeError
	}
	logger.Printf("[DEBUG] %v", config)

	if config.ScrapeEndpoints == nil {
		logger.Printf("[ERROR] Could not find any scrape endpoint configuration")
	}

	ctx := context.Background()
	elasticsearchClient := elasticsearch.NewElasticsearchClient(config)
	mongoClient := mongodb.NewMongoDBClient(ctx, config)

	var closed = make(chan struct{})
	var wg sync.WaitGroup
	for _, scrapeEndpoint := range config.ScrapeEndpoints {
		serviceInstances := mongodb.LoadServiceInstances(ctx, mongoClient, scrapeEndpoint.Type)
		for _, serviceInstance := range serviceInstances {
			for _, host := range serviceInstance.Hosts {
				wg.Add(1)
				go func(name string, ip string, port int, logger *log.Logger) {
					logger.Printf("[INFO] Adding scraper for type %s", name)
					var ticker = time.NewTicker(time.Second * 5)
					wg.Done()
					for {
						select {
						case <-closed:
							logger.Printf("[INFO] Got signal. Aborting...")
							return
						case <-ticker.C:
							mfChan := make(chan *dto.MetricFamily, 1024)
							prometheus.FetchMetricFamilies("http://"+ip+":"+strconv.Itoa(port)+"/metrics", mfChan, "", "", true)

							result := []*prometheus.Family{}
							for mf := range mfChan {
								result = append(result, prometheus.NewFamily(mf))
							}
							jsonText, err := json.Marshal(result)
							if err != nil {
								log.Fatalln("error marshaling JSON:", err)
							}
							if _, err := os.Stdout.Write(jsonText); err != nil {
								log.Fatalln("error writing to stdout:", err)
							}

							//logger.Printf("[INFO]: Response result for type %v with content %s", name, string(data))

							go func(ip string, data string, logger *log.Logger) {
								elasticsearch.WriteMetric(ctx, elasticsearchClient)
							}(ip, string(jsonText), logger)

						}
					}
				}(scrapeEndpoint.Type+"-"+host.IP, host.IP, scrapeEndpoint.Port, logger)
			}
		}

	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	select {
	case sig := <-c:
		fmt.Printf("Got %s signal. Aborting...\n", sig)
		wg.Wait()
	}

	// Attempt to close all the things. Not returns soon even if
	// error is happend while closing.
	isError := false
	logger.Printf("[INFO] Finished scrape elasticsearch")
	if isError {
		return ExitCodeError
	}
	return ExitCodeOK
}

func godoc() error {
	f, err := os.Create(godocFile)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl, err := template.New("godoc").Parse(godocTmpl)
	if err != nil {
		return err
	}
	return tmpl.Execute(f, helpText)
}

var godocTmpl = `// THIS FILE IS GENERATED BY GO GENERATE.
// DO NOT EDIT THIS FILE BY HAND.

/*
{{ . }}
*/
package main
`

// helpText is used for flag usage messages.
var helpText = `scrape-elasticsearch is a HTTP Agent scraping Prometheus endpoints, which
forwards the elements to Elasticsearch.

Usage:

scrape-elasticsearch [options]

Available options:

    -config PATH          Path to configuraiton file    
    -username NAME        username to grant access token to connect firehose
    -password PASS        password to grant access token to connect firehose
    -worker NUM           Number of producer worker. Default is number of CPU core
    -stats-interval TIME  How often display stats info to console  
    -debug                Output event to stdout
    -log-level LEVEL      Log level. Default level is INFO (DEBUG|INFO|ERROR)
`
