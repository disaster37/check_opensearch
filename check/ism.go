package check

import (
	"context"
	"net/http"

	nagiosPlugin "github.com/disaster37/go-nagios"
	"github.com/disaster37/opensearch/v2"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// CheckISMError Wrap cli argument and call check
func CheckISMError(c *cli.Context) error {
	monitor, err := manageOpensearchGlobalParameters(c)
	if err != nil {
		return err
	}

	monitoringData, err := monitor.CheckISMError(c.String("indice"), c.StringSlice("exclude"))
	if err != nil {
		return err
	}
	monitoringData.ToSdtOut()

	return nil
}

func (h *DefaultCheck) CheckISMError(indiceName string, excludeIndices []string) (res *nagiosPlugin.Monitoring, err error) {
	log.Debugf("IndiceName: %s", indiceName)
	log.Debugf("ExcludeIndices: %+v", excludeIndices)
	monitoringData := nagiosPlugin.NewMonitoring()

	// Get current ISM state for given index
	explainResp, err := h.client.IsmExplainPolicy(indiceName).Do(context.Background())
	if err != nil {
		if opensearch.IsStatusCode(err, http.StatusBadRequest) {
			monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_UNKNOWN)
			monitoringData.AddMessage("Indice %s not found", indiceName)
			return monitoringData, nil
		}
		return nil, err
	}

	// Check if there are some ILM polices that failed
	if explainResp.TotalManagedIndices == 0 {
		monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_OK)
		monitoringData.AddMessage("No error found on indice %s", indiceName)
		monitoringData.AddPerfdataOrDie("NbIndiceFailed", 0, "")
		return monitoringData, nil
	}

	// Remove exclude indices
	for _, indiceExcludeName := range excludeIndices {
		if _, ok := explainResp.Indexes[indiceExcludeName]; ok {
			log.Debugf("Indice %s is exclude", indiceExcludeName)
			delete(explainResp.Indexes, indiceExcludeName)
		}
	}

	// Compute error
	if len(explainResp.Indexes) == 0 {
		monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_OK)
		monitoringData.AddMessage("No error found on indice %s", indiceName)
		monitoringData.AddPerfdataOrDie("NbIndiceFailed", 0, "")
		return monitoringData, nil
	}

	nbFailedIndex := 0

	for _, explain := range explainResp.Indexes {
		if explain.Action.Failed {
			nbFailedIndex++
			monitoringData.AddMessage("Indice %s (%s): failed on step %s. %s because of %s", explain.Index, explain.PolicyId, explain.Action.Name, explain.Info.Message, explain.Info.Cause)
		}
	}

	if nbFailedIndex > 0 {
		monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_CRITICAL)
	}

	monitoringData.AddPerfdataOrDie("NbIndiceFailed", nbFailedIndex, "")

	return monitoringData, nil
}
