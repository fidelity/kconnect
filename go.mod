module github.com/fidelity/kconnect

go 1.13

require (
	github.com/AlecAivazis/survey/v2 v2.1.1
	github.com/Azure/go-ntlmssp v0.0.0-20200615164410-66371956d46c // indirect
	github.com/PuerkitoBio/goquery v1.5.1 // indirect
	github.com/andybalholm/cascadia v1.2.0 // indirect
	github.com/aws/aws-sdk-go v1.35.2
	github.com/beevik/etree v1.1.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-playground/validator/v10 v10.3.0
	github.com/golang/mock v1.3.1
	github.com/gophercloud/gophercloud v0.1.0 // indirect
	github.com/imdario/mergo v0.3.10
	github.com/kr/text v0.2.0 // indirect
	github.com/marshallbrekka/go-u2fhost v0.0.0-20200114212649-cc764c209ee9 // indirect
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.3.2 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/oklog/ulid v1.3.1
	github.com/oklog/ulid/v2 v2.0.2 // indirect
	github.com/onsi/gomega v1.10.1
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/spf13/afero v1.3.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/tidwall/gjson v1.6.0 // indirect
	github.com/tidwall/pretty v1.0.1 // indirect
	github.com/versent/saml2aws v1.8.5-0.20200622110128-d94772688a70
	github.com/worr/saml2aws v2.15.0+incompatible
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899 // indirect
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	golang.org/x/sys v0.0.0-20200720211630-cb9d2d5c5666 // indirect
	golang.org/x/text v0.3.3 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.1
	k8s.io/apimachinery v0.19.1
	k8s.io/cli-runtime v0.19.1
	k8s.io/client-go v0.19.1
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kubectl v0.19.1
	sigs.k8s.io/structured-merge-diff/v2 v2.0.1 // indirect
	sigs.k8s.io/yaml v1.2.0

)

replace github.com/versent/saml2aws => ./third_party/saml2aws
