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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"
)

// Environment variables for Datastore configuration:
//
// Emulator mode:
//   - DATASTORE_EMULATOR_HOST: Emulator host (e.g., "localhost:8081")
//   - DATASTORE_PROJECT_ID: Project ID for emulator (default: "test-project")
//
// Real Datastore mode:
//   - DATASTORE_PROJECT_ID: GCP project ID (required)
//   - DATASTORE_CREDENTIALS_FILE: Path to service account JSON (optional, uses ADC if not set)
//   - DATASTORE_TEST_NAMESPACE: Namespace for test entities (required for real Datastore)
//
// Test flags:
//   - -force-delete-ns: Force delete existing entities in test namespace

const (
	envDatastoreCredentials = "DATASTORE_CREDENTIALS_FILE"
	envDatastoreNamespace   = "DATASTORE_TEST_NAMESPACE"
)

// forceDeleteNamespace is a test flag that when set, forces deletion of existing
// entities in the test namespace before running tests. Without this flag, tests
// will fail if entities already exist in the namespace.
var forceDeleteNamespace = flag.Bool("force-delete-ns", false,
	"Force delete existing entities in test namespace before running tests")

// testKinds lists all entity kinds created during tests.
// Used for namespace existence checks and cleanup.
var testKinds = []string{
	testUserKind,
	"TestWorld",
	"TestGame",
	"TestGameState",
	"TestUnit",
}

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
// Supports two modes:
//  1. Real Datastore: Set DATASTORE_PROJECT_ID and DATASTORE_TEST_NAMESPACE
//  2. Emulator: Set DATASTORE_EMULATOR_HOST
//
// Real Datastore takes precedence if DATASTORE_TEST_NAMESPACE is set.
func setupTestClient(t *testing.T) *datastore.Client {
	// Use real Datastore if namespace is configured (and no emulator override)
	if isRealDatastoreConfigured() {
		return setupRealDatastoreClient(t)
	}

	// Fall back to emulator mode
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

// cleanupKindInNamespace deletes all entities of a given kind within a namespace.
func cleanupKindInNamespace(ctx context.Context, client *datastore.Client, kind, namespace string) error {
	q := datastore.NewQuery(kind).Namespace(namespace).KeysOnly()
	keys, err := client.GetAll(ctx, q, nil)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return client.DeleteMulti(ctx, keys)
	}
	return nil
}

// --- Real Datastore helpers ---

// isRealDatastoreConfigured checks if real Datastore (not emulator) is configured.
// Returns true if DATASTORE_TEST_NAMESPACE is set and DATASTORE_EMULATOR_HOST is NOT set.
func isRealDatastoreConfigured() bool {
	// Must have namespace set and NOT be using emulator
	return os.Getenv(envDatastoreNamespace) != "" && !isEmulatorAvailable()
}

// getTestNamespace returns the configured test namespace.
func getTestNamespace() string {
	return os.Getenv(envDatastoreNamespace)
}

// skipIfNoRealDatastore skips the test if real Datastore is not configured.
func skipIfNoRealDatastore(t *testing.T) {
	if !isRealDatastoreConfigured() {
		t.Skip("Skipping: Real Datastore not configured. Set DATASTORE_PROJECT_ID and DATASTORE_TEST_NAMESPACE (without DATASTORE_EMULATOR_HOST).")
	}
}

// skipIfNoDatastore skips the test if neither emulator nor real Datastore is configured.
func skipIfNoDatastore(t *testing.T) {
	if !isEmulatorAvailable() && !isRealDatastoreConfigured() {
		t.Skip("Skipping: No Datastore configured. Set DATASTORE_EMULATOR_HOST for emulator or DATASTORE_TEST_NAMESPACE for real Datastore.")
	}
}

// expandPath expands ~ to the user's home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// setupRealDatastoreClient creates a Datastore client for testing against real GCP Datastore.
// It handles:
//   - Loading credentials from file (with ~ expansion) or ADC
//   - Checking namespace for existing entities (fails unless -force-delete-ns is set)
//   - Registering cleanup to delete all entities in namespace after tests
func setupRealDatastoreClient(t *testing.T) *datastore.Client {
	skipIfNoRealDatastore(t)

	ctx := context.Background()
	projectID := getProjectID()
	namespace := getTestNamespace()

	if projectID == "" || projectID == "test-project" {
		t.Fatal("DATASTORE_PROJECT_ID must be set to a real GCP project ID for real Datastore tests")
	}

	var client *datastore.Client
	var err error

	credFile := os.Getenv(envDatastoreCredentials)
	if credFile != "" {
		credFile = expandPath(credFile)
		// Verify file exists
		if _, statErr := os.Stat(credFile); os.IsNotExist(statErr) {
			t.Fatalf("Credentials file does not exist: %s", credFile)
		}
		client, err = datastore.NewClient(ctx, projectID, option.WithCredentialsFile(credFile))
	} else {
		// Use Application Default Credentials
		client, err = datastore.NewClient(ctx, projectID)
	}

	if err != nil {
		t.Fatalf("Failed to create Datastore client: %v", err)
	}

	// Ensure namespace is empty (or force delete)
	ensureNamespaceEmpty(t, ctx, client, namespace, testKinds)

	// Register cleanup to delete all entities in namespace
	t.Cleanup(func() {
		if cleanupErr := cleanupNamespace(ctx, client, namespace); cleanupErr != nil {
			t.Logf("Warning: Failed to cleanup namespace %q: %v", namespace, cleanupErr)
		}
		client.Close()
	})

	return client
}

// namespaceHasEntities checks if any entities exist in the given namespace for the specified kinds.
func namespaceHasEntities(ctx context.Context, client *datastore.Client, namespace string, kinds []string) (bool, error) {
	for _, kind := range kinds {
		q := datastore.NewQuery(kind).Namespace(namespace).KeysOnly().Limit(1)
		keys, err := client.GetAll(ctx, q, nil)
		if err != nil {
			return false, fmt.Errorf("failed to query kind %q in namespace %q: %w", kind, namespace, err)
		}
		if len(keys) > 0 {
			return true, nil
		}
	}
	return false, nil
}

// cleanupNamespace deletes ALL entities in the given namespace.
// It queries the __kind__ metadata to find all kinds, then deletes entities for each kind.
func cleanupNamespace(ctx context.Context, client *datastore.Client, namespace string) error {
	// First, try to get all kinds in the namespace using metadata query
	// Note: __kind__ metadata may not be available in all cases, so we also
	// iterate over known test kinds as a fallback
	kindsToClean := make(map[string]bool)

	// Add known test kinds
	for _, kind := range testKinds {
		kindsToClean[kind] = true
	}

	// Try to discover additional kinds via metadata
	q := datastore.NewQuery("__kind__").Namespace(namespace).KeysOnly()
	keys, err := client.GetAll(ctx, q, nil)
	if err == nil {
		for _, key := range keys {
			// The kind name is the key's name
			if key.Name != "" && !strings.HasPrefix(key.Name, "__") {
				kindsToClean[key.Name] = true
			}
		}
	}

	// Delete all entities for each kind
	for kind := range kindsToClean {
		if err := cleanupKindInNamespace(ctx, client, kind, namespace); err != nil {
			return fmt.Errorf("failed to cleanup kind %q: %w", kind, err)
		}
	}

	return nil
}

// ensureNamespaceEmpty checks that the test namespace is empty for the specified kinds.
// If entities exist and -force-delete-ns flag is not set, the test fails with a clear error.
// If -force-delete-ns is set, existing entities are deleted before proceeding.
func ensureNamespaceEmpty(t *testing.T, ctx context.Context, client *datastore.Client, namespace string, kinds []string) {
	hasEntities, err := namespaceHasEntities(ctx, client, namespace, kinds)
	if err != nil {
		t.Fatalf("Failed to check namespace for existing entities: %v", err)
	}

	if hasEntities {
		if *forceDeleteNamespace {
			t.Logf("Namespace %q has existing entities. Deleting due to -force-delete-ns flag...", namespace)
			if err := cleanupNamespace(ctx, client, namespace); err != nil {
				t.Fatalf("Failed to cleanup existing entities in namespace %q: %v", namespace, err)
			}
			t.Logf("Cleanup complete.")
		} else {
			t.Fatalf(`Namespace %q already has entities for test kinds %v.

This is a safety check to prevent accidental data loss. To proceed:

Option 1: Use a different namespace
  export DATASTORE_TEST_NAMESPACE=my-unique-test-namespace

Option 2: Force delete existing entities
  go test ./tests/datastore/... -args -force-delete-ns

Option 3: Manually delete entities in the namespace first
`, namespace, kinds)
		}
	}
}
