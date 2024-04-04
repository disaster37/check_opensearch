package check

import (
	"context"

	"emperror.dev/errors"
	nagiosPlugin "github.com/disaster37/go-nagios"
	"github.com/disaster37/opensearch/v2"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// CheckILMError Wrap cli argument and call check
func CheckTransformError(c *cli.Context) error {

	monitorES, err := manageOpensearchGlobalParameters(c)
	if err != nil {
		return err
	}

	monitoringData, err := monitorES.CheckTransformError(c.String("name"), c.StringSlice("exclude"))
	if err != nil {
		return err
	}
	monitoringData.ToSdtOut()

	return nil

}

func (h *DefaultCheck) CheckTransformError(transformName string, excludeTransforms []string) (monitoringData *nagiosPlugin.Monitoring, err error) {
	if transformName == "" {
		transformName = "_all"
	}
	log.Debugf("TransformName: %s", transformName)
	log.Debugf("ExcludeTransform: %+v", excludeTransforms)
	monitoringData = nagiosPlugin.NewMonitoring()

	resSearch, err := h.client.TransformSearchJob().
		Search(transformName).
		Size(1000).
		Do(context.Background())
	if err != nil {
		if opensearch.IsNotFound(err) {
			monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_UNKNOWN)
			monitoringData.AddMessage("Transform %s not found", transformName)
			return monitoringData, nil
		}
		return nil, errors.Wrapf(err, "Error when search Transform  %s", transformName)
	}

	// Handle not found transform when id is provided
	if len(resSearch.Transforms) == 0 && transformName != "_all" && transformName != "*" {
		monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_UNKNOWN)
		monitoringData.AddMessage("Transform %s not found", transformName)
		return monitoringData, nil
	}

	// Loop over transform and exclude transform if needed
	var isExclude bool
	var nbTransformStarted int
	var nbTranformFailed int
	var nbTransformStopped int
	for _, transform := range resSearch.Transforms {
		isExclude = false
		for _, excludeTransform := range excludeTransforms {
			if transform.Id == excludeTransform {
				isExclude = true
				log.Debugf("Transform %s is exclude", transform.Id)
				break
			}
		}
		if !isExclude {
			explain, err := h.client.TransformExplainJob(transform.Id).Do(context.Background())
			if err != nil {
				return nil, errors.Wrapf(err, "Error when explain Transform  %s", transformName)
			}

			// Status available
			// https://github.com/opensearch-project/index-management/blob/main/src/main/kotlin/org/opensearch/indexmanagement/transform/model/TransformMetadata.kt#L38

			if explain[transform.Id].Status == "init" || explain[transform.Id].Status == "started" {
				nbTransformStarted++
				continue
			} else if explain[transform.Id].Status == "stopped" || explain[transform.Id].Status == "finished" {
				nbTransformStopped++
				continue
			} else {
				nbTranformFailed++
				monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_CRITICAL)
				monitoringData.AddMessage("Transform %s %s: %s", explain[transform.Id].TransformId, explain[transform.Id].Status, explain[transform.Id].FailureReason)
				continue
			}
		}

	}

	monitoringData.AddPerfdataOrDie("nbTransformFailed", nbTranformFailed, "")
	monitoringData.AddPerfdataOrDie("nbTransformStopped", nbTransformStopped, "")
	monitoringData.AddPerfdataOrDie("nbTransformStarted", nbTransformStarted, "")

	if monitoringData.Status() == nagiosPlugin.STATUS_OK {
		if transformName == "_all" || transformName == "*" {
			monitoringData.AddMessage("All transform works fine")
		} else {
			monitoringData.AddMessage("Transform %s works fine", transformName)
		}
	}

	return monitoringData, nil
}
