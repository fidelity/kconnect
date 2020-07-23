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

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/eks"

	idprov "github.com/fidelity/kconnect/pkg/plugins/identity/saml"
	"github.com/fidelity/kconnect/pkg/provider"
)

func (p *eksClusterProvider) Discover(ctx *provider.Context, identity provider.Identity) error {
	logger := ctx.Logger.WithField("provider", "eks")
	logger.Info("discovering EKS clusters AWS")

	awsID := identity.(*idprov.AWSIdentity)

	logger.Debugf("creating AWS session with region %s and profile %s", *p.region, awsID.ProfileName)
	session, err := newSession(*p.region, awsID.ProfileName)
	if err != nil {
		return fmt.Errorf("getting aws session: %w", err)
	}

	eksClient := newEKSClient(session)
	input := &eks.ListClustersInput{}

	//TODO: handle cluster name flag

	clusters := []*string{}
	err = eksClient.ListClustersPages(input, func(page *eks.ListClustersOutput, lastPage bool) bool {
		clusters = append(clusters, page.Clusters...)
		return true
	})
	if err != nil {
		return fmt.Errorf("listing clusters: %w", err)
	}

	if len(clusters) == 0 {
		logger.Info("no EKS clusters discovered")
		return nil
	}

	logger.Debugf("found clusters: %#v", clusters)

	// Prompt user to select cluster

	return nil
}
