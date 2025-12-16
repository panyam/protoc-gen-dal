// Copyright 2025 Sri Panyam
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datastore

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/datastore"
)

// Environment variables for Datastore emulator configuration:
//   - DATASTORE_EMULATOR_HOST: Emulator host (e.g., "localhost:8081")
//   - DATASTORE_PROJECT_ID: Project ID for emulator (default: "test-project")

// isEmulatorAvailable checks if the Datastore emulator is running.
func isEmulatorAvailable() bool {
	return os.Getenv("DATASTORE_EMULATOR_HOST") != ""
}

// skipIfNoEmulator skips the test if no emulator is available.
func skipIfNoEmulator(t *testing.T) {
	if !isEmulatorAvailable() {
		t.Skip("Skipping: DATASTORE_EMULATOR_HOST not set. Run with Datastore emulator for integration tests.")
	}
}

// getProjectID returns the project ID for testing.
func getProjectID() string {
	projectID := os.Getenv("DATASTORE_PROJECT_ID")
	if projectID == "" {
		projectID = "test-project"
	}
	return projectID
}

// setupTestClient creates a Datastore client for testing.
// Requires DATASTORE_EMULATOR_HOST to be set.
func setupTestClient(t *testing.T) *datastore.Client {
	skipIfNoEmulator(t)

	ctx := context.Background()
	client, err := datastore.NewClient(ctx, getProjectID())
	if err != nil {
		t.Fatalf("Failed to create Datastore client: %v", err)
	}

	t.Cleanup(func() {
		client.Close()
	})

	return client
}

// cleanupKind deletes all entities of a given kind.
// Used to clean up test data between tests.
func cleanupKind(ctx context.Context, client *datastore.Client, kind string) error {
	q := datastore.NewQuery(kind).KeysOnly()
	keys, err := client.GetAll(ctx, q, nil)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return client.DeleteMulti(ctx, keys)
	}
	return nil
}
