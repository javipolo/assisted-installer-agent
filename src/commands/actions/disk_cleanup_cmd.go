package actions

import (
	"github.com/openshift/assisted-installer-agent/src/config"
	"github.com/openshift/assisted-installer-agent/src/util"
	"github.com/openshift/assisted-service/models"
)

type diskCleanup struct {
	args        []string
	agentConfig *config.AgentConfig
}

func (a *diskCleanup) Validate() error {
	modelToValidate := models.DiskCleanupRequest{}
	err := ValidateCommon("disk cleanup", 1, a.args, &modelToValidate)
	if err != nil {
		return err
	}
	return nil
}

func (a *diskCleanup) Command() string {
	return "sh"
}

func (a *diskCleanup) Args() []string {
	arguments := []string{
		"-c",
		"id=`podman ps --quiet --filter \"name=disk_cleanup\"` ; " +
			"test ! -z \"$id\" || " +
			"podman run --privileged --rm --quiet -v /dev:/dev:rw -v /var/log:/var/log -v /run/systemd/journal/socket:/run/systemd/journal/socket " +
			"--name disk_cleanup " +
			a.agentConfig.AgentVersion + " disk_cleanup '" +
			a.args[0] + "'",
	}
	return arguments
}

func (a *diskCleanup) Run() (stdout, stderr string, exitCode int) {
	return util.ExecutePrivileged(a.Command(), a.Args()...)
}
