package matchers

import (
	"fmt"
	"reflect"

	"github.com/golang/mock/gomock"

	historyv1alpha "github.com/fidelity/kconnect/pkg/history/api/v1alpha1"
)

func MatchHistoryList(expected *historyv1alpha.HistoryEntryList) gomock.Matcher {
	return historyListMatcher{expected: expected}
}

type historyListMatcher struct {
	expected *historyv1alpha.HistoryEntryList
}

func (m historyListMatcher) Matches(x interface{}) bool {
	if !reflect.TypeOf(x).AssignableTo(reflect.TypeOf(m.expected)) {
		return false
	}

	actualHistoryList := x.(*historyv1alpha.HistoryEntryList)

	if len(actualHistoryList.Items) != len(m.expected.Items) {
		return false
	}

	for index := range actualHistoryList.Items {
		actualEntry := actualHistoryList.Items[index]
		expectedEntry := m.expected.Items[index]

		if !actualEntry.Equals(&expectedEntry) {
			return false
		}
	}

	return true
}

func (m historyListMatcher) String() string {
	return fmt.Sprintf("is equal to history list %v", m.expected)
}
