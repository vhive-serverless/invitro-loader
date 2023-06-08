# Loader

A load generator for benchmarking serverless systems.

## Prerequisites

The experiments require a server-grade node running Linux (tested on Ubuntu 20, Intel Xeon). On CloudLab, one
can choose the APT cluster `d430` node.

## Create a cluster

First, configure `script/setup.cfg`. You can specify there which vHive branch to use, loader branch, operation mode
(sandbox type), maximum number of pods per node, and the Github token. All these configurations are mandatory.
We currently support the following modes: containerd (`container`), Firecracker (`firecracker`), and Firecracker with
snapshots (`firecracker_snapshots`).
The token needs `repo` and `admin:public_key` permissions and will be used for adding SSH key to the user's account for
purpose of cloning Github private repositories.
Loader will be cloned on every node specified as argument of the cluster create script. The same holds for Kubernetes
API server certificate.

* To create a multi-node cluster, specify the node addresses as the arguments and run the following command:

```bash
$ bash ./scripts/setup/create_multinode.sh <master_node@IP> <worker_node@IP> ...
```

The loader should be running on a separate node that is part of the Kubernetes cluster. Do not collocate master and
worker node components where the loader is located for performance reasons. Make sure you taint the node where loader is
located prior to running any experiment.

* Single-node cluster (experimental)

```bash
$ bash ./scripts/setup/create_singlenode_container.sh <node@IP>
```

This mode is only for debugging purposes, and there is no guarantees of isolation between the loader and the master-node
components.

### Check cluster health (on the master node)

Once the setup scripts are finished, we need to check if they have completed their jobs fully, because, e.g., there
might be race conditions where a few nodes are unavailable.

* First, log out the status of the control plane components on the master node and monitoring deployments by running the
  following command:

```bash
$ bash ./scripts/util/log_kn_status.sh
```

* If you see everything is `Running`, check if the cluster capacity is stretched to the desired capacity by running the
  following script:

```bash
$ bash ./scripts/util/check_node_capacity.sh
```

If you want to check
the [pod CIDR range](https://www.ibm.com/docs/en/cloud-private/3.1.2?topic=networking-kubernetes-network-model), run the
following

```bash
$ bash ./scripts/util/get_pod_cidr.sh
```

* Next, try deploying a function (`myfunc`) with the desired `scale` (i.e., the number of instances that must be active
  at all times).

```bash
$ bash ./scripts/util/set_function_scale.sh <scale>
```

* One should verify that the system was able to start the requested number of function instances, by using the following
  command.

```bash
$ kubectl -n default get podautoscalers
```

## Tune the timing for the benchmark function

Before start any experiments, the timing of the benchmark function should be tuned so that it consumes the required
service time more precisely.

First, run the following command to deploy the timing benchmark that yields the number of execution units* the function
needs to run given a required service time.

* The execution unit is approximately 100 `SQRTSD` x86 instructions.

```bash
$ kubectl apply -f server/benchmark/timing.yaml
```

Then, monitor and collect the `ITERATIONS_MULTIPLIER` from the job logs as follows:

```bash
$ watch kubectl logs timing
```

Finally, set the `ITERATIONS_MULTIPLIER` in the function template `workloads/$SANDBOX_TYPE/trace_func_go.yaml` to the
value previously collected.

To account for difference in CPU performance set `ITERATIONS_MULTIPLIER=102` if using
Cloudlab `xl170` or `d430` machines. (Date of measurement: 18-Oct-2022)

## Single execution

To run load generator use the following command:

```bash
$ go run cmd/loader.go --config cmd/config.json
```

Additionally, one can specify log verbosity argument as `--verbosity [info, debug, trace]`. The default value is `info`.

For to configure the workload for load generator, please refer to `docs/configuration.md`.

There are a couple of constants that should not be exposed to the users. They can be examined and changed
in `pkg/common/constants.go`.

## Build the image for a synthetic function

The reason for existence of Firecracker and container version is because of different ports for gRPC server. Firecracker
uses port 50051, whereas containers are accessible via port 80.

Synthetic function runs for a given time and allocates the required amount of memory:

* `trace-firecracker` - for Firecracker
* `trace-container` - for regular containers

For testing cold start performance:

* `empty-firecracker` - for Firecracker - returns immediately
* `empty-container` - for regular containers - returns immediately

```bash
$ make build <trace-firecracker|trace-container|empty-firecracker|empty-container>
```

## Clean up between runs

```bash
$ make clean
```

## Running the Experiment Driver

Within the tools/driver folder, Configure the driverConfig.json file based on your username on Cloudlab, 
the address of the loader node that you are running the experiment on and any experiment parameters you'd like to set.
Note that you need to have the relevant trace files on the machine running the experiment driver, they will then get
transferred to the loader node.  
Then run the following command to launch an experiment:
```bash
$ go run experiment_driver.go -c driverConfig.json
```
For more details take a look at the README in the tools/driver folder.

---

For more options, please see the `Makefile`.

## Using OpenWhisk

For instructions on how to use the loader with OpenWhisk go to `openwhisk_setup/README.md`.

