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

	applicationapi "github.com/konflux-ci/application-api/api/v1alpha1"
	konfluxapi "github.com/konflux-ci/release-service/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var _ = Describe("KubeArchive Client", func() {
	Describe("ClientFor", func() {
		It("creates a client successfully", func() {
			config := &rest.Config{
				Host: "https://api.example.konflux.dev",
			}

			client, err := ClientFor(config)
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())

			httpClient, ok := client.(*KubeArchiveHTTPClient)
			Expect(ok).To(BeTrue())
			Expect(httpClient.kubeArchiveHostname).To(Equal("kubearchive-api-server-product-kubearchive.apps.example.konflux.dev"))
		})
	})

	Describe("kubeArchiveHostnameFromClusterHostname", func() {
		DescribeTable("derives correct hostname",
			func(clusterHost, expected string) {
				result := kubeArchiveHostnameFromClusterHostname(clusterHost)
				Expect(result).To(Equal(expected))
			},
			Entry("standard cluster", "https://api.cluster.example.com", "kubearchive-api-server-product-kubearchive.apps.cluster.example.com"),
			Entry("konflux cluster", "https://api.example.konflux.dev", "kubearchive-api-server-product-kubearchive.apps.example.konflux.dev"),
		)
	})

	Describe("GetSnapshot", func() {
		var (
			server       *httptest.Server
			client       *KubeArchiveHTTPClient
			ctx          context.Context
			namespace    string
			snapshotName string
		)

		BeforeEach(func() {
			ctx = context.Background()
			namespace = "test-namespace"
			snapshotName = "test-snapshot"
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		It("retrieves a snapshot successfully", func() {
			expectedSnapshot := applicationapi.Snapshot{
				ObjectMeta: metav1.ObjectMeta{
					Name:      snapshotName,
					Namespace: namespace,
				},
				Spec: applicationapi.SnapshotSpec{
					Application: "test-app",
				},
			}

			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("GET"))
				Expect(r.URL.Path).To(Equal(fmt.Sprintf("/apis/appstudio.redhat.com/v1alpha1/namespaces/%s/snapshots/%s", namespace, snapshotName)))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				Expect(json.NewEncoder(w).Encode(expectedSnapshot)).NotTo(HaveOccurred())
			}))

			client = &KubeArchiveHTTPClient{
				httpClient:          server.Client(),
				kubeArchiveHostname: server.URL[7:], // Remove "http://"
				scheme:              "http",
				releasesCache:       make(map[string]konfluxapi.ReleaseList),
				snapshotsCache:      make(map[string]applicationapi.Snapshot),
			}

			snapshot, err := client.GetSnapshot(ctx, namespace, snapshotName)
			Expect(err).NotTo(HaveOccurred())
			Expect(snapshot.Name).To(Equal(snapshotName))
			Expect(snapshot.Namespace).To(Equal(namespace))
			Expect(snapshot.Spec.Application).To(Equal("test-app"))
		})

		It("returns error when snapshot not found", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				Expect(w.Write([]byte(`{"error": "not found"}`))).NotTo(HaveOccurred())
			}))

			client = &KubeArchiveHTTPClient{
				httpClient:          server.Client(),
				kubeArchiveHostname: server.URL[7:],
				scheme:              "http",
				releasesCache:       make(map[string]konfluxapi.ReleaseList),
				snapshotsCache:      make(map[string]applicationapi.Snapshot),
			}

			_, err := client.GetSnapshot(ctx, namespace, snapshotName)
			Expect(err).To(HaveOccurred())
			Expect(IsNotFound(err)).To(BeTrue())
		})

		It("returns error on malformed JSON", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				Expect(w.Write([]byte(`{invalid json`))).NotTo(HaveOccurred())
			}))

			client = &KubeArchiveHTTPClient{
				httpClient:          server.Client(),
				kubeArchiveHostname: server.URL[7:],
				scheme:              "http",
				releasesCache:       make(map[string]konfluxapi.ReleaseList),
				snapshotsCache:      make(map[string]applicationapi.Snapshot),
			}

			_, err := client.GetSnapshot(ctx, namespace, snapshotName)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetReleasesIterator", func() {
		var (
			server    *httptest.Server
			client    *KubeArchiveHTTPClient
			ctx       context.Context
			namespace string
		)

		BeforeEach(func() {
			ctx = context.Background()
			namespace = "test-namespace"
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		It("iterates through a single page of releases", func() {
			releases := []konfluxapi.Release{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "release-1",
						Namespace: namespace,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "release-2",
						Namespace: namespace,
					},
				},
			}

			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("GET"))
				Expect(r.URL.Path).To(Equal(fmt.Sprintf("/apis/appstudio.redhat.com/v1alpha1/namespaces/%s/releases", namespace)))

				releaseList := konfluxapi.ReleaseList{
					Items: releases,
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				Expect(json.NewEncoder(w).Encode(releaseList)).NotTo(HaveOccurred())
			}))

			client = &KubeArchiveHTTPClient{
				httpClient:          server.Client(),
				kubeArchiveHostname: server.URL[7:],
				scheme:              "http",
				releasesCache:       make(map[string]konfluxapi.ReleaseList),
				snapshotsCache:      make(map[string]applicationapi.Snapshot),
			}

			iter := client.GetReleasesIterator(ctx, namespace)
			Expect(iter).NotTo(BeNil())

			var collected []konfluxapi.Release
			for iter.Next() {
				collected = append(collected, iter.Value())
			}
			Expect(iter.Err()).NotTo(HaveOccurred())
			Expect(collected).To(HaveLen(2))
			Expect(collected[0].Name).To(Equal("release-1"))
			Expect(collected[1].Name).To(Equal("release-2"))
		})

		It("iterates through multiple pages with continue token", func() {
			page1 := []konfluxapi.Release{
				{ObjectMeta: metav1.ObjectMeta{Name: "release-1"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "release-2"}},
			}
			page2 := []konfluxapi.Release{
				{ObjectMeta: metav1.ObjectMeta{Name: "release-3"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "release-4"}},
			}

			requestCount := 0
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++
				continueToken := r.URL.Query().Get("continue")

				var releaseList konfluxapi.ReleaseList
				switch continueToken {
				case "":
					// First page
					releaseList = konfluxapi.ReleaseList{
						ListMeta: metav1.ListMeta{
							Continue: "token-page-2",
						},
						Items: page1,
					}
				case "token-page-2":
					// Second page
					releaseList = konfluxapi.ReleaseList{
						Items: page2,
					}
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				Expect(json.NewEncoder(w).Encode(releaseList)).NotTo(HaveOccurred())
			}))

			client = &KubeArchiveHTTPClient{
				httpClient:          server.Client(),
				kubeArchiveHostname: server.URL[7:],
				scheme:              "http",
				releasesCache:       make(map[string]konfluxapi.ReleaseList),
				snapshotsCache:      make(map[string]applicationapi.Snapshot),
			}

			iter := client.GetReleasesIterator(ctx, namespace)

			var collected []konfluxapi.Release
			for iter.Next() {
				collected = append(collected, iter.Value())
			}
			Expect(iter.Err()).NotTo(HaveOccurred())
			Expect(collected).To(HaveLen(4))
			Expect(collected[0].Name).To(Equal("release-1"))
			Expect(collected[1].Name).To(Equal("release-2"))
			Expect(collected[2].Name).To(Equal("release-3"))
			Expect(collected[3].Name).To(Equal("release-4"))
			Expect(requestCount).To(Equal(2))
		})

		It("handles empty release list", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				releaseList := konfluxapi.ReleaseList{
					Items: []konfluxapi.Release{},
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				Expect(json.NewEncoder(w).Encode(releaseList)).NotTo(HaveOccurred())
			}))

			client = &KubeArchiveHTTPClient{
				httpClient:          server.Client(),
				kubeArchiveHostname: server.URL[7:],
				scheme:              "http",
				releasesCache:       make(map[string]konfluxapi.ReleaseList),
				snapshotsCache:      make(map[string]applicationapi.Snapshot),
			}

			iter := client.GetReleasesIterator(ctx, namespace)

			var collected []konfluxapi.Release
			for iter.Next() {
				collected = append(collected, iter.Value())
			}
			Expect(iter.Err()).NotTo(HaveOccurred())
			Expect(collected).To(BeEmpty())
		})

		It("handles API errors during iteration", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				Expect(w.Write([]byte(`{"error": "unauthorized"}`))).NotTo(HaveOccurred())
			}))

			client = &KubeArchiveHTTPClient{
				httpClient:          server.Client(),
				kubeArchiveHostname: server.URL[7:],
				scheme:              "http",
				releasesCache:       make(map[string]konfluxapi.ReleaseList),
				snapshotsCache:      make(map[string]applicationapi.Snapshot),
			}

			iter := client.GetReleasesIterator(ctx, namespace)

			hasItems := iter.Next()
			Expect(hasItems).To(BeFalse())
			Expect(iter.Err()).To(HaveOccurred())
			Expect(IsUnauthorized(iter.Err())).To(BeTrue())
		})

		It("handles errors on second page", func() {
			requestCount := 0
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++
				continueToken := r.URL.Query().Get("continue")

				if continueToken == "" {
					// First page succeeds
					releaseList := konfluxapi.ReleaseList{
						ListMeta: metav1.ListMeta{
							Continue: "token-page-2",
						},
						Items: []konfluxapi.Release{
							{ObjectMeta: metav1.ObjectMeta{Name: "release-1"}},
						},
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					Expect(json.NewEncoder(w).Encode(releaseList)).NotTo(HaveOccurred())
				} else {
					// Second page fails
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					Expect(w.Write([]byte(`{"error": "server error"}`))).NotTo(HaveOccurred())
				}
			}))

			client = &KubeArchiveHTTPClient{
				httpClient:          server.Client(),
				kubeArchiveHostname: server.URL[7:],
				scheme:              "http",
				releasesCache:       make(map[string]konfluxapi.ReleaseList),
				snapshotsCache:      make(map[string]applicationapi.Snapshot),
			}

			iter := client.GetReleasesIterator(ctx, namespace)

			var collected []konfluxapi.Release
			for iter.Next() {
				collected = append(collected, iter.Value())
			}

			// Should have collected the first page before error
			Expect(collected).To(HaveLen(1))
			Expect(collected[0].Name).To(Equal("release-1"))
			Expect(iter.Err()).To(HaveOccurred())
		})
	})
})
