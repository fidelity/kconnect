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

// Package oauth implements an identity provider that authenticates against
// Microsoft Entra ID (Azure AD) using OAuth2 and exchanges the resulting
// token with AWS STS AssumeRoleWithWebIdentity for temporary AWS credentials.
// This is an alternative to the "saml" identity provider for teams whose
// Entra ID app registration is configured for programmatic workflows rather
// than interactive SAML/WS-Fed sign in (e.g. Ping Federate).
//
// Two grant types are supported:
//   - "password" (Resource Owner Password Credentials): the user's own
//     username/password is exchanged directly for a token. This only works
//     against Entra ID app registrations that do NOT enforce MFA/Conditional
//     Access, since ROPC cannot satisfy those challenges.
//   - "client_credentials": a service principal (client-id/client-secret)
//     is used instead of a user's own credentials, suited to automation.
package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"

	saml2awsconfig "github.com/versent/saml2aws/v2/pkg/awsconfig"

	kaws "github.com/fidelity/kconnect/pkg/aws"
	"github.com/fidelity/kconnect/pkg/config"
	khttp "github.com/fidelity/kconnect/pkg/http"
	"github.com/fidelity/kconnect/pkg/prompt"
	"github.com/fidelity/kconnect/pkg/provider"
	"github.com/fidelity/kconnect/pkg/provider/identity"
	"github.com/fidelity/kconnect/pkg/provider/registry"
)

const (
	// ProviderName is the name the provider is registered under
	ProviderName = "entraid-oauth"

	tokenEndpointFmt = "https://login.microsoftonline.com/%s/oauth2/v2.0/token" //nolint:gosec

	defaultSessionDuration = 3600

	tenantIDConfigKey     = "tenant-id"
	clientIDConfigKey     = "client-id"
	clientSecretConfigKey = "client-secret"
	roleARNConfigKey      = "role-arn"

	grantTypeConfigKey         = "grant-type"
	grantTypePassword          = "password"
	grantTypeClientCredentials = "client_credentials"

	// arnRoleRegexMatchCount is the expected number of matches from
	// arnRoleRegex: the full match plus the accountID and roleName
	// capture groups.
	arnRoleRegexMatchCount = 3

	awsProfileConfigKey         = "aws-profile"
	awsCredentialsFileConfigKey = "aws-shared-credentials-file"

	// awsRolesClaim is the claim some Entra ID app registrations for the
	// "AWS Federation OIDC" pattern emit in place of (or in addition to)
	// "groups". Unlike "groups", its entries are AWS IAM role ARNs directly
	awsRolesClaim = "https://aws.amazon.com/roles"
)

var (
	// ErrRoleARNRequired is returned when role-arn has no value and the
	// provider is running non-interactively, so it cannot be prompted for.
	ErrRoleARNRequired = errors.New("role-arn is required")

	// ErrNoAccessToken is returned when the Entra ID token endpoint does not
	// return an access_token. AWS validates the web identity token's
	// issuer/audience/signature against the OIDC provider configured in AWS
	// IAM, whose trust policy audience is expected to be the requested
	// scope's resource identifier (e.g. "api://<app-id>/AssumeRoleWithWebIdentity")
	// which is the access token's "aud", not the (always client-id-scoped)
	// id_token's "aud". The access_token is therefore what must be sent to
	// AssumeRoleWithWebIdentity here, not an id_token.
	ErrNoAccessToken = errors.New("no access_token returned from token endpoint")

	// ErrNoGroupsClaim is returned when the token has neither a "groups"
	// claim nor an "https://aws.amazon.com/roles" claim to derive AWS roles
	// from. The "groups" claim is optional in Entra ID app registrations, so
	// this provider falls back to the AWS roles claim when it is absent.
	ErrNoGroupsClaim = errors.New(`token has no "groups" or "https://aws.amazon.com/roles" claim: ` +
		"ensure the Entra ID app registration is configured to emit one of these claims")

	// ErrGroupsOverage is returned when Entra ID has omitted the "groups"
	// claim because the user is a member of more groups than fit in the
	// token (group overage), replacing it with a "_claim_names" indicator.
	// See: https://learn.microsoft.com/en-us/entra/identity-platform/id-token-claims-reference#groups-overage-claim
	ErrGroupsOverage = errors.New("token has a groups overage: user is a member of too many groups for Entra ID to include a groups claim; " +
		"use an app role or a different group instead of relying on the groups claim, or reduce the user's group membership count")

	// ErrNoRolesFound is returned when no AWS roles could be derived from the
	// token's group claims (mirrors the "saml" provider's behavior when no
	// roles are found in a SAML assertion)
	ErrNoRolesFound = errors.New("no aws roles found")

	ErrUnsupportedProvider = errors.New("cluster provider not supported")
)

// arnRoleRegex matches a plain AWS IAM role ARN, as found in the "https://aws.amazon.com/roles" claim.
// Some IdPs emit entries as a comma-separated "<role-arn>,<provider-arn>" pair (the SAML convention);
// only the role ARN portion is matched/used here since AssumeRoleWithWebIdentity has no use for a SAML provider ARN.
var arnRoleRegex = regexp.MustCompile(`(?i)^arn:aws:iam::(\d{12}):role/(.+)$`)

// awsRole represents an AWS IAM role discovered from the token's group claims
type awsRole struct {
	Name    string
	RoleARN string
}

func init() {
	if err := registry.RegisterIdentityPlugin(&registry.IdentityPluginRegistration{
		PluginRegistration: registry.PluginRegistration{
			Name:                   ProviderName,
			UsageExample:           "",
			ConfigurationItemsFunc: ConfigurationItems,
		},
		CreateFunc: New,
	}); err != nil {
		zap.S().Fatalw("Failed to register Azure AD OAuth identity plugin", "error", err)
	}
}

// New will create a new Azure AD OAuth identity provider
func New(input *provider.PluginCreationInput) (identity.Provider, error) {
	if input.HTTPClient == nil {
		return nil, provider.ErrHTTPClientRequired
	}

	return &oauthIdentityProvider{
		logger:            input.Logger,
		httpClient:        input.HTTPClient,
		interactive:       input.IsInteractive,
		scopedToDiscovery: *input.ScopedTo,
	}, nil
}

type oauthIdentityProvider struct {
	logger            *zap.SugaredLogger
	httpClient        khttp.Client
	interactive       bool
	scopedToDiscovery string
}

type oauthConfig struct {
	TenantID        string `json:"tenant-id"         validate:"required"`
	ClientID        string `json:"client-id"         validate:"required"`
	ClientSecret    string `json:"client-secret"     validate:"required_if=GrantType client_credentials"`
	Username        string `json:"username"          validate:"required_if=GrantType password"`
	Password        string `json:"password"          validate:"required_if=GrantType password"`
	GrantType       string `json:"grant-type"        validate:"required,oneof=password client_credentials"`
	Scope           string `json:"scope"             validate:"required"`
	RoleARN         string `json:"role-arn"`
	RoleSessionName string `json:"role-session-name" validate:"required"`
	Region          string `json:"region"`
	SessionDuration int    `json:"session-duration"`
}

// tokenResponse represents the response from the Entra ID token endpoint
type tokenResponse struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

func (p *oauthIdentityProvider) Name() string {
	return ProviderName
}

// Authenticate will get an OAuth2 token from Entra ID (using either the
// resource owner password or client credentials grant) and exchange it with
// AWS STS for temporary credentials.
func (p *oauthIdentityProvider) Authenticate(ctx context.Context, input *identity.AuthenticateInput) (*identity.AuthenticateOutput, error) {
	p.logger.Info("authenticating using Entra ID provider")

	if p.interactive {
		if err := p.resolveConfig(input.ConfigSet); err != nil {
			return nil, fmt.Errorf("resolving config: %w", err)
		}
	}

	cfg := &oauthConfig{
		SessionDuration: defaultSessionDuration,
		GrantType:       grantTypePassword,
	}
	if err := config.Unmarshall(input.ConfigSet, cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config into oauthConfig: %w", err)
	}

	if err := p.validateConfig(cfg); err != nil {
		return nil, err
	}

	token, err := p.getOAuthToken(cfg)
	if err != nil {
		return nil, fmt.Errorf("getting oauth token: %w", err)
	}

	roleARN, err := p.resolveRoleARN(input.ConfigSet, token)
	if err != nil {
		return nil, fmt.Errorf("resolving role-arn: %w", err)
	}

	cfg.RoleARN = roleARN

	awsCreds, err := p.assumeRoleWithWebIdentity(ctx, cfg, token)
	if err != nil {
		return nil, fmt.Errorf("assuming role with web identity: %w", err)
	}

	identifier, err := kaws.CreateIDFromCreds(awsCreds)
	if err != nil {
		return nil, fmt.Errorf("creating identifier from AWS creds: %w", err)
	}

	profileName := fmt.Sprintf("kconnect-%s", identifier)

	if err := p.setProfileName(profileName, input.ConfigSet); err != nil {
		return nil, fmt.Errorf("setting profile name: %w", err)
	}

	awsSharedCredentialsFile := ""
	if input.ConfigSet.ExistsWithValue(awsCredentialsFileConfigKey) {
		awsSharedCredentialsFile = input.ConfigSet.Get(awsCredentialsFileConfigKey).Value.(string)
	}
	p.logger.Debugw("using aws shared credentials file", "file", awsSharedCredentialsFile)

	awsIdentity := kaws.MapCredsToIdentity(awsCreds, profileName, awsSharedCredentialsFile)
	awsIdentity.IDProviderName = ProviderName

	store, err := p.createIdentityStore(input.ConfigSet)
	if err != nil {
		return nil, fmt.Errorf("creating identity store for %s: %w", p.scopedToDiscovery, err)
	}

	err = store.Save(awsIdentity)
	if err != nil {
		return nil, fmt.Errorf("saving identity: %w", err)
	}

	return &identity.AuthenticateOutput{
		Identity: awsIdentity,
	}, nil
}

func (p *oauthIdentityProvider) validateConfig(cfg *oauthConfig) error {
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("validating Entra ID OAuth config: %w", err)
	}

	return nil
}

func (p *oauthIdentityProvider) setProfileName(profileName string, cfg config.ConfigurationSet) error {
	if cfg.ExistsWithValue("static-profile") {
		p.logger.Debug("static profile name found")

		item := cfg.Get("static-profile")
		profileName = item.Value.(string)
	}

	p.logger.Debugw("setting aws profile name", "profile", profileName)

	item, err := cfg.String(awsProfileConfigKey, profileName, "AWS profile name to use")
	if err != nil {
		return fmt.Errorf("setting aws-profile: %w", err)
	}

	item.Value = profileName

	return nil
}

func (p *oauthIdentityProvider) createIdentityStore(cfg config.ConfigurationSet) (identity.Store, error) {
	var store identity.Store

	var err error

	switch p.scopedToDiscovery {
	case "eks":
		if !cfg.ExistsWithValue(awsProfileConfigKey) {
			return nil, kaws.ErrNoProfile
		}

		profileCfg := cfg.Get(awsProfileConfigKey)
		profile := profileCfg.Value.(string)
		awsCredsFileCfg := cfg.Get(awsCredentialsFileConfigKey)
		awsCredsFile := awsCredsFileCfg.Value.(string)
		store, err = kaws.NewIdentityStore(profile, ProviderName, awsCredsFile)
	default:
		return nil, ErrUnsupportedProvider
	}

	if err != nil {
		return nil, fmt.Errorf("creating identity store: %w", err)
	}

	return store, nil
}

// resolveConfig will interactively prompt the user for any required
// config items that have no value supplied. Only called when running
// interactively (i.e. --no-input is not set); non-interactive runs rely on
// the flags/config being supplied and will fail validation otherwise.
func (p *oauthIdentityProvider) resolveConfig(cfg config.ConfigurationSet) error {
	p.logger.Debug("resolving Entra ID identity configuration items")

	if err := prompt.InputAndSet(cfg, tenantIDConfigKey, "Enter the Entra ID (Azure AD) tenant id:", true); err != nil {
		return fmt.Errorf("resolving tenant-id: %w", err)
	}

	if err := prompt.InputAndSet(cfg, clientIDConfigKey, "Enter the Entra ID app registration client id:", true); err != nil {
		return fmt.Errorf("resolving client-id: %w", err)
	}

	if err := prompt.InputAndSet(cfg, "scope", "Enter the OAuth2 scope to request:", true); err != nil {
		return fmt.Errorf("resolving scope: %w", err)
	}

	if err := prompt.ChooseAndSet(cfg, grantTypeConfigKey, "Select the OAuth2 grant type to use", true,
		prompt.OptionsFromStringSlice([]string{grantTypePassword, grantTypeClientCredentials})); err != nil {
		return fmt.Errorf("resolving grant-type: %w", err)
	}

	grantType := cfg.ValueString(grantTypeConfigKey)
	if grantType == "" {
		grantType = grantTypePassword
	}

	switch grantType {
	case grantTypePassword:
		if err := prompt.InputAndSet(cfg, "username", "Username:", true); err != nil {
			return fmt.Errorf("resolving username: %w", err)
		}

		if err := prompt.InputSensitiveAndSet(cfg, "password", "Password:", true); err != nil {
			return fmt.Errorf("resolving password: %w", err)
		}
	case grantTypeClientCredentials:
		if err := prompt.InputSensitiveAndSet(cfg, clientSecretConfigKey, "Client Secret:", true); err != nil {
			return fmt.Errorf("resolving client-secret: %w", err)
		}
	}

	if err := kaws.ResolvePartition(cfg); err != nil {
		return fmt.Errorf("resolving partition: %w", err)
	}

	if err := kaws.ResolveRegion(cfg); err != nil {
		return fmt.Errorf("resolving region: %w", err)
	}

	if err := prompt.InputAndSet(cfg, "role-session-name", "Enter the AWS IAM role session name:", false); err != nil {
		return fmt.Errorf("resolving role-session-name: %w", err)
	}

	return nil
}

// resolveRoleARN determines the AWS IAM role ARN to assume. If already
// supplied via flag/config it is used as-is. Otherwise it attempts to derive
// a list of assumable roles from the token using the "https://aws.amazon.com/roles" claim,
// which some "AWS Federation OIDC" app registrations emit directly as a list of AWS IAM role ARNs.
// When interactive, the resulting roles are presented as a selectable dropdown.
// This mirrors the role dropdown the "saml" protocol offers. If no roles can be derived,
// this fails the same way "saml" does when no roles are found in a SAML assertion.
func (p *oauthIdentityProvider) resolveRoleARN(cfg config.ConfigurationSet, token string) (string, error) {
	if cfg.ExistsWithValue(roleARNConfigKey) {
		return cfg.ValueString(roleARNConfigKey), nil
	}

	roles, err := p.rolesFromToken(token)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrNoRolesFound, err)
	}

	if len(roles) == 0 {
		return "", ErrNoRolesFound
	}

	if !p.interactive {
		return "", ErrRoleARNRequired
	}

	options := make(map[string]string, len(roles))
	for _, role := range roles {
		options[role.Name] = role.RoleARN
	}

	if err := prompt.ChooseAndSet(cfg, roleARNConfigKey, "Select AWS role", true, prompt.OptionsFromMap(options)); err != nil {
		return "", fmt.Errorf("selecting aws role: %w", err)
	}

	return cfg.ValueString(roleARNConfigKey), nil
}

// rolesFromToken decodes the JWT access token's claims and extracts a list of
// assumable AWS roles from it.
//
// Note: Entra ID omits the groups claim and replaces it with an overage
// indicator, "_claim_names"/"_claim_sources" for users who are members of
// more than ~200 groups, even when the app registration is correctly
// configured to emit group claims. See ErrGroupsOverage.
func (p *oauthIdentityProvider) rolesFromToken(token string) ([]awsRole, error) {
	claims := jwt.MapClaims{}

	if _, _, err := jwt.NewParser().ParseUnverified(token, claims); err != nil {
		return nil, fmt.Errorf("parsing token claims: %w", err)
	}

	if _, overage := claims["_claim_names"]; overage {
		return nil, ErrGroupsOverage
	}

	if rawRoles, ok := claims[awsRolesClaim]; ok {
		p.logger.Debug("no groups claim found in token, falling back to https://aws.amazon.com/roles claim")
		return rolesFromAWSRolesClaim(rawRoles)
	}

	claimNames := make([]string, 0, len(claims))
	for name := range claims {
		claimNames = append(claimNames, name)
	}

	p.logger.Debugw("no groups or aws roles claim found in token", "claims_present", claimNames)

	return nil, ErrNoGroupsClaim
}

// rolesFromAWSRolesClaim parses the "https://aws.amazon.com/roles" claim's
// entries into AWS roles. Each entry is expected to be an AWS IAM role ARN,
// optionally paired with a SAML provider ARN as a comma-separated
// "<role-arn>,<provider-arn>" string (the SAML convention) only the role
// ARN portion is used, since AssumeRoleWithWebIdentity has no use for a SAML
// provider ARN.
func rolesFromAWSRolesClaim(rawRoles any) ([]awsRole, error) {
	entries, ok := rawRoles.([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected type for %s claim: %T", awsRolesClaim, rawRoles)
	}

	var roles []awsRole

	for _, e := range entries {
		entry, ok := e.(string)
		if !ok {
			continue
		}

		roleARN := strings.SplitN(entry, ",", 2)[0]

		matches := arnRoleRegex.FindStringSubmatch(roleARN)
		if len(matches) != arnRoleRegexMatchCount {
			continue
		}

		accountID := matches[1]
		roleName := matches[2]

		roles = append(roles, awsRole{
			Name:    fmt.Sprintf("%s / %s", accountID, roleName),
			RoleARN: roleARN,
		})
	}

	return roles, nil
}

func (p *oauthIdentityProvider) getOAuthToken(cfg *oauthConfig) (string, error) {
	tokenURL := fmt.Sprintf(tokenEndpointFmt, cfg.TenantID)

	form := url.Values{}
	form.Set("grant_type", cfg.GrantType)
	form.Set("client_id", cfg.ClientID)
	form.Set("scope", cfg.Scope)

	switch cfg.GrantType {
	case grantTypePassword:
		form.Set("username", cfg.Username)
		form.Set("password", cfg.Password)
		// client-secret is optional for ROPC against a public client app registration
		if cfg.ClientSecret != "" {
			form.Set("client_secret", cfg.ClientSecret)
		}
	case grantTypeClientCredentials:
		form.Set("client_secret", cfg.ClientSecret)
	}

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	p.logger.Debugw("requesting oauth token", "url", tokenURL)

	res, err := p.httpClient.Post(tokenURL, form.Encode(), headers)
	if err != nil {
		return "", fmt.Errorf("calling token endpoint: %w", err)
	}

	tokenRes := &tokenResponse{}
	if err := json.Unmarshal([]byte(res.Body()), tokenRes); err != nil {
		return "", fmt.Errorf("unmarshalling token response: %w", err)
	}

	if res.ResponseCode() != khttp.StatusCodeOK {
		return "", fmt.Errorf("token endpoint returned %d: %s: %s", res.ResponseCode(), tokenRes.Error, tokenRes.ErrorDesc)
	}

	// AssumeRoleWithWebIdentity is passed the access_token: AWS validates the token's
	// issuer/audience/signature against the OIDC provider configured in AWS IAM,
	// whose trust policy audience is expected to match the requested scope's resource
	// identifier (e.g. "api://<app-id>/AssumeRoleWithWebIdentity"). This is
	// the access token's "aud".
	if tokenRes.AccessToken == "" {
		return "", fmt.Errorf("%w: %s: %s", ErrNoAccessToken, tokenRes.Error, tokenRes.ErrorDesc)
	}

	return tokenRes.AccessToken, nil
}

// assumeRoleWithWebIdentity exchanges the OAuth2 access token for temporary
// AWS credentials using AssumeRoleWithWebIdentity.
func (p *oauthIdentityProvider) assumeRoleWithWebIdentity(ctx context.Context, cfg *oauthConfig, token string) (*saml2awsconfig.AWSCredentials, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("creating aws session: %w", err)
	}

	svc := sts.NewFromConfig(awsCfg)

	params := &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          aws.String(cfg.RoleARN),
		RoleSessionName:  aws.String(cfg.RoleSessionName),
		WebIdentityToken: aws.String(token),
		DurationSeconds:  aws.Int32(int32(cfg.SessionDuration)),
	}

	p.logger.Info("requesting AWS credentials using web identity token")

	// InvalidIdentityToken can be transient right after a token is freshly
	// issued (IdP-side propagation/clock-skew), so retry it the same way
	// stscreds.WebIdentityRoleProvider does internally.
	invalidIdentityTokenCode := (&ststypes.InvalidIdentityTokenException{}).ErrorCode()

	resp, err := svc.AssumeRoleWithWebIdentity(ctx, params, func(o *sts.Options) {
		o.Retryer = retry.AddWithErrorCodes(o.Retryer, invalidIdentityTokenCode)
	})
	if err != nil {
		return nil, fmt.Errorf("retrieving STS credentials using web identity: %w", err)
	}

	return &saml2awsconfig.AWSCredentials{
		AWSAccessKey:     aws.ToString(resp.Credentials.AccessKeyId),
		AWSSecretKey:     aws.ToString(resp.Credentials.SecretAccessKey),
		AWSSessionToken:  aws.ToString(resp.Credentials.SessionToken),
		AWSSecurityToken: aws.ToString(resp.Credentials.SessionToken),
		PrincipalARN:     aws.ToString(resp.AssumedRoleUser.Arn),
		Expires:          resp.Credentials.Expiration.Local(),
		Region:           cfg.Region,
	}, nil
}

// ConfigurationItems will return the configuration items for the identity plugin
func ConfigurationItems(scopeTo string) (config.ConfigurationSet, error) {
	cs := config.NewConfigurationSet()

	cs.String(tenantIDConfigKey, "", "The Entra ID (Azure AD) tenant id")                                                //nolint: errcheck
	cs.String(clientIDConfigKey, "", "The client id of the Entra ID app registration")                                   //nolint: errcheck
	cs.String(clientSecretConfigKey, "", "The client secret of the Entra ID app registration (client_credentials only)") //nolint: errcheck
	cs.String("username", "", "The username to authenticate with (password grant only)")                                 //nolint: errcheck
	cs.String("password", "", "The password to authenticate with (password grant only)")                                 //nolint: errcheck
	cs.String(grantTypeConfigKey, grantTypePassword, "The OAuth2 grant type to use: password or client_credentials")     //nolint: errcheck
	cs.String("scope", "", "The OAuth2 scope to request from Entra ID (e.g. api://<app-id>/AssumeRoleWithWebIdentity)")  //nolint: errcheck
	cs.String(roleARNConfigKey, "", "The ARN of the AWS IAM role to assume via AssumeRoleWithWebIdentity")               //nolint: errcheck
	cs.String("role-session-name", "kconnect", "The role session name to use when assuming the AWS IAM role")            //nolint: errcheck
	cs.Int("session-duration", defaultSessionDuration, "The duration, in seconds, of the requested AWS STS session")     //nolint: errcheck

	// partition/region are registered via the shared AWS config helpers so that
	// resolveConfig can use kaws. ResolvePartition/ResolveRegion, giving the same
	// dropdown-with-filter UX as the saml provider's AWS region prompt (and
	// honoring the "region-filter" setting from config.yaml).
	kaws.AddPartitionConfig(cs)
	kaws.AddRegionConfig(cs)

	cs.SetSensitive(clientSecretConfigKey) //nolint: errcheck
	cs.SetSensitive("password")            //nolint: errcheck

	// client-secret/username/password/role-arn/region are conditionally
	// required or resolved interactively (see resolveConfig), so they are
	// not marked required on the flag itself.
	cs.SetRequired(tenantIDConfigKey) //nolint: errcheck
	cs.SetRequired(clientIDConfigKey) //nolint: errcheck
	cs.SetRequired("scope")           //nolint: errcheck

	return cs, nil
}
