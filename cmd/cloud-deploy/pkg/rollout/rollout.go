package rollout

import (
	"context"
	"fmt"
	"strings"

	deploy "cloud.google.com/go/deploy/apiv1"
	"cloud.google.com/go/deploy/apiv1/deploypb"
	"github.com/GoogleCloudBuild/cicd-images/cmd/cloud-deploy/pkg/config"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

const (
	RolloutInTargetFilterTemplate = "targetId=\"%s\""
	RolloutIdTemplate             = "%s-to-%s-%04d"
)

func CreateRollout(ctx context.Context, cdClient *deploy.CloudDeployClient, flags *config.ReleaseConfiguration) error {
	releaseGetReq := &deploypb.GetReleaseRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s/releases/%s", flags.ProjectId, flags.Region, flags.DeliveryPipeline, flags.Release),
	}

	release, err := cdClient.GetRelease(ctx, releaseGetReq)
	if err != nil {
		return fmt.Errorf("err getting release: %w", err)
	}

	toTargetId, err := getToTargetId(ctx, release, flags)
	if err != nil {
		return err
	}
	finalRollOutId, err := generateRolloutId(ctx, cdClient, toTargetId, release, flags)
	if err != nil {
		return err
	}

	req := &deploypb.CreateRolloutRequest{
		Parent:          fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s/releases/%s", flags.ProjectId, flags.Region, flags.DeliveryPipeline, flags.Release),
		RolloutId:       finalRollOutId,
		RequestId:       uuid.NewString(),
		StartingPhaseId: flags.InitialRolloutPhaseId,
		Rollout: &deploypb.Rollout{
			Name:        fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s/releases/%s/rollouts/%s", flags.ProjectId, flags.Region, flags.DeliveryPipeline, flags.Release, finalRollOutId),
			TargetId:    toTargetId,
			Annotations: flags.InitialRolloutAnotations,
			Labels:      flags.InitialRolloutLabels,
		},
	}

	fmt.Printf("Creating Cloud Deploy rollout: %s in target %s... \n", finalRollOutId, toTargetId)
	op, err := cdClient.CreateRollout(ctx, req)
	if err != nil {
		return err
	}
	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Created Cloud Deploy rollout\n")

	// check if needs approval
	var targetObj *deploypb.Target
	for _, t := range release.TargetSnapshots {
		if t.TargetId == toTargetId {
			targetObj = t
			break
		}
	}
	if targetObj != nil && targetObj.RequireApproval {
		fmt.Println("The rollout is pending approval...")
	}

	return nil
}

func getToTargetId(ctx context.Context, release *deploypb.Release, flags *config.ReleaseConfiguration) (string, error) {
	var toTargetId string
	if flags.ToTarget != "" {
		toTargetId = flags.ToTarget
	} else {
		// promote to the first target by default
		stages := release.DeliveryPipelineSnapshot.GetSerialPipeline().Stages
		if len(stages) == 0 {
			return "", fmt.Errorf("no pipeline stages in the release %s", flags.Release)
		}

		toTargetId = stages[0].TargetId
	}

	return toTargetId, nil
}

func generateRolloutId(ctx context.Context, cdClient *deploy.CloudDeployClient, toTargetId string, release *deploypb.Release, flags *config.ReleaseConfiguration) (string, error) {
	filterStr := fmt.Sprintf(RolloutInTargetFilterTemplate, toTargetId)

	// get existing rollouts
	req := &deploypb.ListRolloutsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s/releases/%s", flags.ProjectId, flags.Region, flags.DeliveryPipeline, flags.Release),
		Filter: filterStr,
	}
	it := cdClient.ListRollouts(ctx, req)
	existingRollouts := map[string]bool{}
	for {
		r, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "", fmt.Errorf("unexpected error when iterating rollouts: %w", err)
		}

		// get rollout name
		parts := strings.Split(r.Name, "/")
		relateiveName := strings.Split(r.Name, "/")[len(parts)-1]

		existingRollouts[relateiveName] = true
	}

	for i := 1; i <= 1000; i++ {
		newName := fmt.Sprintf(RolloutIdTemplate, flags.Release, toTargetId, i)
		_, ok := existingRollouts[newName]
		if !ok {
			return newName, nil
		}
	}

	return "", fmt.Errorf("rollout name space exhausted in release %s. Use --rollout-id to specify rollout ID", release.Name)
}
