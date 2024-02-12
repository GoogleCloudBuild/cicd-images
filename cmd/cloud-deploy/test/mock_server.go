package test

import (
	"context"
	"net"
	"testing"

	deploy "cloud.google.com/go/deploy/apiv1"
	"cloud.google.com/go/deploy/apiv1/deploypb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"cloud.google.com/go/storage"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"
)

func CreateGCSClient(t *testing.T, content []byte, bucketName, objName string) *storage.Client {
	t.Helper()
	server := fakestorage.NewServer([]fakestorage.Object{{
		Content: content,
		ObjectAttrs: fakestorage.ObjectAttrs{
			BucketName: bucketName,
			Name:       objName,
		},
	},
	})
	t.Cleanup(server.Stop)
	return server.Client()
}

func CreateCloudDeployClient(t *testing.T, ctx context.Context) *deploy.CloudDeployClient {
	t.Helper()
	fakeCloudDeployServer := &FakeCloudDeployServer{}
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	gsrv := grpc.NewServer()
	deploypb.RegisterCloudDeployServer(gsrv, fakeCloudDeployServer)
	fakeServerAddr := l.Addr().String()
	go func() {
		if err := gsrv.Serve(l); err != nil {
			panic(err)
		}
	}()
	// Create a client.
	client, err := deploy.NewCloudDeployClient(ctx,
		option.WithEndpoint(fakeServerAddr),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithInsecure()),
	)
	if err != nil {
		t.Fatal(err)
	}

	return client
}

type FakeCloudDeployServer struct {
	deploypb.UnimplementedCloudDeployServer
}

func (f *FakeCloudDeployServer) CreateRelease(ctx context.Context, req *deploypb.CreateReleaseRequest) (*longrunningpb.Operation, error) {
	return &longrunningpb.Operation{
		Done: true,
		Result: &longrunningpb.Operation_Response{
			Response: &anypb.Any{
				TypeUrl: "google.cloud.deploy.v1.Release",
			}}}, nil
}

func (f *FakeCloudDeployServer) ListRollouts(ctx context.Context, req *deploypb.ListRolloutsRequest) (*deploypb.ListRolloutsResponse, error) {
	return &deploypb.ListRolloutsResponse{}, nil
}

func (f *FakeCloudDeployServer) CreateRollout(ctx context.Context, req *deploypb.CreateRolloutRequest) (*longrunningpb.Operation, error) {
	return &longrunningpb.Operation{
		Done: true,
		Result: &longrunningpb.Operation_Response{
			Response: &anypb.Any{
				TypeUrl: "google.cloud.deploy.v1.Rollout",
			}}}, nil
}

func (f *FakeCloudDeployServer) GetRelease(ctx context.Context, req *deploypb.GetReleaseRequest) (*deploypb.Release, error) {
	return &deploypb.Release{
		TargetSnapshots: []*deploypb.Target{
			{
				Name:     "test",
				TargetId: "test-id",
			},
		},
		DeliveryPipelineSnapshot: &deploypb.DeliveryPipeline{
			Pipeline: &deploypb.DeliveryPipeline_SerialPipeline{
				SerialPipeline: &deploypb.SerialPipeline{
					Stages: []*deploypb.Stage{
						{
							TargetId: "test-id",
						},
					},
				},
			},
		},
	}, nil
}
func (f *FakeCloudDeployServer) GetDeliveryPipeline(ctx context.Context, req *deploypb.GetDeliveryPipelineRequest) (*deploypb.DeliveryPipeline, error) {
	return &deploypb.DeliveryPipeline{
		Uid: "test-uid",
	}, nil
}
