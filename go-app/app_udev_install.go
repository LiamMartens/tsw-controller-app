package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"tsw_controller_app/logger"
)

// InstallUdevRules prompts the user for elevated privileges and installs
// a udev rule so the controller HID device is accessible without root.
// Only runs on Linux. Returns an error if pkexec/polkit isn't available.
func (a *App) InstallUdevRules() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("udev rules are only applicable on Linux")
	}

	udev_rule_path := "/etc/udev/rules.d/99-tsw-controller-udev.rules"
	/* potentially add vendor ID/product ID/serial */
	udev_rule_content := `# TSW Controller Utility — HID device access
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="1209", ATTRS{idProduct}=="5389", TAG+="uaccess"
SUBSYSTEM=="usb", ATTRS{idVendor}=="1209", ATTRS{idProduct}=="5389", TAG+="uaccess"

SUBSYSTEM=="hidraw", ATTRS{idVendor}=="04d8", ATTRS{idProduct}=="e43b", TAG+="uaccess"
SUBSYSTEM=="usb", ATTRS{idVendor}=="04d8", ATTRS{idProduct}=="e43b", TAG+="uaccess"

SUBSYSTEM=="hidraw", ATTRS{idVendor}=="044f", ATTRS{idProduct}=="040a", TAG+="uaccess"
SUBSYSTEM=="usb", ATTRS{idVendor}=="044f", ATTRS{idProduct}=="040a", TAG+="uaccess"`

	tmp_dir := os.TempDir()
	tmp_file, err := os.CreateTemp(tmp_dir, "tsw-udev-rule-*.rules")
	if err != nil {
		return fmt.Errorf("could not create temporary udev file: %w", err)
	}
	defer os.Remove(tmp_file.Name())

	if _, err := tmp_file.WriteString(udev_rule_content); err != nil {
		tmp_file.Close()
		return fmt.Errorf("could not write temp file: %w", err)
	}

	tmp_file.Close()
	install_udev_rule_script := fmt.Sprintf(`cp "%s" "%s" && chmod 644 "%s" && udevadm control --reload && udevadm trigger`, tmp_file.Name(), udev_rule_path, udev_rule_path)
	install_cmd := exec.Command("pkexec", "bash", "-c", install_udev_rule_script)
	install_cmd.Stdout = os.Stdout
	install_cmd.Stderr = os.Stderr

	if err := install_cmd.Run(); err != nil {
		return fmt.Errorf("pkexec failed (is polkit installed?): %w", err)
	}

	logger.Logger.Info("[App] udev rules installed successfully")
	return nil
}
