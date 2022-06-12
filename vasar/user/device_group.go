package user

// DeviceGroup represents a device group that a user may be in.
type DeviceGroup struct {
	device int
}

// DeviceGroupKeyboardMouse returns a device group that allows keyboard and mouse players.
func DeviceGroupKeyboardMouse() DeviceGroup {
	return DeviceGroup{device: 1}
}

// DeviceGroupMobile returns a device group that allows mobile players.
func DeviceGroupMobile() DeviceGroup {
	return DeviceGroup{device: 2}
}

// DeviceGroupController returns a device group that allows controller players.
func DeviceGroupController() DeviceGroup {
	return DeviceGroup{device: 3}
}

// DeviceGroups returns a list of all device groups.
func DeviceGroups() []DeviceGroup {
	return []DeviceGroup{
		DeviceGroupKeyboardMouse(),
		DeviceGroupMobile(),
		DeviceGroupController(),
	}
}

// String ...
func (d DeviceGroup) String() string {
	switch d.device {
	case 1:
		return "Keyboard/Mouse"
	case 2:
		return "Touch"
	case 3:
		return "Controller"
	}
	panic("should never happen")
}

// Compare ...
func (d DeviceGroup) Compare(o DeviceGroup) bool {
	return d.device == o.device
}
