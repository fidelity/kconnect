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
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/onsi/gomega"

	historyv1alpha "github.com/fidelity/kconnect/api/v1alpha1"
	"github.com/fidelity/kconnect/pkg/app"
	"github.com/fidelity/kconnect/pkg/history"
	"github.com/fidelity/kconnect/pkg/history/loader"
	"github.com/fidelity/kconnect/pkg/printer"
)

type historyFixture struct {
	app   *app.App
	store history.Store
	path  string
}

func newHistoryFixture(t *testing.T) *historyFixture {
	t.Helper()

	g := NewWithT(t)
	path := filepath.Join(t.TempDir(), "history.yaml")

	fileLoader, err := loader.NewFileLoader(path)
	g.Expect(err).NotTo(HaveOccurred())

	store, err := history.NewStore(100, fileLoader)
	g.Expect(err).NotTo(HaveOccurred())

	return &historyFixture{app: app.New(app.WithHistoryStore(store)), store: store, path: path}
}

func (f *historyFixture) add(t *testing.T, providerID, alias string, flags map[string]string) *historyv1alpha.HistoryEntry {
	t.Helper()

	entry := historyv1alpha.NewHistoryEntry()
	entry.Spec.Provider = "eks"
	entry.Spec.ProviderID = providerID
	entry.Spec.Flags = flags

	if alias != "" {
		entry.Spec.Alias = &alias
	}

	NewWithT(t).Expect(f.store.Add(entry)).To(Succeed())

	return entry
}

func (f *historyFixture) providerIDs(t *testing.T) []string {
	t.Helper()

	list, err := f.store.GetAll()
	NewWithT(t).Expect(err).NotTo(HaveOccurred())

	ids := []string{}
	for i := range list.Items {
		ids = append(ids, list.Items[i].Spec.ProviderID)
	}

	return ids
}

func TestHistoryRemoveUnknownIDDoesNotPanic(t *testing.T) {
	g := NewWithT(t)
	f := newHistoryFixture(t)
	f.add(t, "one", "dev", nil)

	err := f.app.HistoryRemove(context.Background(), &app.HistoryRemoveInput{RemoveList: []string{"does-not-exist"}})
	g.Expect(err).To(MatchError(history.ErrEntryNotFound))
	g.Expect(err.Error()).To(ContainSubstring("does-not-exist"))
	g.Expect(f.providerIDs(t)).To(ConsistOf("one"))
}

func TestHistoryRemoveByID(t *testing.T) {
	g := NewWithT(t)
	f := newHistoryFixture(t)
	one := f.add(t, "one", "dev", nil)
	f.add(t, "two", "prod", nil)

	g.Expect(f.app.HistoryRemove(context.Background(), &app.HistoryRemoveInput{RemoveList: []string{one.Name}})).To(Succeed())
	g.Expect(f.providerIDs(t)).To(ConsistOf("two"))
}

func TestHistoryRemoveByFilterNeedsEveryConditionToMatch(t *testing.T) {
	f := newHistoryFixture(t)
	f.add(t, "east-dev", "a", map[string]string{"region": "us-east-1", "namespace": "dev"})
	f.add(t, "east-prod", "b", map[string]string{"region": "us-east-1", "namespace": "prod"})
	f.add(t, "west-dev", "c", map[string]string{"region": "us-west-2", "namespace": "dev"})

	input := &app.HistoryRemoveInput{}
	input.Filter = "region=us-east-1,namespace=dev"

	g := NewWithT(t)
	g.Expect(f.app.HistoryRemove(context.Background(), input)).To(Succeed())

	// Only the entry that matched both conditions was removed
	g.Expect(f.providerIDs(t)).To(ConsistOf("east-prod", "west-dev"))
}

func TestHistoryRemoveByFilterWithSpecialCharactersDoesNotPanic(t *testing.T) {
	g := NewWithT(t)
	f := newHistoryFixture(t)
	f.add(t, "one", "prod(", nil)
	f.add(t, "two", "prod", nil)

	input := &app.HistoryRemoveInput{}
	input.Filter = "alias=prod("

	g.Expect(f.app.HistoryRemove(context.Background(), input)).To(Succeed())
	g.Expect(f.providerIDs(t)).To(ConsistOf("two"))
}

func TestAddKeepsTheKubeconfigOfANewEntry(t *testing.T) {
	g := NewWithT(t)
	f := newHistoryFixture(t)

	first := historyv1alpha.NewHistoryEntry()
	first.Spec.Provider = "eks"
	first.Spec.ProviderID = "one"
	first.Spec.ConfigFile = "/home/me/.kube/config"
	g.Expect(f.store.Add(first)).To(Succeed())

	second := historyv1alpha.NewHistoryEntry()
	second.Spec.Provider = "eks"
	second.Spec.ProviderID = "two"
	second.Spec.ConfigFile = "/home/me/other-kubeconfig"
	g.Expect(f.store.Add(second)).To(Succeed())

	saved, err := f.store.GetByID(second.Name)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(saved).NotTo(BeNil())
	g.Expect(saved.Spec.ConfigFile).To(Equal("/home/me/other-kubeconfig"))
}

func TestAddTheSameConnectionWithAnotherKubeconfigIsNotADuplicate(t *testing.T) {
	g := NewWithT(t)
	f := newHistoryFixture(t)

	first := historyv1alpha.NewHistoryEntry()
	first.Spec.Provider = "eks"
	first.Spec.ProviderID = "one"
	first.Spec.ConfigFile = "/home/me/.kube/config"
	g.Expect(f.store.Add(first)).To(Succeed())

	again := historyv1alpha.NewHistoryEntry()
	again.Spec.Provider = "eks"
	again.Spec.ProviderID = "one"
	again.Spec.ConfigFile = "/home/me/other-kubeconfig"
	g.Expect(f.store.Add(again)).To(Succeed())

	g.Expect(f.providerIDs(t)).To(ConsistOf("one"))
	g.Expect(again.Name).To(Equal(first.Name))
}

func TestHistoryWithEntriesWithoutAnAliasCanBeListed(t *testing.T) {
	g := NewWithT(t)
	f := newHistoryFixture(t)
	f.add(t, "one", "", nil)
	f.add(t, "two", "dev", nil)

	list, err := f.store.GetAllSortedByLastUsed()
	g.Expect(err).NotTo(HaveOccurred())

	objPrinter, err := printer.New("table")
	g.Expect(err).NotTo(HaveOccurred())

	buf := new(bytes.Buffer)
	g.Expect(objPrinter.Print(list.ToTable(""), buf)).To(Succeed())
	g.Expect(buf.String()).To(ContainSubstring("one"))
	g.Expect(buf.String()).To(ContainSubstring("dev"))
}

func TestHistoryFileIsPrivate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file permissions")
	}

	g := NewWithT(t)
	f := newHistoryFixture(t)

	info, err := os.Stat(f.path)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o600)))

	// A file that was created by an earlier version is tightened when it is next saved
	g.Expect(os.Chmod(f.path, 0o644)).To(Succeed())
	f.add(t, "one", "dev", map[string]string{"secret-key": "a-secret"})

	info, err = os.Stat(f.path)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o600)))
}
