import re
import os

#Script to extract execute_function from trace_func.py and inject it into azureworkload.py

# Paths to source and target workload files
TRACE_FUNC_PATH = "server/trace-func-py/trace_func.py"
AZURE_WORKLOAD_PATH = "azurefunctions_setup/shared_azure_workload/azurefunctionsworkload.py"

def extract_execute_function(src_path):
    # Extract the execute_function logic from trace_func.py.
    with open(src_path, "r") as f:
        content = f.read()

    # Use regex to extract the execute_function definition and body
    match = re.search(r"def execute_function\(.*?\):.*?(?=def |\Z)", content, re.DOTALL)
    if not match:
        raise ValueError("execute_function() not found in trace_func.py")

    return match.group(0)

def inject_function(func_code, workload_path):
    # Inject or replace execute_function in azureworkload.py.
    with open(workload_path, "r") as f:
        content = f.read()

    # Check if execute_function already exists and replace it
    if "def execute_function" in content:
        updated_content = re.sub(
            r"def execute_function\(.*?\):.*?(?=def |\Z)",  # Match existing function
            func_code,  # Replace with new function code
            content,
            flags=re.DOTALL,
        )
    else:
        # Add execute_function at the end of the file
        updated_content = f"{content}\n\n{func_code}"

    # Write updated content back to the file
    with open(workload_path, "w") as f:
        f.write(updated_content)


def validate_injection(workload_path):
    # Validate that execute_function is present in azureworkload.py after injection.
    with open(workload_path, "r") as f:
        content = f.read()

    if "def execute_function" not in content:
        raise RuntimeError("Injection failed: execute_function() not found in azureworkload.py")


if __name__ == "__main__":
    try:
        # Extract execute_function from trace_func.py
        execute_function_code = extract_execute_function(TRACE_FUNC_PATH)

        # Inject the extracted function into azureworkload.py
        inject_function(execute_function_code, AZURE_WORKLOAD_PATH)

        # Validate the injection
        validate_injection(AZURE_WORKLOAD_PATH)

    except Exception as e:
        print(f"Error during workload injection: {e}")
        exit(1)
