import time
import socket
import json
import azure.functions as func
import logging
from time import process_time_ns
import math
from psutil import virtual_memory
from numpy import empty, float32

# Global variable for hostname
hostname = socket.gethostname()

# Placeholder for `execute_function`, retrieve according to server/trace-func-py/trace_func.py
def execute_function(input, runTime, totalMem):
    startTime = process_time_ns()

    chunkSize = 2**10 # size of a kb or 1024
    totalMem = totalMem*(2**10) # convert Mb to kb
    memory = virtual_memory()
    used = (memory.total - memory.available) // chunkSize # convert to kb
    additional = max(1, (totalMem - used))
    array = empty(additional*chunkSize, dtype=float32) # make an uninitialized array of that size, uninitialized to keep it fast
    # convert to ns
    runTime = (runTime - 1)*(10**6) # -1 because it should be slighly bellow that runtime
    memoryIndex = 0
    while process_time_ns() - startTime < runTime:
        for i in range(0, chunkSize):
            sin_i = math.sin(i)
            cos_i = math.cos(i)
            sqrt_i = math.sqrt(i)
            array[memoryIndex + i] = sin_i
        memoryIndex = (memoryIndex + chunkSize) % additional*chunkSize
    return (process_time_ns() - startTime) // 1000

def main(req: func.HttpRequest) -> func.HttpResponse:
    logging.info("Processing request.")

    start_time = time.time()

    # Parse JSON request body
    try:
        req_body = req.get_json()
        logging.info(f"Request body: {req_body}")
    except ValueError:
        logging.error("Invalid JSON received.")
        return func.HttpResponse(
            json.dumps({"error": "Invalid JSON"}),
            status_code=400,
            mimetype="application/json"
        )

    runtime_milliseconds = req_body.get('RuntimeInMilliSec', 1000)
    memory_mebibytes = req_body.get('MemoryInMebiBytes', 128)

    logging.info(f"Runtime requested: {runtime_milliseconds} ms, Memory: {memory_mebibytes} MiB")

    # Directly call the execute_function
    duration = execute_function("",runtime_milliseconds,memory_mebibytes)
    result_msg = f"Workload completed in {duration} microseconds"

    # Prepare the response
    response = {
        "Status": "Success",
        "Function": req.url.split("/")[-1],
        "MachineName": hostname,
        "ExecutionTime": int((time.time() - start_time) * 1_000_000),  # Total time (includes HTTP, workload, and response prep)
        "DurationInMicroSec": duration,  # Time spent on the workload itself
        "MemoryUsageInKb": memory_mebibytes * 1024,
        "Message": result_msg
    }

    logging.info(f"Response: {response}")

    return func.HttpResponse(
        json.dumps(response),
        status_code=200,
        mimetype="application/json"
    )
