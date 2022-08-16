//=======================================================================
// Copyright (c) 2022 Aarna Networks, Inc.
// All rights reserved.
// ======================================================================
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//           http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ========================================================================

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	edgemetricscollectorv1alpha1 "edgemetricscollector/pkg/apis/edgemetricscollector/v1alpha1"
	versioned "edgemetricscollector/pkg/generated/clientset/versioned"
	internalinterfaces "edgemetricscollector/pkg/generated/informers/externalversions/internalinterfaces"
	v1alpha1 "edgemetricscollector/pkg/generated/listers/edgemetricscollector/v1alpha1"
	time "time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// MetricsCollectorInformer provides access to a shared informer and lister for
// MetricsCollectors.
type MetricsCollectorInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.MetricsCollectorLister
}

type metricsCollectorInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewMetricsCollectorInformer constructs a new informer for MetricsCollector type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewMetricsCollectorInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredMetricsCollectorInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredMetricsCollectorInformer constructs a new informer for MetricsCollector type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredMetricsCollectorInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.EdgemetricscollectorV1alpha1().MetricsCollectors(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.EdgemetricscollectorV1alpha1().MetricsCollectors(namespace).Watch(context.TODO(), options)
			},
		},
		&edgemetricscollectorv1alpha1.MetricsCollector{},
		resyncPeriod,
		indexers,
	)
}

func (f *metricsCollectorInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredMetricsCollectorInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *metricsCollectorInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&edgemetricscollectorv1alpha1.MetricsCollector{}, f.defaultInformer)
}

func (f *metricsCollectorInformer) Lister() v1alpha1.MetricsCollectorLister {
	return v1alpha1.NewMetricsCollectorLister(f.Informer().GetIndexer())
}
