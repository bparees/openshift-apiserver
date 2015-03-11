package simple

import (
	"errors"
	"fmt"

	"code.google.com/p/go-uuid/uuid"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/golang/glog"

	routeapi "github.com/openshift/origin/pkg/route/api"
)

// Default DNS suffix to use if no configuration is passed to this plugin.
// Would be better if we could use "v3.openshift.app", someone bought that!
const defaultDNSSuffix = "v3.openshift.com"

// SimpleAllocationPlugin implements the route.AllocationPlugin interface
// to provide a simple unsharded (or single sharded) allocation plugin.
type SimpleAllocationPlugin struct {
	DNSSuffix string
}

// Creates a new SimpleAllocationPlugin.
func NewSimpleAllocationPlugin(suffix string) (*SimpleAllocationPlugin, error) {
	if len(suffix) == 0 {
		suffix = defaultDNSSuffix
	}

	glog.V(4).Infof("NewSimpleAllocationPlugin: suffix=%s", suffix)

	// Check that the DNS suffix is valid.
	if !util.IsDNSSubdomain(suffix) {
		errmsg := fmt.Sprintf("invalid DNS suffix: %s", suffix)
		glog.Errorf("NewSimpleAllocationPlugin: %s", errmsg)
		return nil, errors.New(errmsg)
	}

	return &SimpleAllocationPlugin{DNSSuffix: suffix}, nil
}

// Allocate a router shard for the given route. This plugin always returns
// the "global" router shard.
func (p *SimpleAllocationPlugin) Allocate(route *routeapi.Route) (*routeapi.RouterShard, error) {

	glog.V(4).Infof("SimpleAllocationPlugin: Allocating global shard *.%s to Route: %s",
		p.DNSSuffix, route.ServiceName)

	return &routeapi.RouterShard{ShardName: "global", DNSSuffix: p.DNSSuffix}, nil
}

// Generate a host name for a route - using the service name,
// namespace (if provided) and the router shard dns suffix.
func (p *SimpleAllocationPlugin) GenerateHostname(route *routeapi.Route, shard *routeapi.RouterShard) string {

	name := route.ServiceName
	if len(name) == 0 {
		name = uuid.NewUUID().String()
		glog.V(4).Infof("SimpleAllocationPlugin: No service name passed, using generated name: %s", name)
	}

	s := ""
	if len(route.Namespace) <= 0 {
		s = fmt.Sprintf("%s.%s", name, shard.DNSSuffix)
	} else {
		s = fmt.Sprintf("%s-%s.%s", name, route.Namespace, shard.DNSSuffix)
	}

	glog.V(4).Infof("SimpleAllocationPlugin: Generated hostname=%s for Route: %s", s, route.ServiceName)

	return s
}
