package daemon

import (
	"context"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"

	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	"github.com/kubeovn/kube-ovn/pkg/ovs"
	"github.com/kubeovn/kube-ovn/pkg/util"
)

func (c *Controller) setIPSet() error {
	return nil
}

func (c *Controller) setPolicyRouting() error {
	return nil
}

func (c *Controller) setIptables() error {
	return nil
}

func (c *Controller) gcIPSet() error {
	return nil
}

func (c *Controller) addEgressConfig(subnet *kubeovnv1.Subnet, ip string) error {
	// nothing to do on Windows
	return nil
}

func (c *Controller) removeEgressConfig(subnet, ip string) error {
	// nothing to do on Windows
	return nil
}

func (c *Controller) setExGateway() error {
	node, err := c.nodesLister.Get(c.config.NodeName)
	if err != nil {
		klog.Errorf("failed to get node, %v", err)
		return err
	}
	enable := node.Labels[util.ExGatewayLabel]
	externalBridge := util.ExternalBridgeName(c.config.ExternalGatewaySwitch)
	if enable == "true" {
		cm, err := c.config.KubeClient.CoreV1().ConfigMaps(c.config.ExternalGatewayConfigNS).Get(context.Background(), util.ExternalGatewayConfig, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("failed to get ovn-external-gw-config, %v", err)
			return err
		}

		linkName, exist := cm.Data["external-gw-nic"]
		if !exist || len(linkName) == 0 {
			err = fmt.Errorf("external-gw-nic not configured in ovn-external-gw-config")
			klog.Error(err)
			return err
		}

		externalBrReady := false
		// if external nic already attached into another bridge
		if existBr, err := ovs.Exec("port-to-br", linkName); err == nil {
			if existBr == externalBridge {
				externalBrReady = true
			} else {
				klog.Infof("external bridge should change from %s to %s, delete external bridge %s", existBr, externalBridge, existBr)
				if _, err := ovs.Exec(ovs.IfExists, "del-br", existBr); err != nil {
					err = fmt.Errorf("failed to del external br %s, %v", existBr, err)
					klog.Error(err)
					return err
				}
			}
		}

		if !externalBrReady {
			if _, err := ovs.Exec(
				ovs.MayExist, "add-br", externalBridge, "--",
				ovs.MayExist, "add-port", externalBridge, linkName,
			); err != nil {
				err = fmt.Errorf("failed to enable external gateway, %v", err)
				klog.Error(err)
			}
		}
		if err = addOvnMapping("ovn-bridge-mappings", c.config.ExternalGatewaySwitch, externalBridge, true); err != nil {
			klog.Error(err)
			return err
		}
	} else {
		brExists, err := ovs.BridgeExists(externalBridge)
		if err != nil {
			return fmt.Errorf("failed to check OVS bridge existence: %v", err)
		}
		if !brExists {
			return nil
		}

		providerNetworks, err := c.providerNetworksLister.List(labels.Everything())
		if err != nil && !k8serrors.IsNotFound(err) {
			klog.Errorf("failed to list provider networks: %v", err)
			return err
		}

		for _, pn := range providerNetworks {
			// if external nic already attached into another bridge
			if existBr, err := ovs.Exec("port-to-br", pn.Spec.DefaultInterface); err == nil {
				if existBr == externalBridge {
					// delete switch after related provider network not exist
					return nil
				}
			}
		}

		keepExternalSubnet := false
		externalSubnet, err := c.subnetsLister.Get(c.config.ExternalGatewaySwitch)
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				klog.Errorf("failed to get subnet %s, %v", c.config.ExternalGatewaySwitch, err)
				return err
			}
		} else {
			if externalSubnet.Spec.Vlan != "" {
				keepExternalSubnet = true
			}
		}

		if !keepExternalSubnet {
			klog.Infof("delete external bridge %s", externalBridge)
			if _, err := ovs.Exec(
				ovs.IfExists, "del-br", externalBridge); err != nil {
				err = fmt.Errorf("failed to disable external gateway, %v", err)
				klog.Error(err)
				return err
			}
		}
	}
	return nil
}
