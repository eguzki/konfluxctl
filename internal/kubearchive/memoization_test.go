/*
Copyright 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package kubearchive

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"

	applicationapi "github.com/konflux-ci/application-api/api/v1alpha1"
	konfluxapi "github.com/konflux-ci/release-service/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Memoization", func() {
	var (
		ctx       context.Context
		namespace string
	)

	BeforeEach(func() {
		ctx = context.Background()
		namespace = "test-namespace"
	})

	Describe("GetSnapshot memoization", func() {
		It("caches successful results and avoids redundant requests", func() {
			var requestCount int32
			snapshotName := "test-snapshot"

			expectedSnapshot := applicationapi.Snapshot{
				ObjectMeta: metav1.ObjectMeta{
					Name:      snapshotName,
					Namespace: namespace,
				},
				Spec: applicationapi.SnapshotSpec{
					Application: "test-app",
				},
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&requestCount, 1)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				Expect(json.NewEncoder(w).Encode(expectedSnapshot)).NotTo(HaveOccurred())
			}))
			defer server.Close()

			client := &KubeArchiveHTTPClient{
				httpClient:          server.Client(),
				kubeArchiveHostname: server.URL[7:],
				scheme:              "http",
				releasesCache:       make(map[string]konfluxapi.ReleaseList),
				snapshotsCache:      make(map[string]applicationapi.Snapshot),
			}

			// First call - should hit the server
			snapshot1, err1 := client.GetSnapshot(ctx, namespace, snapshotName)
			Expect(err1).NotTo(HaveOccurred())
			Expect(snapshot1.Name).To(Equal(snapshotName))
			Expect(atomic.LoadInt32(&requestCount)).To(Equal(int32(1)))

			// Second call - should use cache
			snapshot2, err2 := client.GetSnapshot(ctx, namespace, snapshotName)
			Expect(err2).NotTo(HaveOccurred())
			Expect(snapshot2.Name).To(Equal(snapshotName))
			Expect(atomic.LoadInt32(&requestCount)).To(Equal(int32(1))) // Still 1, not 2

			// Third call - should still use cache
			snapshot3, err3 := client.GetSnapshot(ctx, namespace, snapshotName)
			Expect(err3).NotTo(HaveOccurred())
			Expect(snapshot3.Name).To(Equal(snapshotName))
			Expect(atomic.LoadInt32(&requestCount)).To(Equal(int32(1))) // Still 1
		})

		It("distinguishes between different snapshots", func() {
			var requestCount int32

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count := atomic.AddInt32(&requestCount, 1)
				snapshot := applicationapi.Snapshot{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("snapshot-%d", count),
						Namespace: namespace,
					},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				Expect(json.NewEncoder(w).Encode(snapshot)).NotTo(HaveOccurred())
			}))
			defer server.Close()

			client := &KubeArchiveHTTPClient{
				httpClient:          server.Client(),
				kubeArchiveHostname: server.URL[7:],
				scheme:              "http",
				releasesCache:       make(map[string]konfluxapi.ReleaseList),
				snapshotsCache:      make(map[string]applicationapi.Snapshot),
			}

			// Different snapshots should each hit the server once
			_, err1 := client.GetSnapshot(ctx, namespace, "snapshot-1")
			Expect(err1).NotTo(HaveOccurred())

			_, err2 := client.GetSnapshot(ctx, namespace, "snapshot-2")
			Expect(err2).NotTo(HaveOccurred())

			// But repeating same snapshot uses cache
			_, err3 := client.GetSnapshot(ctx, namespace, "snapshot-1")
			Expect(err3).NotTo(HaveOccurred())

			Expect(atomic.LoadInt32(&requestCount)).To(Equal(int32(2))) // Only 2 requests, not 3
		})
	})

	Describe("getReleases memoization", func() {
		It("caches successful results and avoids redundant requests", func() {
			var requestCount int32

			releases := []konfluxapi.Release{
				{ObjectMeta: metav1.ObjectMeta{Name: "release-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "release-2"}},
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&requestCount, 1)
				releaseList := konfluxapi.ReleaseList{Items: releases}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				Expect(json.NewEncoder(w).Encode(releaseList)).NotTo(HaveOccurred())
			}))
			defer server.Close()

			client := &KubeArchiveHTTPClient{
				httpClient:          server.Client(),
				kubeArchiveHostname: server.URL[7:],
				scheme:              "http",
				releasesCache:       make(map[string]konfluxapi.ReleaseList),
				snapshotsCache:      make(map[string]applicationapi.Snapshot),
			}

			// First call - should hit the server
			list1, err1 := client.getReleases(ctx, namespace, "")
			Expect(err1).NotTo(HaveOccurred())
			Expect(list1.Items).To(HaveLen(2))
			Expect(atomic.LoadInt32(&requestCount)).To(Equal(int32(1)))

			// Second call - should use cache
			list2, err2 := client.getReleases(ctx, namespace, "")
			Expect(err2).NotTo(HaveOccurred())
			Expect(list2.Items).To(HaveLen(2))
			Expect(atomic.LoadInt32(&requestCount)).To(Equal(int32(1))) // Still 1
		})

		It("distinguishes between different continue tokens", func() {
			var requestCount int32

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count := atomic.AddInt32(&requestCount, 1)
				continueToken := r.URL.Query().Get("continue")

				var releaseList konfluxapi.ReleaseList
				if continueToken == "" {
					releaseList = konfluxapi.ReleaseList{
						ListMeta: metav1.ListMeta{Continue: "token-page-2"},
						Items: []konfluxapi.Release{
							{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("release-page1-%d", count)}},
						},
					}
				} else {
					releaseList = konfluxapi.ReleaseList{
						Items: []konfluxapi.Release{
							{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("release-page2-%d", count)}},
						},
					}
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				Expect(json.NewEncoder(w).Encode(releaseList)).NotTo(HaveOccurred())
			}))
			defer server.Close()

			client := &KubeArchiveHTTPClient{
				httpClient:          server.Client(),
				kubeArchiveHostname: server.URL[7:],
				scheme:              "http",
				releasesCache:       make(map[string]konfluxapi.ReleaseList),
				snapshotsCache:      make(map[string]applicationapi.Snapshot),
			}

			// First page
			list1, err1 := client.getReleases(ctx, namespace, "")
			Expect(err1).NotTo(HaveOccurred())
			Expect(list1.Continue).To(Equal("token-page-2"))

			// Second page
			list2, err2 := client.getReleases(ctx, namespace, "token-page-2")
			Expect(err2).NotTo(HaveOccurred())
			Expect(list2.Items).To(HaveLen(1))

			// Repeat first page - should use cache
			list3, err3 := client.getReleases(ctx, namespace, "")
			Expect(err3).NotTo(HaveOccurred())
			Expect(list3.Continue).To(Equal("token-page-2"))

			Expect(atomic.LoadInt32(&requestCount)).To(Equal(int32(2))) // Only 2 requests, not 3
		})
	})

	Describe("Iterator with memoization", func() {
		It("benefits from memoization on repeated iterations", func() {
			var requestCount int32

			releases := []konfluxapi.Release{
				{ObjectMeta: metav1.ObjectMeta{Name: "release-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "release-2"}},
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&requestCount, 1)
				releaseList := konfluxapi.ReleaseList{Items: releases}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				Expect(json.NewEncoder(w).Encode(releaseList)).NotTo(HaveOccurred())
			}))
			defer server.Close()

			client := &KubeArchiveHTTPClient{
				httpClient:          server.Client(),
				kubeArchiveHostname: server.URL[7:],
				scheme:              "http",
				releasesCache:       make(map[string]konfluxapi.ReleaseList),
				snapshotsCache:      make(map[string]applicationapi.Snapshot),
			}

			// First iteration
			iter1 := client.GetReleasesIterator(ctx, namespace)
			var collected1 []konfluxapi.Release
			for iter1.Next() {
				collected1 = append(collected1, iter1.Value())
			}
			Expect(iter1.Err()).NotTo(HaveOccurred())
			Expect(collected1).To(HaveLen(2))
			Expect(atomic.LoadInt32(&requestCount)).To(Equal(int32(1)))

			// Second iteration - should use cache
			iter2 := client.GetReleasesIterator(ctx, namespace)
			var collected2 []konfluxapi.Release
			for iter2.Next() {
				collected2 = append(collected2, iter2.Value())
			}
			Expect(iter2.Err()).NotTo(HaveOccurred())
			Expect(collected2).To(HaveLen(2))
			Expect(atomic.LoadInt32(&requestCount)).To(Equal(int32(1))) // Still 1, cache was used
		})
	})
})
