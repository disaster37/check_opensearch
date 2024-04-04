package check

import (
	"context"
	"fmt"
	"strings"
	"time"

	nagiosPlugin "github.com/disaster37/go-nagios"
	"github.com/disaster37/opensearch/v2"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// CheckSLMError wrap command line to check
func CheckSMError(c *cli.Context) error {

	monitor, err := manageOpensearchGlobalParameters(c)
	if err != nil {
		return err
	}

	monitoringData, err := monitor.CheckSMError(c.String("repository"))
	if err != nil {
		return err
	}
	monitoringData.ToSdtOut()

	return nil
}

// CheckSLMPolicy wrap command line to check
func CheckSLMPolicy(c *cli.Context) error {

	monitorES, err := manageOpensearchGlobalParameters(c)
	if err != nil {
		return err
	}

	monitoringData, err := monitorES.CheckSMPolicy(c.String("name"))
	if err != nil {
		return err
	}
	monitoringData.ToSdtOut()

	return nil

}

// CheckSMError check that there are no SM policy failed on repository
func (h *DefaultCheck) CheckSMError(snapshotRepositoryName string) (res *nagiosPlugin.Monitoring, err error) {

	log.Debugf("snapshotRepositoryName: %s", snapshotRepositoryName)
	monitoringData := nagiosPlugin.NewMonitoring()

	if snapshotRepositoryName != "" {
		if _, err := h.client.SnapshotGetRepository(snapshotRepositoryName).Do(context.Background()); err != nil {
			if opensearch.IsNotFound(err) {
				monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_UNKNOWN)
				monitoringData.AddMessage("Repository %s not found", snapshotRepositoryName)
				return monitoringData, nil
			}
			return nil, err
		}
	}

	snapshotStatusResp, err := h.client.SnapshotStatus().Repository(snapshotRepositoryName).Snapshot("_all").Do(context.Background())
	if err != nil {
		if opensearch.IsNotFound(err) {
			monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_UNKNOWN)
			monitoringData.AddMessage("Repository %s not found", snapshotRepositoryName)
			return monitoringData, nil
		}

		return nil, err
	}

	// Check if there are some snapshot failed
	if (snapshotStatusResp.Snapshots == nil) || (len(snapshotStatusResp.Snapshots) == 0) {
		monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_OK)
		monitoringData.AddMessage("No snapshot on repository %s", snapshotRepositoryName)
		monitoringData.AddPerfdataOrDie("NbSnapshot", 0, "")
		monitoringData.AddPerfdataOrDie("NbSnapshotFailed", 0, "")
		return monitoringData, nil
	}

	nbSnapshot := 0
	snapshotsFailed := make([]opensearch.SnapshotStatus, 0)
	for _, snapshotResponse := range snapshotStatusResp.Snapshots {
		nbSnapshot++
		if snapshotResponse.State != "SUCCESS" && snapshotResponse.State != "IN_PROGRESS" {
			monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_CRITICAL)
			snapshotsFailed = append(snapshotsFailed, snapshotResponse)
		}
	}
	if len(snapshotsFailed) > 0 {
		monitoringData.AddMessage("Some snapshots failed (%d/%d)", nbSnapshot-len(snapshotsFailed), nbSnapshot)
		for _, snapshotFailed := range snapshotsFailed {

			var errorMsg strings.Builder
			for _, failure := range snapshotFailed.Failures {
				errorMsg.WriteString(fmt.Sprintf("\n\tIndice %s on node %s failed with status %s: %s", failure.Indice, failure.NodeID, failure.Status, failure.Reason))
			}

			monitoringData.AddMessage("Snapshot %s failed (%s - %s) with status %s: %s", snapshotFailed.Snapshot, snapshotFailed.StartTime, snapshotFailed.EndTime, snapshotFailed.State, errorMsg.String())
		}
	} else {
		monitoringData.AddMessage("All snapshots are ok (%d/%d)", nbSnapshot, nbSnapshot)
	}

	monitoringData.AddPerfdataOrDie("NbSnapshot", nbSnapshot, "")
	monitoringData.AddPerfdataOrDie("NbSnapshotFailed", len(snapshotsFailed), "")

	return monitoringData, nil
}

func (h *DefaultCheck) CheckSMPolicy(policyName string) (res *nagiosPlugin.Monitoring, err error) {

	log.Debugf("policyName: %s", policyName)
	monitoringData := nagiosPlugin.NewMonitoring()

	if policyName != "" {
		if _, err := h.client.SmGetPolicy(policyName).Do(context.Background()); err != nil {
			if opensearch.IsNotFound(err) {
				monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_UNKNOWN)
				monitoringData.AddMessage("Policy %s not found", policyName)
				return monitoringData, nil
			}
			return nil, err
		}
	}

	var explainSmRes *opensearch.SmExplainPolicyResponse
	if policyName == "" {
		explainSmRes, err = h.client.SmExplainPolicy().Do(context.Background())
	} else {
		explainSmRes, err = h.client.SmExplainPolicy(policyName).Do(context.Background())
	}
	if err != nil {
		if opensearch.IsNotFound(err) {
			monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_OK)
			monitoringData.AddMessage("No SM policy %s", policyName)
			monitoringData.AddPerfdataOrDie("NbSMPolicyt", 0, "")
			monitoringData.AddPerfdataOrDie("NbSMPolicyFailed", 0, "")
			return monitoringData, nil
		}
		return nil, err
	}

	// Check if there are some SLM policy failed
	if len(explainSmRes.Policies) == 0 {
		monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_OK)
		monitoringData.AddMessage("No SM policy %s", policyName)
		monitoringData.AddPerfdataOrDie("NbSMPolicyt", 0, "")
		monitoringData.AddPerfdataOrDie("NbSMPolicyFailed", 0, "")
		return monitoringData, nil
	}

	nbSLMPolicy := 0
	slmPoliciesFailedCreation := make(map[string]opensearch.SmExplainPolicy)
	slmPoliciesFailedDeletion := make(map[string]opensearch.SmExplainPolicy)
	for _, policy := range explainSmRes.Policies {
		nbSLMPolicy++
		if policy.Creation != nil && (policy.Creation.LatestExecution.Status == "FAILED" || policy.Creation.LatestExecution.Status == "TIME_LIMIT_EXCEEDED") {
			monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_CRITICAL)
			slmPoliciesFailedCreation[policy.Name] = policy
		}
		if policy.Deletion != nil && (policy.Deletion.LatestExecution.Status == "FAILED" || policy.Deletion.LatestExecution.Status == "TIME_LIMIT_EXCEEDED") {
			monitoringData.SetStatusOrDie(nagiosPlugin.STATUS_CRITICAL)
			slmPoliciesFailedDeletion[policy.Name] = policy
		}
	}
	if len(slmPoliciesFailedCreation) > 0 {
		monitoringData.AddMessage("Some SM policies failed on doing snapshot (%d/%d)", nbSLMPolicy-len(slmPoliciesFailedCreation), nbSLMPolicy)
		for name, policyFailed := range slmPoliciesFailedCreation {
			monitoringData.AddMessage("SM policy %s failed at %s: %s cause by %s", name, time.UnixMilli(policyFailed.Creation.LatestExecution.EndTime).String(), policyFailed.Creation.LatestExecution.Info.Message, policyFailed.Creation.LatestExecution.Info.Cause)
		}
	}

	if len(slmPoliciesFailedDeletion) > 0 {
		monitoringData.AddMessage("Some SM policies failed on doing clean snapshot (%d/%d)", nbSLMPolicy-len(slmPoliciesFailedDeletion), nbSLMPolicy)
		for name, policyFailed := range slmPoliciesFailedDeletion {
			monitoringData.AddMessage("SM policy %s failed at %s: %s cause by %s", name, time.UnixMilli(policyFailed.Creation.LatestExecution.EndTime).String(), policyFailed.Creation.LatestExecution.Info.Message, policyFailed.Creation.LatestExecution.Info.Cause)
		}
	}

	if len(slmPoliciesFailedCreation) == 0 && len(slmPoliciesFailedDeletion) == 0 {
		monitoringData.AddMessage("All SM policies are ok (%d/%d)", nbSLMPolicy, nbSLMPolicy)
	}

	monitoringData.AddPerfdataOrDie("NbSMPolicy", nbSLMPolicy, "")
	monitoringData.AddPerfdataOrDie("NbSMPolicyFailed", len(slmPoliciesFailedCreation)+len(slmPoliciesFailedDeletion), "")

	return monitoringData, nil
}
