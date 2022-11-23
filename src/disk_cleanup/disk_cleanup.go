package disk_cleanup

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/openshift/assisted-installer-agent/src/config"
	"github.com/openshift/assisted-service/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type DiskCleanup struct {
	dependencies      IDependencies
	subprocessConfig *config.SubprocessConfig
}

func NewDiskCleanup(subprocessConfig *config.SubprocessConfig, dependencies IDependencies) *DiskCleanup {
	return &DiskCleanup{dependencies: dependencies, subprocessConfig: subprocessConfig}
}

func (p *DiskCleanup) CleanupDevice(diskCleanupRequestStr string, log *logrus.Logger) (stdout string, stderr string, exitCode int) {
	var diskCleanupRequest models.DiskCleanupRequest

	if err := json.Unmarshal([]byte(diskCleanupRequestStr), &diskCleanupRequest); err != nil {
		wrapped := errors.Wrap(err, "Failed to unmarshal DiskCleanupRequest")
		log.WithError(err).Error(wrapped.Error())
		return "", wrapped.Error(), -1
	}

	if diskCleanupRequest.Path == nil {
		err := errors.New("Missing Filename in DiskCleanupRequest")
		log.WithError(err).Error(err.Error())
		return "", err.Error(), -1
	}

	// Cleanup disks
	err := p.cleanupInstallDevice(*diskCleanupRequest.Path)
	if err != nil {
		// Return an error if something failed
		errMsg := fmt.Sprintf("Failed to run diskCleanup successfuly on device %s", diskCleanupRequest.Path)
		return createResponse(false, *diskCleanupRequest.Path), errMsg, -1
	}

	// Send response
	response := createResponse(true, *diskCleanupRequest.Path)
	return response, "", 0
}

func createResponse(status bool, path string) string {
	diskCleanupResponse := models.DiskCleanupResponse{
		Successful: status,
		Path:       path,
	}
	bytes, err := json.Marshal(diskCleanupResponse)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func (p *DiskCleanup) cleanupInstallDevice(device string) error {

	// JAVI TODO: Check how to implement this
	if p.subprocessConfig.DryRunEnabled {
		// if i.DryRunEnabled || i.Config.SkipInstallationDiskCleanup {
		return nil
	}

	// In case of symlink, get real file path
	realDevicePath, err := filepath.EvalSymlinks(device)
	if err != nil {
		return errors.Wrapf(err, "Failed to get real file path of installation disk")
	}

	err = p.cleanLVM(realDevicePath)

	if err != nil {
		return err
	}

	if p.IsRaidMember(realDevicePath) {
		var raidDevices []string
		raidDevices, err = p.GetRaidDevices(realDevicePath)

		if err != nil {
			return err
		}

		for _, raidDevice := range raidDevices {
			// Cleaning the raid device itself before removing membership.
			err = p.cleanLVM(raidDevice)

			if err != nil {
				return err
			}
		}

		err = p.CleanRaidMembership(realDevicePath)

		if err != nil {
			return err
		}
	}

	return p.Wipefs(realDevicePath)
}

func (p *DiskCleanup) Wipefs(device string) error {
	// JAVI TODO: Wait for michael levy to clarify why --force runs before the simple command, and if just running the --force one can be enough
	_, stderr, exitCode := p.dependencies.Execute("wipefs", "--all", "--force", device)
	if exitCode != 0 {
		_, stderr, exitCode = p.dependencies.Execute("wipefs", "--all", device)
	}
	if exitCode != 0 {
		return errors.New(stderr)
	}
	return nil
}

// LVM stuff

func (p *DiskCleanup) cleanLVM(device string) error {
	vgNames, err := p.GetVolumeGroupsByDisk(device)
	if err != nil {
		return err
	}

	if len(vgNames) > 0 {
		err = p.removeVolumeGroupsFromDevice(vgNames, device)
		if err != nil {
			return err
		}
	}

	return p.RemoveAllPVsOnDevice(device)
}

func (p *DiskCleanup) GetVolumeGroupsByDisk(diskName string) ([]string, error) {
	var vgs []string

	stdout, stderr, exitCode := p.dependencies.Execute("vgs", "--noheadings", "-o", "vg_name,pv_name")
	if exitCode != 0 {
		return vgs, errors.Errorf("Failed to list VGs in the system: %s", stderr)
	}

	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		res := strings.Fields(line)
		if len(res) < 2 {
			continue
		}

		if strings.Contains(res[1], diskName) {
			vgs = append(vgs, res[0])
		}
	}
	return vgs, nil
}

func (p *DiskCleanup) removeVolumeGroupsFromDevice(vgNames []string, device string) error {
	for _, vgName := range vgNames {
		err := p.RemoveVG(vgName)

		if err != nil {
			return errors.Errorf("Could not delete volume group (%s) due to error: %w", vgName, err)
		}
	}
	return nil
}

func (p *DiskCleanup) getDiskPVs(diskName string) ([]string, error) {
	var pvs []string
	stdout, stderr, exitCode := p.dependencies.Execute("pvs", "--noheadings", "-o", "pv_name")
	if exitCode != 0 {
		return pvs, errors.Errorf("Failed to list PVs in the system: %s", stderr)
	}
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		if strings.Contains(line, diskName) {
			pvs = append(pvs, strings.TrimSpace(line))
		}
	}
	return pvs, nil
}

func (p *DiskCleanup) RemoveAllPVsOnDevice(diskName string) error {
	pvs, err := p.getDiskPVs(diskName)
	if err != nil {
		return err
	}
	for _, pv := range pvs {
		err = p.RemovePV(pv)
		if err != nil {
			return errors.Errorf("Failed to remove pv %s from disk %s: %w", pv, diskName, err)
		}
	}
	return nil
}

func (p *DiskCleanup) RemoveVG(vgName string) error {
	stdout, stderr, exitCode := p.dependencies.Execute("vgremove", vgName, "-y")
	if exitCode != 0 {
		return errors.Errorf("Failed to remove VG %s, output %s, error %s", vgName, stdout, stderr)
	}
	return nil
}

func (p *DiskCleanup) RemovePV(pvName string) error {
	stdout, stderr, exitCode := p.dependencies.Execute("pvremove", pvName, "-y", "-ff")
	if exitCode != 0 {
		return errors.Errorf("Failed to remove PV %s, output %s, error %s", pvName, stdout, stderr)
	}
	return nil
}

// RAID Stuff

func (p *DiskCleanup) IsRaidMember(device string) bool {
	raidDevices, err := p.getRaidDevices2Members()

	if err != nil {
		// Error occurred while trying to get list of raid devices - continue without cleaning
		return false
	}

	// The device itself or one of its partitions
	expression, _ := regexp.Compile(device + "[\\d]*")

	for _, raidArrayMembers := range raidDevices {
		for _, raidMember := range raidArrayMembers {
			if expression.MatchString(raidMember) {
				return true
			}
		}
	}

	return false
}

func (p *DiskCleanup) CleanRaidMembership(device string) error {
	raidDevices, err := p.getRaidDevices2Members()

	if err != nil {
		return err
	}

	for raidDeviceName, raidArrayMembers := range raidDevices {
		err = p.removeDeviceFromRaidArray(device, raidDeviceName, raidArrayMembers)

		if err != nil {
			return err
		}
	}

	return nil
}

func (p *DiskCleanup) GetRaidDevices(deviceName string) ([]string, error) {
	raidDevices, err := p.getRaidDevices2Members()
	var result []string

	if err != nil {
		return result, err
	}

	for raidDeviceName, raidArrayMembers := range raidDevices {
		expression, _ := regexp.Compile(deviceName + "[\\d]*")

		for _, raidMember := range raidArrayMembers {
			// A partition or the device itself is part of the raid array.
			if expression.MatchString(raidMember) {
				result = append(result, raidDeviceName)
				break
			}
		}
	}

	return result, nil
}

func (p *DiskCleanup) getRaidDevices2Members() (map[string][]string, error) {
	stdout, stderr, exitCode := p.dependencies.Execute("mdadm", "-v", "--query", "--detail", "--scan")
	if exitCode != 0 {
		return nil, errors.Errorf("Error listing raid devices: %s", stderr)
	}

	lines := strings.Split(stdout, "\n")
	result := make(map[string][]string)

	/*
		The output pattern is:
		ARRAY /dev/md0 level=raid1 num-devices=2 metadata=1.2 name=0 UUID=77e1b6f2:56530ebd:38bd6808:17fd01c4
		   devices=/dev/vda2,/dev/vda3
		ARRAY /dev/md1 level=raid1 num-devices=1 metadata=1.2 name=1 UUID=aad7aca9:81db82f3:2f1fedb1:f89ddb43
		   devices=/dev/vda1
	*/
	for i := 0; i < len(lines); {
		if !strings.Contains(lines[i], "ARRAY") {
			i++
			continue
		}

		fields := strings.Fields(lines[i])
		// In case of symlink, get real file path
		raidDeviceName, err := filepath.EvalSymlinks(fields[1])
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to get real file path of RAID device")
		}

		i++

		// Ensuring that we have at least two lines per device.
		if len(lines) == i {
			break
		}

		raidArrayMembersStr := strings.TrimSpace(lines[i])
		prefix := "devices="

		if !strings.HasPrefix(raidArrayMembersStr, prefix) {
			continue
		}

		raidArrayMembersStr = raidArrayMembersStr[len(prefix):]
		result[raidDeviceName] = strings.Split(raidArrayMembersStr, ",")
		i++
	}

	return result, nil
}

func (p *DiskCleanup) removeDeviceFromRaidArray(deviceName string, raidDeviceName string, raidArrayMembers []string) error {
	raidStopped := false

	expression, _ := regexp.Compile(deviceName + "[\\d]*")

	for _, raidMember := range raidArrayMembers {
		// A partition or the device itself is part of the raid array.
		if expression.MatchString(raidMember) {
			// Stop the raid device.
			if !raidStopped {
				_, stderr, exitCode := p.dependencies.Execute("mdadm", "--stop", raidDeviceName)
				if exitCode != 0 {
					return errors.Errorf("Error stopping raid device %s: %s", raidDeviceName, stderr)
				}
				raidStopped = true
			}

			// Clean the raid superblock from the device
			_, stderr, exitCode := p.dependencies.Execute("mdadm", "--zero-superblock", raidMember)
			if exitCode != 0 {
				return errors.Errorf("Error cleaning raid member %s superblock: %s", raidMember, stderr)
			}
		}
	}
	return nil
}
