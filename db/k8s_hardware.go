package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/tinkerbell/tink/k8s/api"
	"github.com/tinkerbell/tink/k8s/api/v1alpha1"
	"github.com/tinkerbell/tink/k8s/conversion"
	tinkhw "github.com/tinkerbell/tink/protos/hardware"
)

const (
	ipIndexKey  string = "ip"
	idIndexKey  string = "id"
	macIndexKey string = "mac"
)

// func for indexing by Mac Addr
func hwMacIndexFunc(obj interface{}) ([]string, error) {
	hw, ok := obj.(*v1alpha1.Hardware)
	if !ok {
		return []string{}, nil
	}
	keys := []string{}
	for _, iface := range hw.Status.Interfaces {
		if iface.DHCP == nil {
			continue
		}
		if iface.DHCP.MAC != "" {
			keys = append(keys, iface.DHCP.MAC)
		}
	}
	return keys, nil
}

// func for indexing by IP Addr
func hwIpIndexFunc(obj interface{}) ([]string, error) {
	hw, ok := obj.(*v1alpha1.Hardware)
	if !ok {
		return []string{}, nil
	}
	keys := []string{}
	for _, iface := range hw.Status.Interfaces {
		if iface.DHCP == nil {
			continue
		}
		if iface.DHCP.IP == nil {
			continue
		}
		if iface.DHCP.IP.Address != "" {
			keys = append(keys, iface.DHCP.IP.Address)
		}
	}
	return keys, nil
}

// func for indexing by ID Addr
func hwIDIndexFunc(obj interface{}) ([]string, error) {
	hw, ok := obj.(*v1alpha1.Hardware)
	if !ok {
		return []string{}, nil
	}
	return []string{hw.TinkID()}, nil
}

func NewHardwareIndexerInformer(clientset api.TinkerbellV1Alpha1Interface) cache.Indexer {
	hwIndexer, hwController := cache.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return clientset.Hardware().List(context.Background(), lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return clientset.Hardware().Watch(context.Background(), lo)
			},
		},
		&v1alpha1.Hardware{},
		1*time.Minute,
		cache.ResourceEventHandlerFuncs{},
		cache.Indexers{
			macIndexKey: hwMacIndexFunc,
			ipIndexKey:  hwIpIndexFunc,
			idIndexKey:  hwIDIndexFunc,
		},
	)
	go hwController.Run(wait.NeverStop)
	return hwIndexer
}

// Delete hardware
func (d K8sDB) DeleteFromDB(ctx context.Context, id string) error {
	keys, err := d.hwIndexer.IndexKeys(idIndexKey, id)
	if err != nil {
		return err
	}
	hws := make([]*v1alpha1.Hardware, 0)
	for _, key := range keys {
		obj, exists, err := d.hwIndexer.GetByKey(key)
		if err != nil {
			return nil
		}
		if !exists {
			continue
		}
		hws = append(hws, obj.(*v1alpha1.Hardware))
	}
	if len(hws) > 1 {
		names := []string{}
		for _, hw := range hws {
			names = append(names, hw.Name)
		}
		return fmt.Errorf("found %d hardware with the same ID. Hardware Names: %v", len(hws), names)
	}
	if len(hws) == 0 {
		return nil
	}
	// TODO: This is pretty infrequent, force resync?
	return d.k8sClient.Hardware().Delete(ctx, hws[0].Name, metav1.DeleteOptions{})
}

// Add Hardware
func (d K8sDB) InsertIntoDB(ctx context.Context, data string) error {
	tinkHw := &tinkhw.Hardware{}
	err := json.Unmarshal([]byte(data), tinkHw)
	if err != nil {
		d.logger.Error(err)
		return err
	}

	hw, err := conversion.HardwareToK8s(tinkHw)
	if err != nil {
		d.logger.Error(err)
		return err
	}

	_, exists, err := d.hwIndexer.Get(hw)
	if err != nil {
		d.logger.Error(err)
		return err
	}
	if !exists {
		_, err := d.k8sClient.Hardware().Create(ctx, hw, metav1.CreateOptions{})
		return err
	}
	_, err = d.k8sClient.Hardware().Update(ctx, hw, metav1.UpdateOptions{})
	return err
}

// Get hardware by mac
func (d K8sDB) GetByMAC(ctx context.Context, mac string) (string, error) {
	keys, err := d.hwIndexer.IndexKeys(macIndexKey, mac)
	if err != nil {
		return "", err
	}
	hws := make([]*v1alpha1.Hardware, 0)
	for _, key := range keys {
		obj, exists, err := d.hwIndexer.GetByKey(key)
		if err != nil {
			return "", err
		}
		if !exists {
			continue
		}
		hws = append(hws, obj.(*v1alpha1.Hardware))
	}
	if len(hws) > 1 {
		names := []string{}
		for _, hw := range hws {
			names = append(names, hw.Name)
		}
		return "", fmt.Errorf("found %d hardware with the same IP. Hardware Names: %v", len(hws), names)
	}
	if len(hws) == 0 {
		return "", nil
	}
	output, err := json.Marshal(conversion.HardwareFromK8s(hws[0]))
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// Get hardware by IP
func (d K8sDB) GetByIP(ctx context.Context, ip string) (string, error) {
	keys, err := d.hwIndexer.IndexKeys(ipIndexKey, ip)
	if err != nil {
		return "", err
	}
	hws := make([]*v1alpha1.Hardware, 0)
	for _, key := range keys {
		obj, exists, err := d.hwIndexer.GetByKey(key)
		if err != nil {
			return "", err
		}
		if !exists {
			continue
		}
		hws = append(hws, obj.(*v1alpha1.Hardware))
	}
	if len(hws) > 1 {
		names := []string{}
		for _, hw := range hws {
			names = append(names, hw.Name)
		}
		return "", fmt.Errorf("found %d hardware with the same IP. Hardware Names: %v", len(hws), names)
	}
	if len(hws) == 0 {
		return "", nil
	}
	output, err := json.Marshal(conversion.HardwareFromK8s(hws[0]))
	if err != nil {
		d.logger.Error(err, "error in GetByIP marshal")
		return "", err
	}
	return string(output), nil
}

func ifaceSliceToHw(objs []interface{}) []*v1alpha1.Hardware {
	hws := make([]*v1alpha1.Hardware, 0)
	for _, o := range objs {
		hws = append(hws, o.(*v1alpha1.Hardware))
	}
	return hws
}

// Get hardware by ID
func (d K8sDB) GetByID(ctx context.Context, id string) (string, error) {
	keys, err := d.hwIndexer.IndexKeys(idIndexKey, id)
	if err != nil {
		return "", err
	}
	hws := make([]*v1alpha1.Hardware, 0)
	for _, key := range keys {
		obj, exists, err := d.hwIndexer.GetByKey(key)
		if err != nil {
			return "", err
		}
		if !exists {
			continue
		}
		hws = append(hws, obj.(*v1alpha1.Hardware))
	}
	if len(hws) > 1 {
		names := []string{}
		for _, hw := range hws {
			names = append(names, hw.Name)
		}
		return "", fmt.Errorf("found %d hardware with the same ID. Hardware Names: %v", len(hws), names)
	}
	if len(hws) == 0 {
		return "", nil
	}
	output, err := json.Marshal(conversion.HardwareFromK8s(hws[0]))
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// Get all hardware
func (d K8sDB) GetAll(fn func([]byte) error) error {
	hws := ifaceSliceToHw(d.hwIndexer.List())

	for _, hw := range hws {
		content, err := json.Marshal(conversion.HardwareFromK8s(hw))
		if err != nil {
			d.logger.Error(err)
			return err
		}
		if err = fn(content); err != nil {
			d.logger.Error(err)
			return err
		}
	}

	d.logger.Info(fmt.Sprintf("Returned %d hardwares", len(hws)))
	return nil
}
