/*
Copyright 2020 The kconnect Authors.

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

package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/app"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/history/loader"
)

const (
	// foreignContext is a context that kconnect did not create
	foreignContext = "minikube"
	foreignUser    = "minikube-user"
)

type logoutFixture struct {
	app        *app.App
	kubeconfig string
	awsCreds   string
	ids        map[string]string
}

type testEntry struct {
	key      string
	provider string
	alias    string
	flags    map[string]string
}

// newLogoutFixture creates a history file, a kubeconfig and an AWS credentials
// file in a temporary directory that reflect the supplied history entries.
func newLogoutFixture(t *testing.T, entries ...testEntry) *logoutFixture {
	t.Helper()

	g := NewWithT(t)
	dir := t.TempDir()

	fileLoader, err := loader.NewFileLoader(filepath.Join(dir, "history.yaml"))
	g.Expect(err).NotTo(HaveOccurred())

	store, err := history.NewStore(100, fileLoader)
	g.Expect(err).NotTo(HaveOccurred())

	fixture := &logoutFixture{
		app:        app.New(app.WithHistoryStore(store)),
		kubeconfig: filepath.Join(dir, "kubeconfig"),
		awsCreds:   filepath.Join(dir, "credentials"),
		ids:        map[string]string{},
	}

	kubeCfg := api.NewConfig()
	kubeCfg.Clusters["a-cluster"] = &api.Cluster{Server: "https://cluster.example.com"}
	kubeCfg.Contexts[foreignContext] = &api.Context{Cluster: "a-cluster", AuthInfo: foreignUser}
	kubeCfg.AuthInfos[foreignUser] = &api.AuthInfo{Token: "not-kconnect"}

	awsCreds := ""

	for _, e := range entries {
		entry := historyv1alpha.NewHistoryEntry()
		entry.Spec.Provider = e.provider
		entry.Spec.ProviderID = e.key
		entry.Spec.Flags = e.flags

		if e.alias != "" {
			alias := e.alias
			entry.Spec.Alias = &alias
		}

		g.Expect(store.Add(entry)).To(Succeed())
		fixture.ids[e.key] = entry.Name

		user := "user-" + e.key
		kubeCfg.AuthInfos[user] = &api.AuthInfo{Token: "token-" + e.key}
		kubeCfg.Contexts["context-"+e.key] = &api.Context{
			Cluster:  "a-cluster",
			AuthInfo: user,
			Extensions: map[string]runtime.Object{
				"kconnect": historyv1alpha.NewHistoryReference(entry.Name),
			},
		}

		if profile, ok := e.flags["aws-profile"]; ok {
			awsCreds += "[" + profile + "]\naws_access_key_id = AKIA" + e.key + "\n\n"
		}
	}

	awsCreds += "[keep-me]\naws_access_key_id = AKIAKEEP\n"

	g.Expect(clientcmd.WriteToFile(*kubeCfg, fixture.kubeconfig)).To(Succeed())
	g.Expect(os.WriteFile(fixture.awsCreds, []byte(awsCreds), 0o600)).To(Succeed())
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", fixture.awsCreds)

	return fixture
}

func (f *logoutFixture) logout(input *app.LogoutInput) error {
	input.Kubeconfig = f.kubeconfig
	return f.app.Logout(context.Background(), input)
}

func (f *logoutFixture) users(t *testing.T) []string {
	t.Helper()

	cfg, err := clientcmd.LoadFromFile(f.kubeconfig)
	NewWithT(t).Expect(err).NotTo(HaveOccurred())

	users := []string{}
	for name := range cfg.AuthInfos {
		users = append(users, name)
	}

	return users
}

func (f *logoutFixture) awsCredentials(t *testing.T) string {
	t.Helper()

	data, err := os.ReadFile(f.awsCreds)
	NewWithT(t).Expect(err).NotTo(HaveOccurred())

	return string(data)
}

func TestLogoutByAlias(t *testing.T) {
	g := NewWithT(t)
	f := newLogoutFixture(t,
		testEntry{key: "one", provider: app.RancherProviderName, alias: "dev"},
		testEntry{key: "two", provider: app.AKSProviderName, alias: "prod"},
	)

	g.Expect(f.logout(&app.LogoutInput{Alias: "dev"})).To(Succeed())
	g.Expect(f.users(t)).To(ConsistOf(foreignUser, "user-two"))
}

func TestLogoutByMultipleAliases(t *testing.T) {
	g := NewWithT(t)
	f := newLogoutFixture(t,
		testEntry{key: "one", provider: app.RancherProviderName, alias: "dev"},
		testEntry{key: "two", provider: app.AKSProviderName, alias: "prod"},
		testEntry{key: "three", provider: app.AKSProviderName, alias: "test"},
	)

	g.Expect(f.logout(&app.LogoutInput{Alias: "dev, prod"})).To(Succeed())
	g.Expect(f.users(t)).To(ConsistOf(foreignUser, "user-three"))
}

func TestLogoutByID(t *testing.T) {
	g := NewWithT(t)
	f := newLogoutFixture(t,
		testEntry{key: "one", provider: app.RancherProviderName, alias: "dev"},
		testEntry{key: "two", provider: app.AKSProviderName},
	)

	g.Expect(f.logout(&app.LogoutInput{IDs: f.ids["two"]})).To(Succeed())
	g.Expect(f.users(t)).To(ConsistOf(foreignUser, "user-one"))
}

func TestLogoutByAliasAndID(t *testing.T) {
	g := NewWithT(t)
	f := newLogoutFixture(t,
		testEntry{key: "one", provider: app.RancherProviderName, alias: "dev"},
		testEntry{key: "two", provider: app.AKSProviderName},
		testEntry{key: "three", provider: app.AKSProviderName, alias: "test"},
	)

	g.Expect(f.logout(&app.LogoutInput{Alias: "dev", IDs: f.ids["two"]})).To(Succeed())
	g.Expect(f.users(t)).To(ConsistOf(foreignUser, "user-three"))
}

func TestLogoutUnknownAlias(t *testing.T) {
	g := NewWithT(t)
	f := newLogoutFixture(t,
		testEntry{key: "one", provider: app.RancherProviderName, alias: "dev"},
	)

	err := f.logout(&app.LogoutInput{Alias: "nope"})
	g.Expect(err).To(MatchError(app.ErrAliasNotFound))
	g.Expect(err.Error()).To(ContainSubstring("nope"))
	g.Expect(f.users(t)).To(ConsistOf(foreignUser, "user-one"))
}

func TestLogoutUnknownID(t *testing.T) {
	g := NewWithT(t)
	f := newLogoutFixture(t,
		testEntry{key: "one", provider: app.RancherProviderName, alias: "dev"},
	)

	err := f.logout(&app.LogoutInput{IDs: "does-not-exist"})
	g.Expect(err).To(MatchError(history.ErrEntryNotFound))
	g.Expect(f.users(t)).To(ConsistOf(foreignUser, "user-one"))
}

func TestLogoutAllWithEntriesWithoutAnAlias(t *testing.T) {
	g := NewWithT(t)
	f := newLogoutFixture(t,
		testEntry{key: "one", provider: app.RancherProviderName},
		testEntry{key: "two", provider: app.AKSProviderName},
		testEntry{key: "three", provider: app.OIDCProviderName},
		testEntry{key: "four", provider: app.EKSProviderName},
		testEntry{key: "five", provider: app.EKSProviderName, flags: map[string]string{"aws-profile": "profile-five"}},
	)

	g.Expect(f.logout(&app.LogoutInput{All: true})).To(Succeed())

	// The EKS users are left alone as their credentials are in the AWS profile
	g.Expect(f.users(t)).To(ConsistOf(foreignUser, "user-four", "user-five"))
	g.Expect(f.awsCredentials(t)).NotTo(ContainSubstring("profile-five"))
	g.Expect(f.awsCredentials(t)).To(ContainSubstring("keep-me"))
}

func TestLogoutOIDC(t *testing.T) {
	g := NewWithT(t)
	f := newLogoutFixture(t,
		testEntry{key: "one", provider: app.OIDCProviderName, alias: "sso"},
		testEntry{key: "two", provider: app.AKSProviderName},
	)

	g.Expect(f.logout(&app.LogoutInput{Alias: "sso"})).To(Succeed())
	g.Expect(f.users(t)).To(ConsistOf(foreignUser, "user-two"))
}

func TestLogoutEKSRemovesOnlyItsProfile(t *testing.T) {
	g := NewWithT(t)
	f := newLogoutFixture(t,
		testEntry{key: "one", provider: app.EKSProviderName, alias: "a", flags: map[string]string{"aws-profile": "profile-one"}},
		testEntry{key: "two", provider: app.EKSProviderName, alias: "b", flags: map[string]string{"aws-profile": "profile-two"}},
	)

	g.Expect(f.logout(&app.LogoutInput{Alias: "a"})).To(Succeed())

	creds := f.awsCredentials(t)
	g.Expect(creds).NotTo(ContainSubstring("profile-one"))
	g.Expect(creds).To(ContainSubstring("profile-two"))
	g.Expect(creds).To(ContainSubstring("keep-me"))
}

func TestLogoutDefaultsToTheMostRecentEntry(t *testing.T) {
	g := NewWithT(t)
	f := newLogoutFixture(t,
		testEntry{key: "one", provider: app.RancherProviderName},
	)

	g.Expect(f.logout(&app.LogoutInput{})).To(Succeed())
	g.Expect(f.users(t)).To(ConsistOf(foreignUser))
}

func TestLogoutWithNoHistory(t *testing.T) {
	g := NewWithT(t)
	f := newLogoutFixture(t)

	g.Expect(f.logout(&app.LogoutInput{})).To(MatchError(history.ErrNoEntries))
	g.Expect(f.logout(&app.LogoutInput{All: true})).To(MatchError(app.ErrNoEntriesFound))
	g.Expect(f.users(t)).To(ConsistOf(foreignUser))
}

func TestLogoutUnknownProvider(t *testing.T) {
	g := NewWithT(t)
	f := newLogoutFixture(t,
		testEntry{key: "one", provider: "something-else", alias: "odd"},
	)

	err := f.logout(&app.LogoutInput{Alias: "odd"})
	g.Expect(err).To(MatchError(app.ErrUnknownProvider))
	g.Expect(err.Error()).To(ContainSubstring("something-else"))
}

func TestLogoutWhenTheUserIsAlreadyGone(t *testing.T) {
	g := NewWithT(t)
	f := newLogoutFixture(t,
		testEntry{key: "one", provider: app.RancherProviderName, alias: "dev"},
	)

	g.Expect(f.logout(&app.LogoutInput{Alias: "dev"})).To(Succeed())
	// A second logout has nothing to remove and must not fail
	g.Expect(f.logout(&app.LogoutInput{Alias: "dev"})).To(Succeed())
	g.Expect(f.users(t)).To(ConsistOf(foreignUser))
}
