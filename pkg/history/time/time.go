package time

import (
	"fmt"
	"time"

	"github.com/fidelity/kconnect/pkg/aws/awsconfig"
	"gopkg.in/ini.v1"
)

func GetExpireTimeFromAWSCredentials(profileName string) (time.Time, error) {

	path, err := awsconfig.LocateConfigFile()
	if err != nil {
		return time.Time{}, err
	}
	cfg, err := ini.Load(path)
	if err != nil {
		return time.Time{}, err
	}
	tokenTime := cfg.Section(profileName).Key("x_security_token_expires").String()
	return time.Parse(time.RFC3339, tokenTime)

}

func GetRemainingTime(expiresAt time.Time) string {

	timeRemaining := expiresAt.Sub(time.Now())
	timeRemaining = timeRemaining.Round(time.Second)
	if timeRemaining.Seconds() < 0 {
		timeRemaining = 0
	}
	return fmt.Sprintf("%s", timeRemaining)
}
