package libs

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	firebase "firebase.google.com/go"
)

func GetFirebaseClient(ctx context.Context, projectId string) (*firestore.Client, error) {
	config := &firebase.Config{
		ProjectID: projectId,
	}

	app, err := firebase.NewApp(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firebase app: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firestore service: %v", err)
	}

	return client, nil
}

func GetLineBotAccessToken(ctx context.Context, projectId, secretName string) (string, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secret manager client: %v", err)
	}
	defer client.Close()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectId, secretName),
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %v", err)
	}

	return string(result.Payload.Data), nil
}
