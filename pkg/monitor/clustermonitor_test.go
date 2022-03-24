// Copyright Contributors to the Open Cluster Management project

package monitor

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"testing"

	sanitize "github.com/kennygrant/sanitize"
	"github.com/stolostron/insights-client/pkg/types"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

func unmarshalFile(filepath string, resourceType interface{}, t *testing.T) {
	// open given filepath string
	rawBytes, err := ioutil.ReadFile("../../test-data/" + sanitize.Name(filepath))
	if err != nil {
		t.Fatal("Unable to read test data", err)
	}

	// unmarshal file into given resource type
	err = json.Unmarshal(rawBytes, resourceType)
	if err != nil {
		t.Fatalf("Unable to unmarshal json to type %T %s", resourceType, err)
	}
}

func Test_addCluster(t *testing.T) {
	monitor := NewClusterMonitor()
	managedCluster := clusterv1.ManagedCluster{}
	unmarshalFile("managed-cluster.json", &managedCluster, t)
	monitor.addCluster(&managedCluster)

	assert.Equal(t, types.ManagedClusterInfo{Namespace: "managed-cluster", ClusterID: "323a00cd-428a-49fb-80ab-201d2a5d3050"}, monitor.ManagedClusterInfo[0], "Test Add ManagedCluster (MangedClusterInfo): local-cluster")
	assert.Equal(t, map[string]bool{"323a00cd-428a-49fb-80ab-201d2a5d3050": true}, monitor.ClusterNeedsCCX, "Test Add ManagedCluster (ClusterNeedsCCX): local-cluster")

}

func Test_addCluster_nonOpenshift(t *testing.T) {
	monitor := NewClusterMonitor()
	monitor.ManagedClusterInfo = []types.ManagedClusterInfo{}
	monitor.ClusterNeedsCCX = map[string]bool{}
	managedCluster := clusterv1.ManagedCluster{}
	unmarshalFile("managed-cluster-nonopenshift.json", &managedCluster, t)
	monitor.addCluster(&managedCluster)

	assert.Equal(t, types.ManagedClusterInfo{Namespace: "managed-cluster", ClusterID: "local-cluster-non-openshift"}, monitor.ManagedClusterInfo[0], "Test Add ManagedCluster: local-cluster-non-openshift")
	assert.Equal(t, map[string]bool{"local-cluster-non-openshift": false}, monitor.ClusterNeedsCCX, "Test Add ManagedCluster (ClusterNeedsCCX): local-cluster-non-openshift")

}

func Test_updateCluster(t *testing.T) {
	monitor := NewClusterMonitor()
	monitor.ManagedClusterInfo = []types.ManagedClusterInfo{{Namespace: "managed-cluster", ClusterID: "123a00cd-428a-49fb-80ab-201d2a5d3050"}}
	monitor.ClusterNeedsCCX = map[string]bool{"123a00cd-428a-49fb-80ab-201d2a5d3050": true}
	managedCluster := clusterv1.ManagedCluster{}
	unmarshalFile("managed-cluster.json", &managedCluster, t)

	monitor.updateCluster(&managedCluster)

	assert.Equal(t, types.ManagedClusterInfo{Namespace: "managed-cluster", ClusterID: "323a00cd-428a-49fb-80ab-201d2a5d3050"}, monitor.ManagedClusterInfo[0], "Test Add ManagedCluster: local-cluster")
	assert.Equal(t, map[string]bool{"323a00cd-428a-49fb-80ab-201d2a5d3050": true}, monitor.ClusterNeedsCCX, "Test Update ManagedCluster (ClusterNeedsCCX): local-cluster")

}

func Test_updateCluster_nonOpenshift(t *testing.T) {
	monitor := NewClusterMonitor()
	monitor.ManagedClusterInfo = []types.ManagedClusterInfo{{Namespace: "managed-cluster", ClusterID: "test-cluster-non-openshift"}}
	monitor.ClusterNeedsCCX = map[string]bool{"test-cluster-non-openshift": false}
	managedCluster := clusterv1.ManagedCluster{}
	unmarshalFile("managed-cluster-nonopenshift.json", &managedCluster, t)

	monitor.updateCluster(&managedCluster)

	assert.Equal(t, types.ManagedClusterInfo{Namespace: "managed-cluster", ClusterID: "local-cluster-non-openshift"}, monitor.ManagedClusterInfo[0], "Test Update ManagedCluster: local-cluster-non-openshift")
	assert.Equal(t, map[string]bool{"local-cluster-non-openshift": false}, monitor.ClusterNeedsCCX, "Test Update ManagedCluster (ClusterNeedsCCX): local-cluster-non-openshift")

}

func Test_deleteCluster(t *testing.T) {
	monitor := NewClusterMonitor()
	monitor.ManagedClusterInfo = []types.ManagedClusterInfo{{Namespace: "managed-cluster", ClusterID: "323a00cd-428a-49fb-80ab-201d2a5d3050"}}
	monitor.ClusterNeedsCCX = map[string]bool{"323a00cd-428a-49fb-80ab-201d2a5d3050": true}

	managedCluster := clusterv1.ManagedCluster{}
	unmarshalFile("managed-cluster.json", &managedCluster, t)

	monitor.deleteCluster(&managedCluster)

	assert.Equal(t, []types.ManagedClusterInfo{}, monitor.ManagedClusterInfo, "Test Delete ManagedCluster: managed-cluster")
	assert.Equal(t, map[string]bool{}, monitor.ClusterNeedsCCX, "Test Delete ManagedCluster: managed-cluster (ClusterNeedsCCX)")

}

func Test_deleteCluster_nonOpenshift(t *testing.T) {
	monitor := NewClusterMonitor()
	monitor.ManagedClusterInfo = []types.ManagedClusterInfo{{Namespace: "managed-cluster", ClusterID: "local-cluster-non-openshift"}}
	monitor.ClusterNeedsCCX = map[string]bool{"local-cluster-non-openshift": false}

	managedCluster := clusterv1.ManagedCluster{}
	unmarshalFile("managed-cluster-nonopenshift.json", &managedCluster, t)

	monitor.deleteCluster(&managedCluster)

	assert.Equal(t, []types.ManagedClusterInfo{}, monitor.ManagedClusterInfo, "Test Delete ManagedCluster: local-cluster-non-openshift")
	assert.Equal(t, map[string]bool{}, monitor.ClusterNeedsCCX, "Test Delete ManagedCluster: local-cluster-non-openshift (ClusterNeedsCCX)")

}

func Test_isClustermissing(t *testing.T) {
	resultFalse := isClusterMissing(nil)
	assert.Equal(t, false, resultFalse, "Test isClusterMissing - false")

	err := errors.New("could not find the requested resource")
	resultTrue := isClusterMissing(err)
	assert.Equal(t, true, resultTrue, "Test isClusterMissing - true")
}

func Test_AddLocalCluster(t *testing.T) {
	monitor := NewClusterMonitor()
	versionU := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "clusterversions",
			"metadata": map[string]interface{}{
				"namespace": "namespace",
				"name":      "version",
			},
			"spec": map[string]interface{}{
				"channel":   "stable - 4.6",
				"clusterID": "58bd7441-812e-4fab-9aa6-eec452059c59",
				"upstream":  "https://api.openshift.com/api/upgrades_info/v1/graph",
			},
		},
	}
	monitor.AddLocalCluster(versionU)
	assert.Equal(t, types.ManagedClusterInfo{Namespace: "local-cluster", ClusterID: "58bd7441-812e-4fab-9aa6-eec452059c59"}, monitor.ManagedClusterInfo[0], "Test AddLocalCluster: local-cluster")

}

func Test_GetLocalCluster(t *testing.T) {
	monitor := NewClusterMonitor()
	versionU := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "clusterversions",
			"metadata": map[string]interface{}{
				"namespace": "namespace",
				"name":      "version",
			},
			"spec": map[string]interface{}{
				"channel":   "stable - 4.6",
				"clusterID": "58bd7441-812e-4fab-9aa6-eec452059c59",
				"upstream":  "https://api.openshift.com/api/upgrades_info/v1/graph",
			},
		},
	}
	monitor.AddLocalCluster(versionU)
	assert.Equal(t, "58bd7441-812e-4fab-9aa6-eec452059c59", monitor.GetLocalCluster(), "Test GetLocalCluster: local-cluster")

}
