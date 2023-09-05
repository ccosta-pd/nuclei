package javascript_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/projectdiscovery/nuclei/v2/pkg/catalog/config"
	"github.com/projectdiscovery/nuclei/v2/pkg/catalog/disk"
	"github.com/projectdiscovery/nuclei/v2/pkg/parsers"
	"github.com/projectdiscovery/nuclei/v2/pkg/progress"
	"github.com/projectdiscovery/nuclei/v2/pkg/protocols"
	"github.com/projectdiscovery/nuclei/v2/pkg/templates"
	"github.com/projectdiscovery/nuclei/v2/pkg/testutils"
	"github.com/projectdiscovery/ratelimit"
	"github.com/stretchr/testify/require"
)

var (
	testcases = []string{
		"testcases/ms-sql-detect.yaml",
		"testcases/redis-pass-brute.yaml",
		"testcases/ssh-server-fingerprint.yaml",
	}
	executerOpts protocols.ExecutorOptions
)

func setup() {
	options := testutils.DefaultOptions
	testutils.Init(options)
	progressImpl, _ := progress.NewStatsTicker(0, false, false, false, false, 0)

	executerOpts = protocols.ExecutorOptions{
		Output:       testutils.NewMockOutputWriter(),
		Options:      options,
		Progress:     progressImpl,
		ProjectFile:  nil,
		IssuesClient: nil,
		Browser:      nil,
		Catalog:      disk.NewCatalog(config.DefaultConfig.TemplatesDirectory),
		RateLimiter:  ratelimit.New(context.Background(), uint(options.RateLimit), time.Second),
	}
	workflowLoader, err := parsers.NewLoader(&executerOpts)
	if err != nil {
		log.Fatalf("Could not create workflow loader: %s\n", err)
	}
	executerOpts.WorkflowLoader = workflowLoader
}

func TestCompile(t *testing.T) {
	setup()
	for index, tpl := range testcases {
		// parse template
		template, err := templates.Parse(tpl, nil, executerOpts)
		require.Nilf(t, err, "failed to parse %v", tpl)

		// compile template
		err = template.Executer.Compile()
		require.Nilf(t, err, "failed to compile %v", tpl)

		switch index {
		case 0:
			// requests count should be 1
			require.Equal(t, 1, template.TotalRequests, "template : %v", tpl)
		case 1:
			// requests count should be 6 i.e 5 generator payloads + 1 precondition request
			require.Equal(t, 5+1, template.TotalRequests, "template : %v", tpl)
		case 2:
			// requests count should be 1
			require.Equal(t, 1, template.TotalRequests, "template : %v", tpl)
		}
	}
}
