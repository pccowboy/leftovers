package ec2

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/common"
)

type snapshotsClient interface {
	DescribeSnapshots(*awsec2.DescribeSnapshotsInput) (*awsec2.DescribeSnapshotsOutput, error)
	DeleteSnapshot(*awsec2.DeleteSnapshotInput) (*awsec2.DeleteSnapshotOutput, error)
}

type Snapshots struct {
	client snapshotsClient
	logger logger
}

func NewSnapshots(client snapshotsClient, logger logger) Snapshots {
	return Snapshots{
		client: client,
		logger: logger,
	}
}

func (s Snapshots) List(filter string) ([]common.Deletable, error) {
	output, err := s.client.DescribeSnapshots(&awsec2.DescribeSnapshotsInput{
		Filters: []*awsec2.Filter{{
			Name:   aws.String("status"),
			Values: []*string{aws.String("completed")},
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("Describe EC2 Snapshots: %s", err)
	}

	var resources []common.Deletable
	for _, snapshot := range output.Snapshots {
		r := NewSnapshot(s.client, snapshot.SnapshotId)

		if !strings.Contains(r.Name(), filter) {
			continue
		}

		proceed := s.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}