package check

import (
	"context"
	"os"
	"testing"

	"github.com/disaster37/opensearch/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"k8s.io/utils/ptr"
)

type CheckESTestSuite struct {
	suite.Suite
	check Check
}

func (s *CheckESTestSuite) SetupSuite() {
	// Init logger
	logrus.SetFormatter(new(prefixed.TextFormatter))
	logrus.SetLevel(logrus.DebugLevel)

	// Init client
	username := os.Getenv("OPENSEARCH_USERNAME")
	password := os.Getenv("OPENSEARCH_PASSWORD")

	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "vLPeJYa8.3RqtZCcAK6jNz"
	}

	checkOpensearch, err := NewCheck("https://opensearch.svc:9200", username, password, true)
	if err != nil {
		panic(err)
	}

	s.check = checkOpensearch

	client := s.check.(*DefaultCheck).client

	// Create indexes
	if _, err := client.CreateIndex("lock").Body(`
{
	"settings": {
		"index": {
			"blocks": {
				"read_only_allow_delete": true
			}
		}
	}
}`).Do(context.Background()); err != nil {
		panic(err)
	}

	if _, err := client.CreateIndex("test").Body(`
{
	"settings":{
		"number_of_shards":1,
		"number_of_replicas":0
	},
	"mappings":{
		"properties":{
			"user":{
				"type":"keyword"
			},
			"message":{
				"type":"text",
				"store": true,
				"fielddata": true
			},
			"tags":{
				"type":"keyword"
			},
			"location":{
				"type":"geo_point"
			},
			"suggest_field":{
				"type":"completion"
			}
		}
	}
}
	`).Do(context.Background()); err != nil {
		panic(err)
	}

	// Create transform
	if _, err := client.TransformPutJob("test").Body(&opensearch.TransformPutJob{
		Transform: opensearch.TransformJobBase{
			Enabled:    ptr.To[bool](true),
			Continuous: ptr.To[bool](true),
			Schedule: map[string]any{
				"interval": map[string]any{
					"period":     1,
					"unit":       "Minutes",
					"start_time": 1602100553,
				},
			},
			Description:        ptr.To[string]("Sample transform job"),
			SourceIndex:        "test",
			TargetIndex:        "sample_target",
			DataSelectionQuery: opensearch.NewMatchAllQuery(),
			PageSize:           1,
			Groups: []any{
				map[string]any{
					"terms": map[string]any{
						"source_field": "user",
						"target_field": "user",
					},
				},
				map[string]any{
					"terms": map[string]any{
						"source_field": "tags",
						"target_field": "tags",
					},
				},
			},
			Aggregations: map[string]any{
				"quantity": map[string]any{
					"sum": map[string]any{
						"field": "total_quantity",
					},
				},
			},
		},
	}).Do(context.Background()); err != nil {
		panic(err)
	}

	// Create repository
	if _, err := client.SnapshotCreateRepository("snapshot").BodyString(`
{
	"type": "fs",
		"settings": {
		"location": "/usr/share/opensearch/backup",
		"compress": true
		}
}
	`).Do(context.Background()); err != nil {
		panic(err)
	}

	// Create Snapshot Management policy
	if _, err := client.SmPostPolicy("test").Body(&opensearch.SmPutPolicy{
		Enabled: ptr.To[bool](true),
		SnapshotConfig: opensearch.SmPolicySnapshotConfig{
			Repository: "snapshot",
			Indices:    ptr.To[string]("*"),
		},
		Creation: opensearch.SmPolicyCreation{
			Schedule: map[string]any{
				"cron": map[string]any{
					"expression": "0 8 * * *",
					"timezone":   "UTC",
				},
			},
			TimeLimit: ptr.To[string]("1h"),
		},
	}).Do(context.Background()); err != nil {
		panic(err)
	}
}

func (s *CheckESTestSuite) TearDownSuite() {
	client := s.check.(*DefaultCheck).client

	// Delete transform
	if _, err := client.TransformDeleteJob("test").Force(true).Do(context.Background()); err != nil {
		panic(err)
	}

	// Delete indexs
	if _, err := client.DeleteIndex("lock", "test").Do(context.Background()); err != nil {
		panic(err)
	}

	// Delete SM policy
	if _, err := client.SmDeletePolicy("test").Do(context.Background()); err != nil {
		panic(err)
	}
}

func (s *CheckESTestSuite) SetupTest() {
	// Do somethink before each test
}

func TestCheckESTestSuite(t *testing.T) {
	suite.Run(t, new(CheckESTestSuite))
}
