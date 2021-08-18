package conversion

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/tinkerbell/tink/k8s/api/v1alpha1"
	"github.com/tinkerbell/tink/protos/hardware"
	"k8s.io/utils/pointer"
)

func dhcpFromK8s(d *v1alpha1.DHCP) *hardware.Hardware_DHCP {
	if d == nil {
		return nil
	}
	resp := &hardware.Hardware_DHCP{
		Mac:         d.MAC,
		Hostname:    d.Hostname,
		LeaseTime:   d.LeaseTime,
		NameServers: d.NameServers,
		TimeServers: d.TimeServers,
		Arch:        d.Arch,
		Uefi:        d.UEFI,
		IfaceName:   d.IfaceName,
	}
	if d.IP != nil {
		resp.Ip = &hardware.Hardware_DHCP_IP{
			Address: d.IP.Address,
			Netmask: d.IP.Netmask,
			Gateway: d.IP.Gateway,
			Family:  d.IP.Family,
		}
	}
	return resp
}

func netbootFromK8s(nb *v1alpha1.Netboot) *hardware.Hardware_Netboot {
	if nb == nil {
		return nil
	}
	resp := &hardware.Hardware_Netboot{
		AllowPxe:      pointer.BoolDeref(nb.AllowPXE, false),
		AllowWorkflow: pointer.BoolDeref(nb.AllowWorkflow, false),
	}
	if nb.IPXE != nil {
		resp.Ipxe = &hardware.Hardware_Netboot_IPXE{
			Url:      nb.IPXE.URL,
			Contents: nb.IPXE.Contents,
		}
	}
	if nb.OSIE != nil {
		resp.Osie = &hardware.Hardware_Netboot_Osie{
			BaseUrl: nb.OSIE.BaseURL,
			Kernel:  nb.OSIE.Kernel,
			Initrd:  nb.OSIE.Initrd,
		}
	}
	return resp
}

func HardwareFromK8s(hw *v1alpha1.Hardware) *hardware.Hardware {
	ifaces := []*hardware.Hardware_Network_Interface{}
	for _, iface := range hw.Status.Interfaces {
		ifaces = append(ifaces, &hardware.Hardware_Network_Interface{
			Dhcp:    dhcpFromK8s(iface.DHCP),
			Netboot: netbootFromK8s(iface.Netboot),
		})
	}
	resp := &hardware.Hardware{
		Id:       hw.Spec.ID,
		Version:  hw.Status.TinkVersion,
		Metadata: hw.Status.TinkMetadata,
	}
	if len(ifaces) > 0 {
		resp.Network = &hardware.Hardware_Network{Interfaces: ifaces}
	}
	return resp
}

func HardwareToK8s(hw *hardware.Hardware) (*v1alpha1.Hardware, error) {
	resp := &v1alpha1.Hardware{}
	if err := reconcileUserData(resp, hw); err != nil {
		return nil, err
	}
	if err := reconcileStatus(resp, hw); err != nil {
		return nil, err
	}
	metaName := &struct {
		Instance struct {
			Hostname string `json:"hostname"`
		} `json:"instance"`
	}{}
	err := json.Unmarshal([]byte(hw.Metadata), metaName)
	if err != nil {
		return nil, err
	}
	resp.Name = metaName.Instance.Hostname
	resp.Spec.ID = hw.GetId()
	return resp, nil
}

func reconcileUserData(h *v1alpha1.Hardware, tinkHardware *hardware.Hardware) error {
	if h.Spec.UserData == nil {
		return nil
	}
	metadata := tinkHardware.GetMetadata()
	hwMetaData := make(map[string]interface{})
	if err := json.Unmarshal([]byte(metadata), &hwMetaData); err != nil {
		return fmt.Errorf("failed to unmarshal metadata from json: %w", err)
	}
	if hwMetaData["userdata"] != *h.Spec.UserData {
		hwMetaData["userdata"] = *h.Spec.UserData
		newHWMetaData, err := json.Marshal(hwMetaData)
		if err != nil {
			return fmt.Errorf("failed to marshal updated metadata to json: %w", err)
		}
		tinkHardware.Metadata = string(newHWMetaData)
	}
	return nil
}

func interfaceFromTinkInterface(iface *hardware.Hardware_Network_Interface) v1alpha1.Interface {
	tinkInterface := v1alpha1.Interface{}
	if netboot := iface.GetNetboot(); netboot != nil {
		tinkInterface.Netboot = &v1alpha1.Netboot{
			AllowPXE:      pointer.BoolPtr(netboot.GetAllowPxe()),
			AllowWorkflow: pointer.BoolPtr(netboot.GetAllowWorkflow()),
		}
		if ipxe := netboot.GetIpxe(); ipxe != nil {
			tinkInterface.Netboot.IPXE = &v1alpha1.IPXE{
				URL:      ipxe.GetUrl(),
				Contents: ipxe.GetContents(),
			}
		}

		if osie := netboot.GetOsie(); osie != nil {
			tinkInterface.Netboot.OSIE = &v1alpha1.OSIE{
				BaseURL: osie.GetBaseUrl(),
				Kernel:  osie.GetKernel(),
				Initrd:  osie.GetInitrd(),
			}
		}
	}

	if dhcp := iface.GetDhcp(); dhcp != nil {
		tinkInterface.DHCP = &v1alpha1.DHCP{
			MAC:         dhcp.GetMac(),
			Hostname:    dhcp.GetHostname(),
			LeaseTime:   dhcp.GetLeaseTime(),
			NameServers: dhcp.GetNameServers(),
			TimeServers: dhcp.GetTimeServers(),
			Arch:        dhcp.GetArch(),
			UEFI:        dhcp.GetUefi(),
			IfaceName:   dhcp.GetIfaceName(),
		}

		if ip := dhcp.GetIp(); ip != nil {
			tinkInterface.DHCP.IP = &v1alpha1.IP{
				Address: ip.GetAddress(),
				Netmask: ip.GetNetmask(),
				Gateway: ip.GetGateway(),
				Family:  ip.GetFamily(),
			}
		}
	}

	return tinkInterface
}

func reconcileStatus(h *v1alpha1.Hardware, tinkHardware *hardware.Hardware) error {
	h.Status.TinkMetadata = tinkHardware.GetMetadata()
	h.Status.TinkVersion = tinkHardware.GetVersion()
	h.Status.Interfaces = []v1alpha1.Interface{}

	for _, iface := range tinkHardware.GetNetwork().GetInterfaces() {
		tinkInterface := interfaceFromTinkInterface(iface)
		h.Status.Interfaces = append(h.Status.Interfaces, tinkInterface)
	}

	h.Status.State = v1alpha1.HardwareReady

	disks, err := disksFromMetaData(h.Status.TinkMetadata)
	if err != nil {
		return fmt.Errorf("Failed to parse disk info from metadat: %w", err)
	}

	h.Status.Disks = disks
	return nil
}

func disksFromMetaData(metadata string) ([]v1alpha1.Disk, error) {
	// Attempt to extract disk information from metadata
	hwMetaData := make(map[string]interface{})
	if err := json.Unmarshal([]byte(metadata), &hwMetaData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata from json: %w", err)
	}

	if instanceData, ok := hwMetaData["instance"]; ok {
		id := reflect.ValueOf(instanceData)
		if id.Kind() == reflect.Map && id.Type().Key().Kind() == reflect.String {
			storage := reflect.ValueOf(id.MapIndex(reflect.ValueOf("storage")).Interface())
			if storage.Kind() == reflect.Map && storage.Type().Key().Kind() == reflect.String {
				return parseDisks(storage.MapIndex(reflect.ValueOf("disks")).Interface()), nil
			}
		}
	}

	return nil, nil
}

func parseDisks(disks interface{}) []v1alpha1.Disk {
	d := reflect.ValueOf(disks)
	if d.Kind() == reflect.Slice {
		foundDisks := make([]v1alpha1.Disk, 0, d.Len())

		for i := 0; i < d.Len(); i++ {
			disk := reflect.ValueOf(d.Index(i).Interface())
			if disk.Kind() == reflect.Map && disk.Type().Key().Kind() == reflect.String {
				device := reflect.ValueOf(disk.MapIndex(reflect.ValueOf("device")).Interface())
				if device.Kind() == reflect.String {
					foundDisks = append(foundDisks, v1alpha1.Disk{Device: device.String()})
				}
			}
		}

		return foundDisks
	}

	return nil
}
