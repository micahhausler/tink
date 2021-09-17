/*
Copyright 2020 The Kubernetes Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HardwareState represents the hardware state.
type HardwareState string

const (
	// HardwareError represents hardware that is in an error state.
	HardwareError = HardwareState("Error")

	// HardwareReady represents hardware that is in a ready state.
	HardwareReady = HardwareState("Ready")
)

type Metadata_Manufacturer struct {
	Id   string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Slug string `protobuf:"bytes,2,opt,name=slug,proto3" json:"slug,omitempty"`
}

type Metadata_Instance_OperatingSystem struct {
	Slug     string `protobuf:"bytes,1,opt,name=slug,proto3" json:"slug,omitempty"`
	Distro   string `protobuf:"bytes,2,opt,name=distro,proto3" json:"distro,omitempty"`
	Version  string `protobuf:"bytes,3,opt,name=version,proto3" json:"version,omitempty"`
	ImageTag string `protobuf:"bytes,4,opt,name=image_tag,json=imageTag,proto3" json:"image_tag,omitempty"`
	OsSlug   string `protobuf:"bytes,5,opt,name=os_slug,json=osSlug,proto3" json:"os_slug,omitempty"`
}

type Metadata_Instance_IP struct {
	Address    string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Netmask    string `protobuf:"bytes,2,opt,name=netmask,proto3" json:"netmask,omitempty"`
	Gateway    string `protobuf:"bytes,3,opt,name=gateway,proto3" json:"gateway,omitempty"`
	Family     int64  `protobuf:"varint,4,opt,name=family,proto3" json:"family,omitempty"`
	Public     bool   `protobuf:"varint,5,opt,name=public,proto3" json:"public,omitempty"`
	Management bool   `protobuf:"varint,6,opt,name=management,proto3" json:"management,omitempty"`
}

type Metadata_Instance_Storage_Disk_Partition struct {
	Label    string `protobuf:"bytes,1,opt,name=label,proto3" json:"label,omitempty"`
	Number   int64  `protobuf:"varint,2,opt,name=number,proto3" json:"number,omitempty"`
	Size     int64  `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	Start    int64  `protobuf:"varint,4,opt,name=start,proto3" json:"start,omitempty"`
	TypeGuid string `protobuf:"bytes,5,opt,name=type_guid,json=typeGuid,proto3" json:"type_guid,omitempty"`
}

type Metadata_Instance_Storage_RAID struct {
	Name    string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Level   string   `protobuf:"bytes,2,opt,name=level,proto3" json:"level,omitempty"`
	Devices []string `protobuf:"bytes,3,rep,name=devices,proto3" json:"devices,omitempty"`
	Spare   int64    `protobuf:"varint,4,opt,name=spare,proto3" json:"spare,omitempty"`
}

type Metadata_Instance_Storage_File struct {
	Path     string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	Contents string `protobuf:"bytes,2,opt,name=contents,proto3" json:"contents,omitempty"`
	Mode     int64  `protobuf:"varint,3,opt,name=mode,proto3" json:"mode,omitempty"`
	Uid      int64  `protobuf:"varint,4,opt,name=uid,proto3" json:"uid,omitempty"`
	Gid      int64  `protobuf:"varint,5,opt,name=gid,proto3" json:"gid,omitempty"`
}

type Metadata_Instance_Storage_Mount_FilesystemOptions struct {
	Force   bool     `protobuf:"varint,1,opt,name=force,proto3" json:"force,omitempty"`
	Options []string `protobuf:"bytes,2,rep,name=options,proto3" json:"options,omitempty"`
}

type Metadata_Instance_Storage_Mount struct {
	Device string                                             `protobuf:"bytes,1,opt,name=device,proto3" json:"device,omitempty"`
	Format string                                             `protobuf:"bytes,2,opt,name=format,proto3" json:"format,omitempty"`
	Files  []*Metadata_Instance_Storage_File                  `protobuf:"bytes,3,rep,name=files,proto3" json:"files,omitempty"`
	Create *Metadata_Instance_Storage_Mount_FilesystemOptions `protobuf:"bytes,4,opt,name=create,proto3" json:"create,omitempty"`
	Point  string                                             `protobuf:"bytes,5,opt,name=point,proto3" json:"point,omitempty"`
}

type Metadata_Instance_Storage_Filesystem struct {
	Mount *Metadata_Instance_Storage_Mount `protobuf:"bytes,1,opt,name=mount,proto3" json:"mount,omitempty"`
}

type Metadata_Instance_Storage_Disk struct {
	Device     string                                      `protobuf:"bytes,1,opt,name=device,proto3" json:"device,omitempty"`
	WipeTable  bool                                        `protobuf:"varint,2,opt,name=wipe_table,json=wipeTable,proto3" json:"wipe_table,omitempty"`
	Partitions []*Metadata_Instance_Storage_Disk_Partition `protobuf:"bytes,3,rep,name=partitions,proto3" json:"partitions,omitempty"`
}

type Metadata_Instance_Storage struct {
	Disks       []*Metadata_Instance_Storage_Disk       `protobuf:"bytes,1,rep,name=disks,proto3" json:"disks,omitempty"`
	Raid        []*Metadata_Instance_Storage_RAID       `protobuf:"bytes,2,rep,name=raid,proto3" json:"raid,omitempty"`
	Filesystems []*Metadata_Instance_Storage_Filesystem `protobuf:"bytes,3,rep,name=filesystems,proto3" json:"filesystems,omitempty"`
}

type Metadata_Instance struct {
	Id                  string                             `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	State               string                             `protobuf:"bytes,2,opt,name=state,proto3" json:"state,omitempty"`
	Hostname            string                             `protobuf:"bytes,3,opt,name=hostname,proto3" json:"hostname,omitempty"`
	AllowPxe            bool                               `protobuf:"varint,4,opt,name=allow_pxe,json=allowPxe,proto3" json:"allow_pxe,omitempty"`
	Rescue              bool                               `protobuf:"varint,5,opt,name=rescue,proto3" json:"rescue,omitempty"`
	OperatingSystem     *Metadata_Instance_OperatingSystem `protobuf:"bytes,6,opt,name=operating_system,json=operatingSystem,proto3" json:"operating_system,omitempty"`
	AlwaysPxe           bool                               `protobuf:"varint,7,opt,name=always_pxe,json=alwaysPxe,proto3" json:"always_pxe,omitempty"`
	IpxeScriptUrl       string                             `protobuf:"bytes,8,opt,name=ipxe_script_url,json=ipxeScriptUrl,proto3" json:"ipxe_script_url,omitempty"`
	Ips                 []*Metadata_Instance_IP            `protobuf:"bytes,9,rep,name=ips,proto3" json:"ips,omitempty"`
	Userdata            string                             `protobuf:"bytes,10,opt,name=userdata,proto3" json:"userdata,omitempty"`
	CryptedRootPassword string                             `protobuf:"bytes,11,opt,name=crypted_root_password,json=cryptedRootPassword,proto3" json:"crypted_root_password,omitempty"`
	Tags                []string                           `protobuf:"bytes,12,rep,name=tags,proto3" json:"tags,omitempty"`
	Storage             *Metadata_Instance_Storage         `protobuf:"bytes,13,opt,name=storage,proto3" json:"storage,omitempty"`
	SshKeys             []string                           `protobuf:"bytes,14,rep,name=ssh_keys,json=sshKeys,proto3" json:"ssh_keys,omitempty"`
	NetworkReady        bool                               `protobuf:"varint,15,opt,name=network_ready,json=networkReady,proto3" json:"network_ready,omitempty"`
}

type Metadata_Custom struct {
	PreinstalledOperatingSystemVersion *Metadata_Instance_OperatingSystem `protobuf:"bytes,1,opt,name=preinstalled_operating_system_version,json=preinstalledOperatingSystemVersion,proto3" json:"preinstalled_operating_system_version,omitempty"`
	PrivateSubnets                     []string                           `protobuf:"bytes,2,rep,name=private_subnets,json=privateSubnets,proto3" json:"private_subnets,omitempty"`
}

type Metadata_Facility struct {
	PlanSlug        string `protobuf:"bytes,1,opt,name=plan_slug,json=planSlug,proto3" json:"plan_slug,omitempty"`
	PlanVersionSlug string `protobuf:"bytes,2,opt,name=plan_version_slug,json=planVersionSlug,proto3" json:"plan_version_slug,omitempty"`
	FacilityCode    string `protobuf:"bytes,3,opt,name=facility_code,json=facilityCode,proto3" json:"facility_code,omitempty"`
}

type HardwareMetadata struct {
	State        string                 `protobuf:"bytes,1,opt,name=state,proto3" json:"state,omitempty"`
	BondingMode  int64                  `protobuf:"varint,2,opt,name=bonding_mode,json=bondingMode,proto3" json:"bonding_mode,omitempty"`
	Manufacturer *Metadata_Manufacturer `protobuf:"bytes,3,opt,name=manufacturer,proto3" json:"manufacturer,omitempty"`
	Instance     *Metadata_Instance     `protobuf:"bytes,4,opt,name=instance,proto3" json:"instance,omitempty"`
	Custom       *Metadata_Custom       `protobuf:"bytes,5,opt,name=custom,proto3" json:"custom,omitempty"`
	Facility     *Metadata_Facility     `protobuf:"bytes,6,opt,name=facility,proto3" json:"facility,omitempty"`
}

// HardwareSpec defines the desired state of Hardware.
type HardwareSpec struct {

	//+optional
	Interfaces []Interface `json:"interfaces,omitempty"`

	//+optional
	// Metadata string `json:"metadata,omitempty"`

	//+optional
	Metadata *HardwareMetadata `json:"metadata,omitempty"`

	//+optional
	TinkVersion int64 `json:"tinkVersion,omitempty"`

	//+optional
	Disks []Disk `json:"disks,omitempty"`

	// UserData is the user data to configure in the hardware's
	// metadata
	//+optional
	UserData *string `json:"userData,omitempty"`
}

// HardwareStatus defines the observed state of Hardware.
type HardwareStatus struct {
	//+optional
	State HardwareState `json:"state,omitempty"`
}

// Disk represents a disk device for Tinkerbell Hardware.
type Disk struct {
	//+optional
	Device string `json:"device,omitempty"`
}

// Interface represents a network interface configuration for Hardware.
type Interface struct {
	//+optional
	Netboot *Netboot `json:"netboot,omitempty"`

	//+optional
	DHCP *DHCP `json:"dhcp,omitempty"`
}

// Netboot configuration.
type Netboot struct {
	//+optional
	AllowPXE *bool `json:"allowPXE,omitempty"`

	//+optional
	AllowWorkflow *bool `json:"allowWorkflow,omitempty"`

	//+optional
	IPXE *IPXE `json:"ipxe,omitempty"`

	//+optional
	OSIE *OSIE `json:"osie,omitempty"`
}

// IPXE configuration.
type IPXE struct {
	URL      string `json:"url,omitempty"`
	Contents string `json:"contents,omitempty"`
}

// OSIE configuration.
type OSIE struct {
	BaseURL string `json:"baseURL,omitempty"`
	Kernel  string `json:"kernel,omitempty"`
	Initrd  string `json:"initrd,omitempty"`
}

// DHCP configuration.
type DHCP struct {
	MAC         string   `json:"mac,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	LeaseTime   int64    `json:"lease_time,omitempty"`
	NameServers []string `json:"name_servers,omitempty"`
	TimeServers []string `json:"time_servers,omitempty"`
	Arch        string   `json:"arch,omitempty"`
	UEFI        bool     `json:"uefi,omitempty"`
	IfaceName   string   `json:"iface_name,omitempty"`
	IP          *IP      `json:"ip,omitempty"`
}

// IP configuration.
type IP struct {
	Address string `json:"address,omitempty"`
	Netmask string `json:"netmask,omitempty"`
	Gateway string `json:"gateway,omitempty"`
	Family  int64  `json:"family,omitempty"`
}

// +kubebuilder:subresource:status
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=hardware,scope=Cluster,categories=tinkerbell,singular=hardware
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:JSONPath=".status.state",name=State,type=string

// Hardware is the Schema for the Hardware API.
type Hardware struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HardwareSpec   `json:"spec,omitempty"`
	Status HardwareStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HardwareList contains a list of Hardware.
type HardwareList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Hardware `json:"items"`
}

//nolint:gochecknoinits
func init() {
	SchemeBuilder.Register(&Hardware{}, &HardwareList{})
}
