// Package utilization implements the Utilization spec, available at
// https://source.datanerd.us/agents/agent-specs/blob/master/Utilization.md
//
package utilization

import (
	"net/http"
	"runtime"
	"sync"

	"github.com/newrelic/go-agent/internal/logger"
	"github.com/newrelic/go-agent/internal/sysinfo"
)

const (
	metadataVersion = 3
)

// Config controls the behavior of utilization information capture.
type Config struct {
	DetectAWS         bool
	DetectAzure       bool
	DetectGCP         bool
	DetectPCF         bool
	DetectDocker      bool
	LogicalProcessors int
	TotalRAMMIB       int
	BillingHostname   string
}

type override struct {
	LogicalProcessors *int   `json:"logical_processors,omitempty"`
	TotalRAMMIB       *int   `json:"total_ram_mib,omitempty"`
	BillingHostname   string `json:"hostname,omitempty"`
}

// Data contains utilization system information.
type Data struct {
	MetadataVersion int `json:"metadata_version"`
	// Although `runtime.NumCPU()` will never fail, this field is a pointer
	// to facilitate the cross agent tests.
	LogicalProcessors *int      `json:"logical_processors"`
	RAMMiB            *uint64   `json:"total_ram_mib"`
	Hostname          string    `json:"hostname"`
	BootID            string    `json:"boot_id,omitempty"`
	Vendors           *vendors  `json:"vendors,omitempty"`
	Config            *override `json:"config,omitempty"`
}

var (
	sampleRAMMib    = uint64(1024)
	sampleLogicProc = int(16)
	// SampleData contains sample utilization data useful for testing.
	SampleData = Data{
		MetadataVersion:   metadataVersion,
		LogicalProcessors: &sampleLogicProc,
		RAMMiB:            &sampleRAMMib,
		Hostname:          "my-hostname",
	}
)

type docker struct {
	ID string `json:"id,omitempty"`
}

type vendors struct {
	AWS    *aws    `json:"aws,omitempty"`
	Azure  *azure  `json:"azure,omitempty"`
	GCP    *gcp    `json:"gcp,omitempty"`
	PCF    *pcf    `json:"pcf,omitempty"`
	Docker *docker `json:"docker,omitempty"`
}

func (v *vendors) isEmpty() bool {
	return v.AWS == nil && v.Azure == nil && v.GCP == nil && v.PCF == nil && v.Docker == nil
}

func overrideFromConfig(config Config) *override {
	ov := &override{}

	if 0 != config.LogicalProcessors {
		x := config.LogicalProcessors
		ov.LogicalProcessors = &x
	}
	if 0 != config.TotalRAMMIB {
		x := config.TotalRAMMIB
		ov.TotalRAMMIB = &x
	}
	ov.BillingHostname = config.BillingHostname

	if "" == ov.BillingHostname &&
		nil == ov.LogicalProcessors &&
		nil == ov.TotalRAMMIB {
		ov = nil
	}
	return ov
}

// Gather gathers system utilization data.
func Gather(config Config, lg logger.Logger) *Data {
	client := &http.Client{
		Timeout: providerTimeout,
	}
	return gatherWithClient(config, lg, client)
}

func gatherWithClient(config Config, lg logger.Logger, client *http.Client) *Data {
	var wg sync.WaitGroup

	cpu := runtime.NumCPU()
	uDat := &Data{
		MetadataVersion:   metadataVersion,
		LogicalProcessors: &cpu,
		Vendors:           &vendors{},
	}

	warnGatherError := func(datatype string, err error) {
		lg.Debug("error gathering utilization data", map[string]interface{}{
			"error":    err.Error(),
			"datatype": datatype,
		})
	}

	// This closure allows us to run each gather function in a separate goroutine
	// and wait for them at the end by closing over the wg WaitGroup we
	// instantiated at the start of the function.
	goGather := func(datatype string, gather func(*Data, *http.Client) error) {
		wg.Add(1)
		go func() {
			// Note that locking around util is not necessary since
			// WaitGroup provides acts as a memory barrier:
			// https://groups.google.com/d/msg/golang-nuts/5oHzhzXCcmM/utEwIAApCQAJ
			// Thus this code is fine as long as each routine is
			// modifying a different field of util.
			defer wg.Done()
			if err := gather(uDat, client); err != nil {
				warnGatherError(datatype, err)
			}
		}()
	}

	// Kick off gathering which requires network calls in goroutines.

	if config.DetectAWS {
		goGather("aws", gatherAWS)
	}

	if config.DetectAzure {
		goGather("azure", gatherAzure)
	}

	if config.DetectPCF {
		goGather("pcf", gatherPCF)
	}

	if config.DetectGCP {
		goGather("gcp", gatherGCP)
	}

	// Do non-network gathering sequentially since it is fast.

	if id, err := sysinfo.BootID(); err != nil {
		if err != sysinfo.ErrFeatureUnsupported {
			warnGatherError("bootid", err)
		}
	} else {
		uDat.BootID = id
	}

	if config.DetectDocker {
		if id, err := sysinfo.DockerID(); err != nil {
			if err != sysinfo.ErrFeatureUnsupported &&
				err != sysinfo.ErrDockerNotFound {
				warnGatherError("docker", err)
			}
		} else {
			uDat.Vendors.Docker = &docker{ID: id}
		}
	}

	if hostname, err := sysinfo.Hostname(); nil == err {
		uDat.Hostname = hostname
	} else {
		warnGatherError("hostname", err)
	}

	if bts, err := sysinfo.PhysicalMemoryBytes(); nil == err {
		mib := sysinfo.BytesToMebibytes(bts)
		uDat.RAMMiB = &mib
	} else {
		warnGatherError("memory", err)
	}

	// Now we wait for everything!
	wg.Wait()

	// Override whatever needs to be overridden.
	uDat.Config = overrideFromConfig(config)

	if uDat.Vendors.isEmpty() {
		// Per spec, we MUST NOT send any vendors hash if it's empty.
		uDat.Vendors = nil
	}

	return uDat
}
