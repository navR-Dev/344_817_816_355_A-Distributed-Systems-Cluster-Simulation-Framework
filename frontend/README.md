
## TO-DO


# Test 1 { node stops , then comes back online after some time }

* node stops 
* all pods assigned to node go offline/pending 
* clear pods array in stopped NODE 
* clear nodeID in POD

* restart node
* pod scheduler should keep checking if a valid node exists
* if valid node exists assign pods to node

