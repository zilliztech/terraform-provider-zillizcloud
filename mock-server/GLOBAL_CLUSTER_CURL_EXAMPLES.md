# Global Cluster Curl Examples

These examples target the local mock server.

Start the server first:

```bash
cd mock-server
go run main.go
```

Use this base URL in another terminal:

```bash
BASE_URL="http://localhost:8080/v2"
PROJECT_ID="proj-ebc5ac7f430702aec8c57b"
```

## Create Global Cluster

```bash
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/globalClusters/create" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "globalClusterName": "my-global-cluster",
    "projectId": "'"$PROJECT_ID"'",
    "cuType": "Performance-optimized",
    "cuSize": 4,
    "primaryCluster": {
      "clusterName": "primary-cluster",
      "regionId": "aws-us-west-2"
    },
    "secondaryClusters": [
      {
        "clusterName": "secondary-cluster-eu",
        "regionId": "aws-eu-west-1"
      }
    ]
  }')

echo "$CREATE_RESPONSE" | jq .
GLOBAL_CLUSTER_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.globalClusterId')
echo "GLOBAL_CLUSTER_ID=$GLOBAL_CLUSTER_ID"
```

## List Global Clusters

```bash
curl -s -X GET "$BASE_URL/globalClusters?projectId=$PROJECT_ID&currentPage=1&pageSize=10" \
  -H "Accept: application/json" | jq .
```

## Describe Global Cluster

```bash
DESCRIBE_RESPONSE=$(curl -s -X GET "$BASE_URL/globalClusters/$GLOBAL_CLUSTER_ID" \
  -H "Accept: application/json")

echo "$DESCRIBE_RESPONSE" | jq .
PRIMARY_CLUSTER_ID=$(echo "$DESCRIBE_RESPONSE" | jq -r '.data.clusters[] | select(.role == "PRIMARY") | .clusterId')
SECONDARY_CLUSTER_ID=$(echo "$DESCRIBE_RESPONSE" | jq -r '.data.clusters[] | select(.role == "SECONDARY") | .clusterId' | head -n 1)
echo "PRIMARY_CLUSTER_ID=$PRIMARY_CLUSTER_ID"
echo "SECONDARY_CLUSTER_ID=$SECONDARY_CLUSTER_ID"
```

## Describe Member Clusters

The mock server writes member clusters into the existing cluster store. Global members include `globalClusterMeta`.

```bash
curl -s -X GET "$BASE_URL/clusters/$PRIMARY_CLUSTER_ID" \
  -H "Accept: application/json" | jq .

curl -s -X GET "$BASE_URL/clusters/$SECONDARY_CLUSTER_ID" \
  -H "Accept: application/json" | jq .
```

## Modify Global Cluster CU

Manual CU scaling:

```bash
curl -s -X POST "$BASE_URL/globalClusters/$GLOBAL_CLUSTER_ID/modifyCU" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "cuSize": 8
  }' | jq .
```

Dynamic CU autoscaling:

```bash
curl -s -X POST "$BASE_URL/globalClusters/$GLOBAL_CLUSTER_ID/modifyCU" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "autoscaling": {
      "cu": {
        "min": 4,
        "max": 16
      }
    }
  }' | jq .
```

Scheduled CU autoscaling:

```bash
curl -s -X POST "$BASE_URL/globalClusters/$GLOBAL_CLUSTER_ID/modifyCU" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "autoscaling": {
      "cu": {
        "schedules": [
          {
            "cron": "10 0 0 0 0 ?",
            "target": 8
          }
        ]
      }
    }
  }' | jq .
```

## Add Secondary Cluster

```bash
curl -s -X POST "$BASE_URL/globalClusters/$GLOBAL_CLUSTER_ID/secondaryClusters" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "secondaryClusters": [
      {
        "clusterName": "secondary-cluster-ap",
        "regionId": "aws-ap-southeast-1"
      }
    ]
  }' | jq .
```

Refresh the topology and capture the newly added secondary cluster ID:

```bash
DESCRIBE_RESPONSE=$(curl -s -X GET "$BASE_URL/globalClusters/$GLOBAL_CLUSTER_ID" \
  -H "Accept: application/json")

echo "$DESCRIBE_RESPONSE" | jq .
NEW_SECONDARY_CLUSTER_ID=$(echo "$DESCRIBE_RESPONSE" | jq -r '.data.clusters[] | select(.clusterName == "secondary-cluster-ap") | .clusterId')
echo "NEW_SECONDARY_CLUSTER_ID=$NEW_SECONDARY_CLUSTER_ID"
```

## Delete Secondary Cluster

```bash
curl -s -X DELETE "$BASE_URL/globalClusters/$GLOBAL_CLUSTER_ID/secondaryClusters/$SECONDARY_CLUSTER_ID" \
  -H "Accept: application/json" | jq .
```

## Remove Global Endpoint

This endpoint is available only after all secondary clusters have been deleted. It removes the global endpoint and keeps the remaining primary cluster as a regular dedicated cluster.

If you added an extra secondary cluster earlier, delete it first:

```bash
curl -s -X DELETE "$BASE_URL/globalClusters/$GLOBAL_CLUSTER_ID/secondaryClusters/$NEW_SECONDARY_CLUSTER_ID" \
  -H "Accept: application/json" | jq .
```

```bash
curl -s -X POST "$BASE_URL/globalClusters/$GLOBAL_CLUSTER_ID/removeGlobalEndpoint" \
  -H "Accept: application/json" | jq .
```

Verify the global cluster was removed and the former primary remains:

```bash
curl -s -X GET "$BASE_URL/globalClusters/$GLOBAL_CLUSTER_ID" \
  -H "Accept: application/json" | jq .

curl -s -X GET "$BASE_URL/clusters/$PRIMARY_CLUSTER_ID" \
  -H "Accept: application/json" | jq .
```
