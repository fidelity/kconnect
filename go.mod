module github.com/fidelity/kconnect

go 1.13

require (
	github.com/AlecAivazis/survey/v2 v2.1.1
	github.com/Azure/azure-sdk-for-go v48.0.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.10
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.3
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.0 // indirect
	github.com/Azure/go-ntlmssp v0.0.0-20200615164410-66371956d46c // indirect
	github.com/PuerkitoBio/goquery v1.5.1 // indirect
	github.com/andybalholm/cascadia v1.2.0 // indirect
	github.com/aws/aws-sdk-go v1.35.2
	github.com/beevik/etree v1.1.0
	github.com/blang/semver v3.5.0+incompatible
	github.com/brianvoe/gofakeit/v5 v5.10.1
	github.com/go-playground/validator/v10 v10.3.0
	github.com/golang/mock v1.4.1
	github.com/golangci/golangci-lint v1.31.0 // indirect
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.1.2
	github.com/imdario/mergo v0.3.10 // indirect
	github.com/marshallbrekka/go-u2fhost v0.0.0-20200114212649-cc764c209ee9 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.3.2 // indirect
	github.com/oklog/ulid v1.3.1
	github.com/oklog/ulid/v2 v2.0.2 // indirect
	github.com/onsi/gomega v1.10.1
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/spf13/afero v1.3.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/tidwall/gjson v1.6.0 // indirect
	github.com/tidwall/pretty v1.0.1 // indirect
	github.com/versent/saml2aws v1.8.5-0.20200622110128-d94772688a70
	github.com/worr/saml2aws v2.15.0+incompatible // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897 // indirect
	golang.org/x/tools v0.0.0-20201103235415-b653051172e4 // indirect
	gopkg.in/ini.v1 v1.62.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.20.0 // indirect
	k8s.io/apimachinery v0.20.0
	k8s.io/cli-runtime v0.20.0
	k8s.io/client-go v0.20.0
	k8s.io/kubectl v0.19.1 // indirect
	sigs.k8s.io/controller-tools v0.4.0 // indirect
	sigs.k8s.io/kubebuilder/docs/book/utils v0.0.0-20201009223647-5031c94d9175 // indirect
	sigs.k8s.io/structured-merge-diff/v2 v2.0.1 // indirect
	sigs.k8s.io/yaml v1.2.0

)

replace (
	github.com/spf13/cobra => github.com/richardcase/cobra v1.0.1-0.20200717133916-3a09287ba25e
	github.com/versent/saml2aws => ./third_party/saml2aws
)
