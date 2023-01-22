// Package sitedossier logic
package sitedossier

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/projectdiscovery/subfinder/v2/pkg/core"
	"github.com/projectdiscovery/subfinder/v2/pkg/session"
	"github.com/projectdiscovery/subfinder/v2/pkg/subscraping"
)

var reNext = regexp.MustCompile(`<a href="([A-Za-z0-9/.]+)"><b>`)

// Source is the passive scraping agent
type Source struct {
	subscraping.BaseSource
}

// Source Daemon
func (s *Source) Daemon(ctx context.Context, e *session.Extractor, input <-chan string, output chan<- core.Task) {
	s.init()
	s.BaseSource.Daemon(ctx, e, input, output)
}

// inits the source before passing to daemon
func (s *Source) init() {
	s.BaseSource.SourceName = "sitedossier"
	s.BaseSource.Default = false
	s.BaseSource.Recursive = false
	s.BaseSource.RequiresKey = true
	s.BaseSource.CreateTask = s.dispatcher
}

func (s *Source) dispatcher(domain string) core.Task {
	task := core.Task{
		Domain: domain,
	}

	task.RequestOpts = &session.RequestOpts{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("http://www.sitedossier.com/parentdomain/%s", domain),
		Source: "sitedossier",
	}

	task.OnResponse = func(t *core.Task, resp *http.Response, executor *core.Executor) error {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("expected status code 200 got %v", resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		src := string(body)
		for _, match := range executor.Extractor.Get(domain).FindAllString(src, -1) {
			executor.Result <- core.Result{Source: "sitedossier", Type: core.Subdomain, Value: match}
		}
		match1 := reNext.FindStringSubmatch(src)
		if len(match1) > 0 {
			tx := t.Clone()
			tx.RequestOpts.URL = "http://www.sitedossier.com" + match1[1]
		}
		return nil
	}
	return task
}
