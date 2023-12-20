// Copyright © 2021 Cisco Systems, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tests

import (
	"context"
	"fmt"
	"strconv"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/banzaicloud/koperator/api/v1beta1"
	kafkautils "github.com/banzaicloud/koperator/pkg/util/kafka"
	properties "github.com/banzaicloud/koperator/properties/pkg"
)

func expectDefaultBrokerSettingsForExternalListenerBinding(ctx context.Context, kafkaCluster *v1beta1.KafkaCluster, randomGenTestNumber uint64) {
	// Check Brokers
	for _, broker := range kafkaCluster.Spec.Brokers {
		broker := broker
		// expect ConfigMap
		configMap := corev1.ConfigMap{}
		Eventually(ctx, func() error {
			return k8sClient.Get(ctx, types.NamespacedName{
				Namespace: kafkaCluster.Namespace,
				Name:      fmt.Sprintf("%s-config-%d", kafkaCluster.Name, broker.Id),
			}, &configMap)
		}).Should(Succeed())

		Expect(configMap.Labels).To(HaveKeyWithValue(v1beta1.AppLabelKey, "kafka"))
		Expect(configMap.Labels).To(HaveKeyWithValue(v1beta1.KafkaCRLabelKey, kafkaCluster.Name))
		Expect(configMap.Labels).To(HaveKeyWithValue(v1beta1.BrokerIdLabelKey, strconv.Itoa(int(broker.Id))))

		brokerConfig, err := properties.NewFromString(configMap.Data["broker-config"])
		Expect(err).NotTo(HaveOccurred())
		advertisedListener, found := brokerConfig.Get("advertised.listeners")
		Expect(found).To(BeTrue())
		Expect(advertisedListener.Value()).To(Equal(fmt.Sprintf("TEST://external.az1.host.com:%d,CONTROLLER://kafkacluster-%d-%d.kafka-%d.svc.cluster.local:29093,INTERNAL://kafkacluster-%d-%d.kafka-%d.svc.cluster.local:29092",
			19090+broker.Id, randomGenTestNumber, broker.Id, randomGenTestNumber, randomGenTestNumber, broker.Id, randomGenTestNumber)))
		listeners, found := brokerConfig.Get("listeners")
		Expect(found).To(BeTrue())
		Expect(listeners.Value()).To(Equal("TEST://:9094,INTERNAL://:29092,CONTROLLER://:29093"))
		listenerSecMap, found := brokerConfig.Get(kafkautils.KafkaConfigListenerSecurityProtocolMap)
		Expect(found).To(BeTrue())
		Expect(listenerSecMap.Value()).To(Equal("TEST:PLAINTEXT,INTERNAL:PLAINTEXT,CONTROLLER:PLAINTEXT"))
		// check service
		service := corev1.Service{}
		Eventually(ctx, func() error {
			return k8sClient.Get(ctx, types.NamespacedName{
				Namespace: kafkaCluster.Namespace,
				Name:      fmt.Sprintf("%s-%d", kafkaCluster.Name, broker.Id),
			}, &service)
		}).Should(Succeed())

		Expect(service.Labels).To(HaveKeyWithValue(v1beta1.AppLabelKey, "kafka"))
		Expect(service.Labels).To(HaveKeyWithValue(v1beta1.KafkaCRLabelKey, kafkaCluster.Name))
		Expect(service.Labels).To(HaveKeyWithValue(v1beta1.BrokerIdLabelKey, strconv.Itoa(int(broker.Id))))

		Expect(service.Spec.Ports).To(ConsistOf(
			corev1.ServicePort{
				Name:       "tcp-internal",
				Protocol:   corev1.ProtocolTCP,
				Port:       29092,
				TargetPort: intstr.FromInt(29092),
			},
			corev1.ServicePort{
				Name:       "tcp-controller",
				Protocol:   corev1.ProtocolTCP,
				Port:       29093,
				TargetPort: intstr.FromInt(29093),
			},
			corev1.ServicePort{
				Name:       "tcp-test",
				Protocol:   corev1.ProtocolTCP,
				Port:       9094,
				TargetPort: intstr.FromInt(9094),
			},
			corev1.ServicePort{
				Name:       "metrics",
				Protocol:   corev1.ProtocolTCP,
				Port:       9020,
				TargetPort: intstr.FromInt(9020),
			}))
	}
}

func expectBrokerConfigmapForAz1ExternalListener(ctx context.Context, kafkaCluster *v1beta1.KafkaCluster, randomGenTestNumber uint64) {
	configMap := corev1.ConfigMap{}
	Eventually(ctx, func() error {
		return k8sClient.Get(ctx, types.NamespacedName{
			Namespace: kafkaCluster.Namespace,
			Name:      fmt.Sprintf("%s-config-%d", kafkaCluster.Name, 0),
		}, &configMap)
	}).Should(Succeed())

	brokerConfig, err := properties.NewFromString(configMap.Data["broker-config"])
	Expect(err).NotTo(HaveOccurred())
	advertisedListener, found := brokerConfig.Get("advertised.listeners")
	Expect(found).To(BeTrue())
	Expect(advertisedListener.Value()).To(Equal(fmt.Sprintf("TEST://external.az1.host.com:%d,CONTROLLER://kafkacluster-%d-%d.kafkaconfigtest-%d.svc.cluster.local:29093,INTERNAL://kafkacluster-%d-%d.kafkaconfigtest-%d.svc.cluster.local:29092",
		19090, randomGenTestNumber, 0, randomGenTestNumber, randomGenTestNumber, 0, randomGenTestNumber)))
}

func expectBrokerConfigmapForAz1ExternalListenerTls(kafkaCluster *v1beta1.KafkaCluster, randomGenTestNumber uint64) {
	configMap := corev1.ConfigMap{}
	Eventually(func() error {
		return k8sClient.Get(context.Background(), types.NamespacedName{
			Namespace: kafkaCluster.Namespace,
			Name:      fmt.Sprintf("%s-config-%d", kafkaCluster.Name, 0),
		}, &configMap)
	}).Should(Succeed())

	brokerConfig, err := properties.NewFromString(configMap.Data["broker-config"])
	Expect(err).NotTo(HaveOccurred())
	advertisedListener, found := brokerConfig.Get("advertised.listeners")
	Expect(found).To(BeTrue())
	Expect(advertisedListener.Value()).To(Equal(fmt.Sprintf("TEST://broker-0:%d,CONTROLLER://kafkaclustertls-%d-%d.kafkatlsconfigtest-%d.svc.cluster.local:29093,INTERNAL://kafkaclustertls-%d-%d.kafkatlsconfigtest-%d.svc.cluster.local:29092",
		29092, randomGenTestNumber, 0, randomGenTestNumber, randomGenTestNumber, 0, randomGenTestNumber)))
}

func expectBrokerConfigmapForAz2ExternalListener(ctx context.Context, kafkaCluster *v1beta1.KafkaCluster, randomGenTestNumber uint64) {
	configMap := corev1.ConfigMap{}
	Eventually(ctx, func() error {
		return k8sClient.Get(ctx, types.NamespacedName{
			Namespace: kafkaCluster.Namespace,
			Name:      fmt.Sprintf("%s-config-%d", kafkaCluster.Name, 1),
		}, &configMap)
	}).Should(Succeed())

	brokerConfig, err := properties.NewFromString(configMap.Data["broker-config"])
	Expect(err).NotTo(HaveOccurred())
	advertisedListener, found := brokerConfig.Get("advertised.listeners")
	Expect(found).To(BeTrue())
	Expect(advertisedListener.Value()).To(Equal(fmt.Sprintf("TEST://external.az2.host.com:%d,CONTROLLER://kafkacluster-%d-%d.kafkaconfigtest-%d.svc.cluster.local:29093,INTERNAL://kafkacluster-%d-%d.kafkaconfigtest-%d.svc.cluster.local:29092",
		19091, randomGenTestNumber, 1, randomGenTestNumber, randomGenTestNumber, 1, randomGenTestNumber)))

	configMap = corev1.ConfigMap{}
	Eventually(ctx, func() error {
		return k8sClient.Get(ctx, types.NamespacedName{
			Namespace: kafkaCluster.Namespace,
			Name:      fmt.Sprintf("%s-config-%d", kafkaCluster.Name, 2),
		}, &configMap)
	}).Should(Succeed())

	brokerConfig, err = properties.NewFromString(configMap.Data["broker-config"])
	Expect(err).NotTo(HaveOccurred())
	advertisedListener, found = brokerConfig.Get("advertised.listeners")
	Expect(found).To(BeTrue())
	Expect(advertisedListener.Value()).To(Equal(fmt.Sprintf("TEST://external.az2.host.com:%d,CONTROLLER://kafkacluster-%d-%d.kafkaconfigtest-%d.svc.cluster.local:29093,INTERNAL://kafkacluster-%d-%d.kafkaconfigtest-%d.svc.cluster.local:29092",
		19092, randomGenTestNumber, 2, randomGenTestNumber, randomGenTestNumber, 2, randomGenTestNumber)))
}

// func expectBrokerConfigmapForAz2ExternalListenerTls(kafkaCluster *v1beta1.KafkaCluster, randomGenTestNumber uint64) {
// 	configMap := corev1.ConfigMap{}
// 	Eventually(func() error {
// 		return k8sClient.Get(context.Background(), types.NamespacedName{
// 			Namespace: kafkaCluster.Namespace,
// 			Name:      fmt.Sprintf("%s-config-%d", kafkaCluster.Name, 1),
// 		}, &configMap)
// 	}).Should(Succeed())

// 	brokerConfig, err := properties.NewFromString(configMap.Data["broker-config"])
// 	Expect(err).NotTo(HaveOccurred())
// 	advertisedListener, found := brokerConfig.Get("advertised.listeners")
// 	Expect(found).To(BeTrue())
// 	Expect(advertisedListener.Value()).To(Equal(fmt.Sprintf("TEST://broker-1:%d,CONTROLLER://kafkaclustertls-%d-%d.kafkatlsconfigtest-%d.svc.cluster.local:29093,INTERNAL://kafkaclustertls-%d-%d.kafkatlsconfigtest-%d.svc.cluster.local:29092",
// 		29092, randomGenTestNumber, 1, randomGenTestNumber, randomGenTestNumber, 1, randomGenTestNumber)))

// 	configMap = corev1.ConfigMap{}
// 	Eventually(func() error {
// 		return k8sClient.Get(context.Background(), types.NamespacedName{
// 			Namespace: kafkaCluster.Namespace,
// 			Name:      fmt.Sprintf("%s-config-%d", kafkaCluster.Name, 2),
// 		}, &configMap)
// 	}).Should(Succeed())

// 	brokerConfig, err = properties.NewFromString(configMap.Data["broker-config"])
// 	Expect(err).NotTo(HaveOccurred())
// 	advertisedListener, found = brokerConfig.Get("advertised.listeners")
// 	Expect(found).To(BeTrue())
// 	Expect(advertisedListener.Value()).To(Equal(fmt.Sprintf("TEST://broker-2:%d,CONTROLLER://kafkaclustertls-%d-%d.kafkatlsconfigtest-%d.svc.cluster.local:29093,INTERNAL://kafkaclustertls-%d-%d.kafkatlsconfigtest-%d.svc.cluster.local:29092",
// 		29092, randomGenTestNumber, 2, randomGenTestNumber, randomGenTestNumber, 2, randomGenTestNumber)))
// }

func expectBrokerConfigmapForAz2ExternalListenerTls(kafkaCluster *v1beta1.KafkaCluster, randomGenTestNumber uint64) {
	configMap := corev1.ConfigMap{}
	Eventually(func() error {
		return k8sClient.Get(context.Background(), types.NamespacedName{
			Namespace: kafkaCluster.Namespace,
			Name:      fmt.Sprintf("%s-config-%d", kafkaCluster.Name, 1),
		}, &configMap)
	}).Should(Succeed())

	brokerConfig, err := properties.NewFromString(configMap.Data["broker-config"])
	Expect(err).NotTo(HaveOccurred())
	advertisedListener, found := brokerConfig.Get("advertised.listeners")
	Expect(found).To(BeTrue())
	Expect(advertisedListener.Value()).To(Equal(fmt.Sprintf("CONTROLLER://kafkaclustertls-%d-%d.kafkatlsconfigtest-%d.svc.cluster.local:29093,INTERNAL://kafkaclustertls-%d-%d.kafkatlsconfigtest-%d.svc.cluster.local:29092,TEST://broker-1:%d",
		randomGenTestNumber, 1, randomGenTestNumber, randomGenTestNumber, 1, randomGenTestNumber, 29092)))

	configMap = corev1.ConfigMap{}
	Eventually(func() error {
		return k8sClient.Get(context.Background(), types.NamespacedName{
			Namespace: kafkaCluster.Namespace,
			Name:      fmt.Sprintf("%s-config-%d", kafkaCluster.Name, 2),
		}, &configMap)
	}).Should(Succeed())

	brokerConfig, err = properties.NewFromString(configMap.Data["broker-config"])
	Expect(err).NotTo(HaveOccurred())
	advertisedListener, found = brokerConfig.Get("advertised.listeners")
	Expect(found).To(BeTrue())
	Expect(advertisedListener.Value()).To(Equal(fmt.Sprintf("CONTROLLER://kafkaclustertls-%d-%d.kafkatlsconfigtest-%d.svc.cluster.local:29093,INTERNAL://kafkaclustertls-%d-%d.kafkatlsconfigtest-%d.svc.cluster.local:29092,TEST://broker-2:%d",
		randomGenTestNumber, 2, randomGenTestNumber, randomGenTestNumber, 2, randomGenTestNumber, 29092)))
}
