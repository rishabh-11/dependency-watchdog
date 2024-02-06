// Copyright 2022 SAP SE or an SAP affiliate company
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prober

import (
	"time"

	papi "github.com/gardener/dependency-watchdog/api/prober"
	"github.com/gardener/dependency-watchdog/internal/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// DefaultProbeInterval is the default duration representing the interval for a probe.
	DefaultProbeInterval = 10 * time.Second
	// DefaultProbeInitialDelay is the default duration representing an initial delay to start a probe.
	DefaultProbeInitialDelay = 30 * time.Second
	// DefaultScaleInitialDelay is the default duration representing an initial delay to start to scale up a kubernetes resource.
	DefaultScaleInitialDelay = 0 * time.Second
	// DefaultProbeTimeout is the default duration representing total timeout for a probe to complete.
	DefaultProbeTimeout = 30 * time.Second
	// DefaultBackoffJitterFactor is the default jitter value with which successive probe runs are scheduled.
	DefaultBackoffJitterFactor = 0.2
	// DefaultScaleUpdateTimeout is the default duration representing a timeout for the scale operation to complete.
	DefaultScaleUpdateTimeout = 30 * time.Second
	// DefaultNodeLeaseFailureFraction is used to determine the maximum number of node leases that can be expired for a node lease probe to succeed.
	// Eg:- 1. numberOfOwnedLeases = 10, numberOfExpiredLeases = 6.
	// 		   numberOfExpiredLeases/numberOfOwnedLeases = 0.6, which is >= DefaultNodeLeaseFailureFraction and so the lease probe will fail.
	//		2. numberOfOwnedLeases = 10, numberOfExpiredLeases = 5.
	//	 	   numberOfExpiredLeases/numberOfOwnedLeases = 0.5, which is < DefaultNodeLeaseFailureFraction and so the lease probe will succeed.
	DefaultNodeLeaseFailureFraction = 0.60
)

// LoadConfig reads the prober configuration from a file, unmarshalls it, fills in the default values and
// validates the unmarshalled configuration If all validations pass it will return papi.Config else it will return an error.
func LoadConfig(file string, scheme *runtime.Scheme) (*papi.Config, error) {
	config, err := util.ReadAndUnmarshall[papi.Config](file)
	if err != nil {
		return nil, err
	}
	fillDefaultValues(config)
	err = validate(config, scheme)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func validate(c *papi.Config, scheme *runtime.Scheme) error {
	v := new(util.Validator)
	// Check the mandatory config parameters for which a default will not be set
	v.MustNotBeEmpty("KubeConfigSecretName", c.KubeConfigSecretName)
	v.MustNotBeZeroDuration("KCMNodeMonitorGraceDuration", c.KCMNodeMonitorGraceDuration)
	v.MustNotBeEmpty("ScaleResourceInfos", c.DependentResourceInfos)
	for _, resInfo := range c.DependentResourceInfos {
		v.ResourceRefMustBeValid(resInfo.Ref, scheme)
		v.MustNotBeNil("scaleUp", resInfo.ScaleUpInfo)
		v.MustNotBeNil("scaleDown", resInfo.ScaleDownInfo)
	}
	if v.Error != nil {
		return v.Error
	}
	return nil
}

func fillDefaultValues(c *papi.Config) {
	c.ProbeInterval = util.GetValOrDefault(c.ProbeInterval, metav1.Duration{Duration: DefaultProbeInterval})
	c.InitialDelay = util.GetValOrDefault(c.InitialDelay, metav1.Duration{Duration: DefaultProbeInitialDelay})
	c.ProbeTimeout = util.GetValOrDefault(c.ProbeTimeout, metav1.Duration{Duration: DefaultProbeTimeout})
	c.BackoffJitterFactor = util.GetValOrDefault(c.BackoffJitterFactor, DefaultBackoffJitterFactor)
	c.NodeLeaseFailureFraction = util.GetValOrDefault(c.NodeLeaseFailureFraction, DefaultNodeLeaseFailureFraction)
	fillDefaultValuesForResourceInfos(c.DependentResourceInfos)
}

func fillDefaultValuesForResourceInfos(resourceInfos []papi.DependentResourceInfo) {
	for _, resInfo := range resourceInfos {
		fillDefaultValuesForScaleInfo(resInfo.ScaleUpInfo)
		fillDefaultValuesForScaleInfo(resInfo.ScaleDownInfo)
	}
}

func fillDefaultValuesForScaleInfo(scaleInfo *papi.ScaleInfo) {
	if scaleInfo != nil {
		scaleInfo.Timeout = util.GetValOrDefault(scaleInfo.Timeout, metav1.Duration{Duration: DefaultScaleUpdateTimeout})
		scaleInfo.InitialDelay = util.GetValOrDefault(scaleInfo.InitialDelay, metav1.Duration{Duration: DefaultScaleInitialDelay})
	}
}
