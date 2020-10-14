package v1alpha1

import "testing"

func TestEquals(t *testing.T) {

	testCases := []struct {
		name   string
		input1 *HistoryEntry
		input2 *HistoryEntry
		expect bool
	}{
		{
			name:   "both nil history",
			input1: nil,
			input2: nil,
			expect: true,
		},
		{
			name:   "single nil history (1)",
			input1: nil,
			input2: &HistoryEntry{},
			expect: false,
		},
		{
			name:   "single nil history (2)",
			input1: &HistoryEntry{},
			input2: nil,
			expect: false,
		},
		{
			name: "unequals (provider)",
			input1: &HistoryEntry{
				Spec: HistoryEntrySpec{
					Provider: "eks",
				},
			},
			input2: &HistoryEntry{
				Spec: HistoryEntrySpec{
					Provider: "aks",
				},
			},
			expect: false,
		},
		{
			name: "unequals (flags 1)",
			input1: &HistoryEntry{
				Spec: HistoryEntrySpec{
					Provider: "eks",
					Flags: map[string]string{
						"namespace": "",
					},
				},
			},
			input2: &HistoryEntry{
				Spec: HistoryEntrySpec{
					Provider: "eks",
					Flags: map[string]string{
						"namespace": "namespace1",
					},
				},
			},
			expect: false,
		},
		{
			name: "unequals (flags 2)",
			input1: &HistoryEntry{
				Spec: HistoryEntrySpec{
					Provider: "eks",
					Flags: map[string]string{
						"namespace": "namespace1",
					},
				},
			},
			input2: &HistoryEntry{
				Spec: HistoryEntrySpec{
					Provider: "eks",
					Flags:    map[string]string{},
				},
			},
			expect: false,
		},
		{
			name: "unequals (flags 3)",
			input1: &HistoryEntry{
				Spec: HistoryEntrySpec{
					Provider: "eks",
					Flags: map[string]string{
						"namespace": "namespace1",
					},
				},
			},
			input2: &HistoryEntry{
				Spec: HistoryEntrySpec{
					Provider: "eks",
					Flags: map[string]string{
						"namespace": "namespace2",
					},
				},
			},
			expect: false,
		},
		{
			name: "equals (flags 1)",
			input1: &HistoryEntry{
				Spec: HistoryEntrySpec{
					Provider: "eks",
					Flags: map[string]string{
						"namespace": "namespace1",
					},
				},
			},
			input2: &HistoryEntry{
				Spec: HistoryEntrySpec{
					Provider: "eks",
					Flags: map[string]string{
						"namespace": "namespace1",
					},
				},
			},
			expect: true,
		},
		{
			name: "equals (flags 2)",
			input1: &HistoryEntry{
				Spec: HistoryEntrySpec{
					Provider: "eks",
					Flags: map[string]string{
						"namespace":  "",
						"region":     "us-east-1",
						"kubeconfig": "",
					},
				},
			},
			input2: &HistoryEntry{
				Spec: HistoryEntrySpec{
					Provider: "eks",
					Flags: map[string]string{
						"namespace":  "",
						"region":     "us-east-1",
						"cluster-id": "",
					},
				},
			},
			expect: true,
		},
		{
			name: "equals (ignore flags)",
			input1: &HistoryEntry{
				Spec: HistoryEntrySpec{
					Provider: "eks",
					Flags: map[string]string{
						"namespace":  "",
						"region":     "us-east-1",
						"kubeconfig": "",
						"profile":    "profile1",
					},
				},
			},
			input2: &HistoryEntry{
				Spec: HistoryEntrySpec{
					Provider: "eks",
					Flags: map[string]string{
						"namespace":  "",
						"region":     "us-east-1",
						"cluster-id": "",
						"profile":    "profile2",
					},
				},
			},
			expect: true,
		},
	}
	for _, tc := range testCases {
		actualEqual := tc.input1.Equals(tc.input2)
		if actualEqual != tc.expect {
			t.Fatalf("expected %t, got %t", tc.expect, actualEqual)
		}
	}
}
