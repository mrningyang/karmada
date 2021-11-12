package defaultexplorer

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	workv1alpha2 "github.com/karmada-io/karmada/pkg/apis/work/v1alpha2"
	"github.com/karmada-io/karmada/pkg/util"
	"github.com/karmada-io/karmada/pkg/util/helper"
)

// replicaExplorer is the function that used to parse replica and requirements from object.
type replicaExplorer func(object runtime.Object) (int32, *workv1alpha2.ReplicaRequirements, error)

func getAllDefaultReplicaExplorer() map[schema.GroupVersionKind]replicaExplorer {
	explorers := make(map[schema.GroupVersionKind]replicaExplorer)
	explorers[appsv1.SchemeGroupVersion.WithKind(util.DeploymentKind)] = deployReplicaExplorer
	explorers[batchv1.SchemeGroupVersion.WithKind(util.JobKind)] = jobReplicaExplorer
	return explorers
}

func deployReplicaExplorer(object runtime.Object) (int32, *workv1alpha2.ReplicaRequirements, error) {
	unstructuredObj, ok := object.(*unstructured.Unstructured)
	if !ok {
		return 0, nil, fmt.Errorf("unexpected object type, requires unstructured")
	}

	deploy, err := helper.ConvertToDeployment(unstructuredObj)
	if err != nil {
		klog.Errorf("Failed to convert object(%s), err", object.GetObjectKind().GroupVersionKind().String(), err)
		return 0, nil, err
	}

	var replica int32
	if deploy.Spec.Replicas != nil {
		replica = *deploy.Spec.Replicas
	}
	requirement := helper.GenerateReplicaRequirements(&deploy.Spec.Template)

	return replica, requirement, nil
}

func jobReplicaExplorer(object runtime.Object) (int32, *workv1alpha2.ReplicaRequirements, error) {
	unstructuredObj, ok := object.(*unstructured.Unstructured)
	if !ok {
		return 0, nil, fmt.Errorf("unexpected object type, requires unstructured")
	}

	job, err := helper.ConvertToJob(unstructuredObj)
	if err != nil {
		klog.Errorf("Failed to convert object(%s), err", object.GetObjectKind().GroupVersionKind().String(), err)
		return 0, nil, err
	}

	var replica int32
	// parallelism might never be nil as the kube-apiserver will set it to 1 by default if not specified.
	if job.Spec.Parallelism != nil {
		replica = *job.Spec.Parallelism
	}
	requirement := helper.GenerateReplicaRequirements(&job.Spec.Template)

	return replica, requirement, nil
}