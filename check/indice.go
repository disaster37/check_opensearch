package check

import (
	"context"

	"emperror.dev/errors"
	nagiosPlugin "github.com/disaster37/go-nagios"
	"github.com/disaster37/opensearch/v2"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// CheckIndiceLocked wrap command line to check
func CheckIndiceLocked(c *cli.Context) error {

	monitorES, err := manageOpensearchGlobalParameters(c)
	if err != nil {
		return err
	}

	monitoringData, err := monitorES.CheckIndiceLocked(c.String("indice"))
	if err != nil {
		return err
	}
	monitoringData.ToSdtOut()

	return nil

}

func (h *DefaultCheck) CheckIndiceLocked(indiceName string) (monitoringData *nagiosPlugin.Monitoring, err error) {

	if indiceName == "" {
		return nil, errors.New("IndiceName can't be empty")
	}
	log.Debugf("IndiceName: %s", indiceName)
	monitoringData = nagiosPlugin.NewMonitoring()

	res, err := h.client.IndexGetSettings(indiceName).Do(context.Background())
	if err != nil {
		if opensearch.IsNotFound(err) {
			monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_UNKNOWN)
			monitoringData.AddMessage("Indice %s not found", indiceName)
			return monitoringData, nil
		}
		return nil, errors.Wrapf(err, "Error when get indice %s", indiceName)
	}

	// Check if there are index that are read only by security
	brokenIndices := make([]string, 0)
	nbIndice := 0
	for indiceName, indiceSetting := range res {
		log.Debugf("%s: %+v", indiceName, indiceSetting)
		if indice, ok := indiceSetting.Settings["index"]; ok {
			if block, ok := indice.(map[string]any)["blocks"]; ok {
				if readOnly, ok := block.(map[string]any)["read_only_allow_delete"]; ok {
					if readOnly.(string) == "true" {
						brokenIndices = append(brokenIndices, indiceName)
					}
				}
			}
		}
		nbIndice++
	}

	if len(brokenIndices) > 0 {
		monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_CRITICAL)
		monitoringData.AddMessage("There are some indice locked (%d/%d)", nbIndice-len(brokenIndices), nbIndice)
		for _, indiceName := range brokenIndices {
			monitoringData.AddMessage("\tIndice %s", indiceName)
		}

	} else {
		monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_OK)
		monitoringData.AddMessage("No indice locked (%d/%d)", nbIndice, nbIndice)
	}

	monitoringData.AddPerfdataOrDie("nbIndices", nbIndice, "")
	monitoringData.AddPerfdataOrDie("nbIndicesLocked", len(brokenIndices), "")

	return monitoringData, nil
}
