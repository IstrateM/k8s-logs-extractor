# k8s-logs-extractor

A go tool to quickly extract all the logs in a cluster.

### commands

- **o** - Set the output files location
- **kc** - Set the kubeconfig directory
- **diff** - Enable creation of .diff files base on previous extracted logs

### example

`k8s-log-extractor --kc="/home/user/.kube/" --o="/home/user/cluster-logs/"`
`k8s-log-extractor --kc="/home/user/.kube/" --o="/home/user/cluster-logs/" --diff`