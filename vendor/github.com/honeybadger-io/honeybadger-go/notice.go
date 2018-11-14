package honeybadger

import (
	"encoding/json"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/pborman/uuid"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

// ErrorClass represents the class name of the error which is sent to
// Honeybadger.
type ErrorClass struct {
	Name string
}

// Fingerprint represents the fingerprint of the error, which controls grouping
// in Honeybadger.
type Fingerprint struct {
	Content string
}

func (f *Fingerprint) String() string {
	return f.Content
}

// Notice is a representation of the error which is sent to Honeybadger, and
// implements the Payload interface.
type Notice struct {
	APIKey       string
	Error        Error
	Token        string
	ErrorMessage string
	ErrorClass   string
	Tags         []string
	Hostname     string
	Env          string
	Backtrace    []*Frame
	ProjectRoot  string
	Context      Context
	Params       Params
	CGIData      CGIData
	URL          string
	Fingerprint  string
}

func (n *Notice) asJSON() *hash {
	return &hash{
		"api_key": n.APIKey,
		"notifier": &hash{
			"name":    "honeybadger",
			"url":     "https://github.com/honeybadger-io/honeybadger-go",
			"version": VERSION,
		},
		"error": &hash{
			"token":       n.Token,
			"message":     n.ErrorMessage,
			"class":       n.ErrorClass,
			"tags":        n.Tags,
			"backtrace":   n.Backtrace,
			"fingerprint": n.Fingerprint,
		},
		"request": &hash{
			"context":  n.Context,
			"params":   n.Params,
			"cgi_data": n.CGIData,
			"url":      n.URL,
		},
		"server": &hash{
			"project_root":     n.ProjectRoot,
			"environment_name": n.Env,
			"hostname":         n.Hostname,
			"time":             time.Now().UTC(),
			"pid":              os.Getpid(),
			"stats":            getStats(),
		},
	}
}

func bytesToKB(bytes uint64) float64 {
	return float64(bytes) / 1024.0
}

func getStats() *hash {
	var m, l *hash

	if stat, err := mem.VirtualMemory(); err == nil {
		m = &hash{
			"total":      bytesToKB(stat.Total),
			"free":       bytesToKB(stat.Free),
			"buffers":    bytesToKB(stat.Buffers),
			"cached":     bytesToKB(stat.Cached),
			"free_total": bytesToKB(stat.Free + stat.Buffers + stat.Cached),
		}
	}

	if stat, err := load.Avg(); err == nil {
		l = &hash{
			"one":     stat.Load1,
			"five":    stat.Load5,
			"fifteen": stat.Load15,
		}
	}

	return &hash{"mem": m, "load": l}
}

func (n *Notice) toJSON() []byte {
	out, err := json.Marshal(n.asJSON())
	if err == nil {
		return out
	}
	panic(err)
}

func (n *Notice) setContext(context Context) {
	n.Context.Update(context)
}

func composeStack(stack []*Frame, root string) (frames []*Frame) {
	if root == "" {
		return stack
	}

	re, err := regexp.Compile("^" + regexp.QuoteMeta(root))
	if err != nil {
		return stack
	}

	for _, frame := range stack {
		file := re.ReplaceAllString(frame.File, "[PROJECT_ROOT]")
		frames = append(frames, &Frame{
			File:   file,
			Number: frame.Number,
			Method: frame.Method,
		})
	}
	return
}

func newNotice(config *Configuration, err Error, extra ...interface{}) *Notice {
	notice := Notice{
		APIKey:       config.APIKey,
		Error:        err,
		Token:        uuid.NewRandom().String(),
		ErrorMessage: err.Message,
		ErrorClass:   err.Class,
		Env:          config.Env,
		Hostname:     config.Hostname,
		Backtrace:    composeStack(err.Stack, config.Root),
		ProjectRoot:  config.Root,
		Context:      Context{},
	}

	for _, thing := range extra {
		switch t := thing.(type) {
		case Context:
			notice.setContext(t)
		case ErrorClass:
			notice.ErrorClass = t.Name
		case Tags:
			for _, tag := range t {
				notice.Tags = append(notice.Tags, tag)
			}
		case Fingerprint:
			notice.Fingerprint = t.String()
		case Params:
			notice.Params = t
		case CGIData:
			notice.CGIData = t
		case url.URL:
			notice.URL = t.String()
		}
	}

	return &notice
}
