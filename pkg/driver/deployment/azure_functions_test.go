package deployment_test

import (
	"archive/zip"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vhive-serverless/loader/pkg/common"
	"github.com/vhive-serverless/loader/pkg/driver/deployment"
)

func TestZipHealth(t *testing.T) {
	// Define paths and expected structure
	zipFilePath := "azurefunctions.zip"
	baseDir := "azure_functions_for_zip"
	expectedFunctionCount := 2 // Update if needed

	expectedFunctionFiles := []string{
		"azurefunctionsworkload.py",
		"function.json",
	}

	expectedRootFiles := []string{
		"requirements.txt",
		"host.json",
	}

	// Create test functions to simulate a real deployment
	functions := []*common.Function{
		{Name: "function0"},
		{Name: "function1"},
	}

	// Change the working directory to project root
	err := os.Chdir("../../../")
	if err != nil {
		t.Fatalf("Failed to change working directory: %s", err)
	}

	// Log current working directory for debugging
	wd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current working directory")
	t.Logf("Current working directory: %s", wd)

	// Step 1: Create function folders
	err = deployment.CreateFunctionFolders(baseDir, functions)
	assert.NoError(t, err, "Failed to create function folders")

	// Step 2: Zip the function app files
	err = deployment.ZipFunctionAppFiles()
	assert.NoError(t, err, "Failed to create function app zip")

	// Step 3: Validate if zip file exists
	if _, err := os.Stat(zipFilePath); os.IsNotExist(err) {
		t.Fatalf("Zip file does not exist: %s", zipFilePath)
	}

	// Step 4: Open zip file
	r, err := zip.OpenReader(zipFilePath)
	assert.NoError(t, err, "Failed to open zip file")
	defer r.Close()

	// Step 5: Prepare expected files map
	expectedFiles := make(map[string]bool)

	for i := 0; i < expectedFunctionCount; i++ {
		functionFolder := fmt.Sprintf("function%d/", i)
		for _, file := range expectedFunctionFiles {
			expectedFiles[functionFolder+file] = false
		}
	}
	for _, file := range expectedRootFiles {
		expectedFiles[file] = false
	}

	// Step 6: Check the files inside the zip
	for _, f := range r.File {
		filePath := f.Name
		filePath = strings.TrimPrefix(filePath, "./") // Normalize path

		if _, exists := expectedFiles[filePath]; exists {
			expectedFiles[filePath] = true
		}
	}

	// Step 7: Ensure all expected files are present
	for file, found := range expectedFiles {
		assert.True(t, found, "Missing expected file in zip: "+file)
	}

	t.Log("Zip file structure validation passed!")
}

func TestCleanup(t *testing.T) {
	err := deployment.CleanUpDeploymentFiles("azure_functions_for_zip", "azurefunctions.zip")
	assert.NoError(t, err, "Cleanup function returned an error")

	// Verify Cleanup
	t.Log("Verifying cleanup of temp local files...")

	// Check that temporary files and folders are removed
	_, err = os.Stat("azure_functions_for_zip")
	assert.True(t, os.IsNotExist(err), "Temp directory should be deleted")
	_, err = os.Stat("azurefunctions.zip")
	assert.True(t, os.IsNotExist(err), "Temp zip file should be deleted")

	t.Log("Cleaning of local files passed!")

}
