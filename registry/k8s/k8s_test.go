package k8s

import (
	"context"
	"fmt"
	"testing"
)

func TestK8sDiscovery(t *testing.T) {
	// discovery:///svc-name.test.svc.cluster_name
	registry := NewRegistry(Token("token"))
	w, err := registry.Discover(context.TODO(), "svc-name.test.svc.cluster_name:9090")
	if err != nil {
		fmt.Printf("registery discover err:%+v\n", err)
		return
	}
	_, err = w.List()
	if err != nil {
		fmt.Printf("registery list err:%+v\n", err)
		return
	}
}

// windows设置host(api server地址) 192.168.xx.xx xxxx.ccs.tencent-cloud.com
