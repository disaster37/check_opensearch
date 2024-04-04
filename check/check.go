package check

import (
	"crypto/tls"
	"net/http"

	"emperror.dev/errors"
	nagiosPlugin "github.com/disaster37/go-nagios"
	"github.com/disaster37/opensearch/v2"
	"github.com/disaster37/opensearch/v2/config"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"k8s.io/utils/ptr"
)

// DefaultCheck implement the interface of Check
type DefaultCheck struct {
	client *opensearch.Client
}

// Check is interface of Opensearch monitoring
type Check interface {
	CheckISMError(indiceName string, excludeIndices []string) (*nagiosPlugin.Monitoring, error)
	CheckSMError(snapshotRepositoryName string) (*nagiosPlugin.Monitoring, error)
	CheckSMPolicy(policyName string) (*nagiosPlugin.Monitoring, error)
	CheckIndiceLocked(indiceName string) (*nagiosPlugin.Monitoring, error)
	CheckTransformError(transformName string, excludeTransforms []string) (*nagiosPlugin.Monitoring, error)
}

func manageOpensearchGlobalParameters(c *cli.Context) (Check, error) {

	if c.String("url") == "" {
		return nil, errors.New("You must set --url parameter")
	}

	return NewCheck(c.String("url"), c.String("user"), c.String("password"), c.Bool("self-signed-certificate"))

}

// NewCheck permit to initialize connexion on Opensearch cluster
func NewCheck(URL string, username string, password string, disableTLSVerification bool) (Check, error) {

	if URL == "" {
		return nil, errors.New("URL can't be empty")
	}
	log.Debugf("URL: %s", URL)
	log.Debugf("User: %s", username)
	log.Debugf("Password: xxx")
	check := &DefaultCheck{}

	cfg := &config.Config{
		URLs:        []string{URL},
		Sniff:       ptr.To[bool](false),
		Healthcheck: ptr.To[bool](false),
	}
	if username != "" && password != "" {
		cfg.Username = username
		cfg.Password = password
	}
	if disableTLSVerification {
		cfg.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}
	client, err := opensearch.NewClientFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	check.client = client
	return check, nil
}
