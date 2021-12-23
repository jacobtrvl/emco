/// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package istiosubc

import (
	"encoding/json"
	"net"
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/project-emco/core/emco-base/src/dtc/pkg/module"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	clusterPkg "gitlab.com/project-emco/core/emco-base/src/clm/pkg/cluster"
	log "gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/logutils"
	pkgerrors "github.com/pkg/errors"
)

const sesubnet string = "240.0.0.1"

type clusterData struct {
	Reslist          []map[string][]byte //resname: res
	ClusterName      string
	GwAddress        string
	GwExternalPort   uint32
	GwInternalPort   uint32
}
type client struct {
	ClientName        string
	ClientServiceName string
	InstallClientRes  bool
	ClusterData       []clusterData
}
type serverData struct {
	AppName		 string
	ServiceName      string
	ClusterData	 []clusterData
	Clients		 []client
	InstallServerRes bool
}

// Action applies the supplied intent against the given AppContext ID
func UpdateAppContext(intentName, appContextId string) error {
	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(appContextId)
	if err != nil {
		log.Error("Error loading AppContext", log.Fields{
			"error": err,
		})
		return pkgerrors.Wrapf(err, "Error loading AppContext with Id: %v", appContextId)
	}

	caMeta, err := ac.GetCompositeAppMeta()
	if err != nil {
		log.Error("Error getting metadata from AppContext", log.Fields{
			"error": err,
		})
		return pkgerrors.Wrapf(err, "Error getting metadata from AppContext with Id: %v", appContextId)
	}

	project := caMeta.Project
	compositeapp := caMeta.CompositeApp
	compositeappversion := caMeta.Version
	deployIntentGroup := caMeta.DeploymentIntentGroup
	namespace := caMeta.Namespace

	// Get all server inbound intents
	iss, err := module.NewServerInboundIntentClient().GetServerInboundIntents(project, compositeapp, compositeappversion, deployIntentGroup, intentName)
	if err != nil {
		log.Error("Error getting server inbound intents", log.Fields{
			"error": err,
		})
		return pkgerrors.Wrapf(err, "Error getting server inbound intents %v for %v/%v%v/%v not found", intentName, project, compositeapp, deployIntentGroup, compositeappversion)
	}

	l := len(iss)
	servers := make([]serverData, l)
	index := 0
	sa := strings.Split(sesubnet, ".")
	if len(sa) != 4 {
		log.Error("Invalid subnet string", log.Fields{})
		return pkgerrors.Wrapf(err, "Invalid subnet string")
	}
	var b [4]byte
	for i:=0; i<len(sa); i++ {
		iv, _ := strconv.Atoi(sa[i])
		b[i] = byte(iv)
	}
	ips := newIP{Ip: net.IP{b[0], b[1], b[2], b[3]},}
	for _, is := range iss {
		if(is.Spec.ServiceMesh != "istio") {
			log.Error("Error ISTIO not enabled for this server", log.Fields{
				"error": err,
				"app name": is.Spec.AppName,
			})
			return pkgerrors.Wrapf(err, "Error ISTIO not enabled for this server")
		}
		clusters, err := ac.GetClusterNames(is.Spec.AppName)
		if err != nil {
			log.Error("Error retrieving clusters from App Context", log.Fields{
				"error":    err,
				"app name": is.Spec.AppName,
			})
			return pkgerrors.Wrapf(err,
				"Error retrieving clusters from App Context for app %v", is.Spec.AppName)
		}

		servers[index].AppName = is.Spec.AppName
		servers[index].ServiceName = is.Spec.ServiceName
		lc := len(clusters)
		servers[index].ClusterData = make([]clusterData, lc)
		for ci, c := range clusters {
			obj, err := getClusterKvPair(c, "istioingressgatewayaddress")
			if err != nil {
				log.Error("Error getting istio ingress gateway address", log.Fields{
					"error":    err,
				})
				return pkgerrors.Wrapf(err,
					"Error getting istio ingress gateway address")
			}
			servers[index].ClusterData[ci].GwAddress = obj
			obj, err = getClusterKvPair(c, "istioingressgatewayport")
			if err != nil {
				log.Error("Error getting istio ingress gateway port", log.Fields{
					"error":    err,
				})
				return pkgerrors.Wrapf(err,
					"Error getting istio ingress gateway port")
			}
			port, err := strconv.Atoi(obj)
			if err != nil {
				log.Error("Error converting port from string to uint32", log.Fields{
					"error":    err,
				})
				return pkgerrors.Wrapf(err,
					"Error converting port from string to uint32")
			}
			servers[index].ClusterData[ci].GwExternalPort = uint32(port)
			obj, err = getClusterKvPair(c, "istioingressgatewayinternalport")
			if err != nil {
				log.Error("Error getting istio ingress gateway internal port", log.Fields{
					"error":    err,
				})
				return pkgerrors.Wrapf(err,
					"Error getting istio ingress gateway internal port")
			}
			port, err = strconv.Atoi(obj)
			if err != nil {
				log.Error("Error converting port from string to uint32", log.Fields{
					"error":    err,
				})
				return pkgerrors.Wrapf(err,
					"Error converting port from string to uint32")
			}
			servers[index].ClusterData[ci].GwInternalPort = uint32(port)
			servers[index].ClusterData[ci].ClusterName = c
			servers[index].ClusterData[ci].Reslist = make([]map[string][]byte, 0)
		}
		ics, err := module.NewClientsInboundIntentClient().GetClientsInboundIntents(project,
			compositeapp,
			compositeappversion,
			deployIntentGroup,
			intentName,
			is.Metadata.Name)
		if err != nil {
			log.Error("Error getting clients inbound intents", log.Fields{
				"error": err,
			})
			return pkgerrors.Wrapf(err,
				"Error getting clients inbound intents %v under server inbound intent %v for %v/%v%v/%v not found",
				is.Metadata.Name, intentName, project, compositeapp, compositeappversion, deployIntentGroup)
		}

		li := len(ics)
		servers[index].Clients = make([]client, li)
		for i, ic := range ics {
			servers[index].Clients[i].ClientName = ic.Spec.AppName
			servers[index].Clients[i].ClientServiceName = ic.Spec.ServiceName
			clusters, err = ac.GetClusterNames(ic.Spec.AppName)
			if err != nil {
				log.Error("Error retrieving clusters from App Context", log.Fields{
					"error":    err,
					"app name": ic.Spec.AppName,
				})
				return pkgerrors.Wrapf(err,
					"Error retrieving clusters from App Context for app %v", is.Spec.AppName)
			}
			lc := len(clusters)
			servers[index].Clients[i].ClusterData = make([]clusterData, lc)
			for cci, c := range clusters {
				done := false
				// check if the server and client are on the same cluster
				for _, scd := range servers[index].ClusterData {
					if scd.ClusterName  == c {
						servers[index].Clients[i].ClusterData[cci].ClusterName = c
						servers[index].Clients[i].ClusterData[cci].Reslist = make([]map[string][]byte, 0)
						done = true
						break
					}
				}

				if done {
					continue
				}
				// check if the client side resources are alreay created for this cluster
				done = false
				for j:=0; j<i; j++ {
					for _, cd := range servers[index].Clients[j].ClusterData {
						if cd.ClusterName  == c {
							servers[index].Clients[i].ClusterData[cci].ClusterName = c
							servers[index].Clients[i].ClusterData[cci].Reslist = make([]map[string][]byte, 0)
							done = true
							break
						}
					}
					if done {
						break
					}
				}
				if done {
					continue
				}

				servers[index].Clients[i].ClusterData[cci].ClusterName = c
				servers[index].Clients[i].ClusterData[cci].Reslist = make([]map[string][]byte, 0)
				err = createClientResources(is, c, servers , namespace, index, i, cci, ips)
				if err != nil {
					log.Error("Error creating client resources", log.Fields{
						"error":    err,
						"svc name": ic.Spec.ServiceName,
					})
					return err
				}
			}
		}
		// check if the server and clients are on the same cluster
		for ci, scd := range servers[index].ClusterData {
			create := false
			for _, cli := range servers[index].Clients {
				for _, ccd := range cli.ClusterData {
					if scd.ClusterName != ccd.ClusterName {
						create = true
						break
					}
				}
				if create {
					break
				}
			}
			if !create {
				continue
			}
			err = createServerResources(is, scd.ClusterName, servers, namespace, index, ci, ips)
			if err != nil {
				log.Error("Error creating server resources", log.Fields{
					"error":    err,
					"svc name": is.Spec.ServiceName,
				})
				return pkgerrors.Wrapf(err,
					"Error creating server resources")
			}
		}
		index = index + 1

	}
	for _, s := range servers {
		// Add server resources
		for _, cd := range s.ClusterData {
			if len(cd.Reslist) <= 0 {
				continue
			}
			for _, r := range cd.Reslist {
				err = addClusterResource(ac, s.AppName,cd.ClusterName, r)
				if err != nil {
					log.Error("Error adding cluster Resource", log.Fields{
						"error":    err,
						"app name": s.AppName,
					})
					return pkgerrors.Wrapf(err, "Error adding cluster resource for %v", s.AppName)
				}
			}
		}
		for ci, cc := range s.Clients {
			//Add client resources
			for _, clu := range s.Clients[ci].ClusterData {
				if len(clu.Reslist) <= 0 {
					continue
				}
				for _, r := range clu.Reslist {
					err = addClusterResource(ac, cc.ClientName, clu.ClusterName, r)
					if err != nil {
						log.Error("Error adding cluster Resource", log.Fields{
							"error":    err,
							"app name": cc.ClientName,
						})
						return pkgerrors.Wrapf(err, "Error adding cluster resource for %v", s.AppName)
					}
				}
			}
		}
	}

	return nil
}

//func addClusterResource(ac appcontext.AppContext, is module.InboundServerIntent, c string)(error) {
func addClusterResource(ac appcontext.AppContext, appname string, c string, res map[string][]byte)(error) {
	ch, err := ac.GetClusterHandle(appname, c)
	if err != nil {
		log.Error("Error getting clusters handle App Context", log.Fields{
			"error":	err,
			"app name":	appname,
			"cluster name": c,
		})
		return pkgerrors.Wrapf(err,
			"Error getting clusters from App Context for app %v and cluster %v", appname, c)
	}
	// Add resource to the cluster


	if len(res) != 1 {
		log.Error("Error validating  resource value", log.Fields{
			"error":	err,
			"app name":	appname,
			"cluster name": c,
		})
		return pkgerrors.Wrapf(err,"Error validating resource value")
	}
	var resname string
	var r []byte
	for rname, ro := range res {
		resname = rname
		r = ro
	}

	_, err = ac.AddResource(ch, resname, string(r))
	if err != nil {
		log.Error("Error adding Resource to AppContext", log.Fields{
			"error":	err,
			"app name":	appname,
			"cluster name": c,
		})
		return pkgerrors.Wrap(err, "Error adding Resource to AppContext")
	}
	resorder, err := ac.GetResourceInstruction(appname, c, "order")
	if err != nil {
		log.Error("Error getting Resource order", log.Fields{
			"error":	err,
			"app name":	appname,
			"cluster name": c,
		})
		return pkgerrors.Wrap(err, "Error getting Resource order")
	}
	aov := make(map[string][]string)
	json.Unmarshal([]byte(resorder.(string)), &aov)
	aov["resorder"] = append(aov["resorder"], resname)
	jresord, _ := json.Marshal(aov)

	_, err = ac.AddInstruction(ch, "resource", "order", string(jresord))
	if err != nil {
		log.Error("Error updating Resource order", log.Fields{
			"error":	err,
			"app name":	appname,
			"cluster name": c,
		})
		return pkgerrors.Wrap(err, "Error updating Resource order")
	}
	return nil
}
func getProviderAndCluster(c string) (string, string, error) {
	s := strings.Split(c, "+")
	if len(s) != 2 {
		return "", "", pkgerrors.New("Not a valid cluster name")
	}

	return s[0], s[1], nil
}
func createServerResources(is module.InboundServerIntent, c string, servers []serverData, namespace string,index, ci int, ips newIP)(error) {
	pro, clu, err := getProviderAndCluster(c)
	if err != nil {
		log.Error("Not a valid cluster name", log.Fields{
                                        "cluster name": c,
		})
		return pkgerrors.Wrap(err, "Invalid cluster name")
	}
	host := is.Spec.ServiceName + "." + namespace + "." + pro + "." + clu
	hosts := []string{host}
	gwinp := servers[index].ClusterData[ci].GwInternalPort
	res, err := createGateway(is.Spec.ServiceName, hosts, gwinp)
	if err != nil {
		log.Error("Error creating Gateway", log.Fields{
			"error":        err,
			"app name":     is.Spec.ServiceName,
			"cluster name": c,
		})
		return pkgerrors.Wrap(err, "Error creating Gateway")
	}
	servers[index].ClusterData[ci].Reslist = append(servers[index].ClusterData[ci].Reslist, res)

	res, err = createServerServiceEntry(is, hosts, namespace)
	if err != nil {
		log.Error("Error creating Service Entry", log.Fields{
			"error":        err,
			"svc name":     is.Spec.ServiceName,
			"cluster name": c,
		})
		return pkgerrors.Wrap(err, "Error creating Service Entry")
	}
	servers[index].ClusterData[ci].Reslist = append(servers[index].ClusterData[ci].Reslist, res)

	svcname := is.Spec.ServiceName+"-se-server-dr"
	res, err = createDestinationRule(svcname, host, "istio-system")
	if err != nil {
		log.Error("Error creating Destination Rule", log.Fields{
			"error":        err,
			"svc name":     svcname,
			"cluster name": c,
		})
		return pkgerrors.Wrap(err, "Error creating Destination Rule")
	}
	servers[index].ClusterData[ci].Reslist = append(servers[index].ClusterData[ci].Reslist, res)
	return nil
}

func createGateway(svcname string, hosts []string, gwport uint32)(map[string][]byte, error){
	// Create gateway resource
	smap := make(map[string]string)
	smap["istio"] = "ingressgateway"
	port := Port{Name: "tls", Number: gwport, Protocol: "TLS",}
	var sts  = ServerTLSSettings{
		Mode:"AUTO_PASSTHROUGH",
	}
	csr := createServerItem(port, "", hosts, sts, "tls")
	var svs = []Server{csr}
	gspec := createGatewaySpec(svs, smap)
	resname := svcname + "-gateway"
	meta := createGenericMetadata(resname, "istio-system", "")
	out, err := createGatewayResource(meta, gspec)
	if err != nil {
		log.Error("Error creating Gateway Resource", log.Fields{
			"error":        err,
			"svc name":     svcname,
		})
		return nil, pkgerrors.Wrap(err, "Error creating Gateway Resource")
	}

	res := make(map[string][]byte)
	res[resname] = out
	return res, nil

}

func createServerServiceEntry(is module.InboundServerIntent, hosts []string, namespace string)(map[string][]byte, error){
	addresses := []string{}
	wle := []WorkloadEntry{}
	addr := is.Spec.ServiceName + "." + namespace + "." + "svc.cluster.local"
	wle = []WorkloadEntry{{Address: addr,},}

	ports := []Port{{Name: "tcp", Number: uint32(is.Spec.Port), Protocol: is.Spec.Protocol,},}
	resname := is.Spec.ServiceName + "-se-server"
	meta := createGenericMetadata(resname, "istio-system", "")
	vsspec := createServiceEntrySpec(hosts, addresses, []string{".",}, wle, ports, "MESH_INTERNAL", "DNS")
	out, err := createServieEntryResource(meta, vsspec)
	if err != nil {
		log.Error("Error creating Servie Entry Resource", log.Fields{
			"error":        err,
			"svc name":     is.Spec.ServiceName,
		})
		return nil, pkgerrors.Wrap(err, "Error creating Servie Entry Resource")
	}

	res := make(map[string][]byte)
	res[resname] = out

	return res, nil
}

func createDestinationRule(svcname, host, namespace string)(map[string][]byte, error){
	// Create dr resource
	var cts  = ClientTLSSettings{
		Mode:"ISTIO_MUTUAL",
	}
	var tp = TrafficPolicy {
		Tls: cts,
	}
	drspec, err := createDestinationRuleSpec(host, tp)
	meta := createGenericMetadata(svcname, namespace, "")
	out, err := createDestinationRuleResource(meta, drspec)
	if err != nil {
		log.Error("Error creating Destination Rule Resource", log.Fields{
			"error":        err,
			"svc name":     svcname,
		})
		return nil, pkgerrors.Wrap(err, "Error creating Destination Rule Resource")
	}

	res := make(map[string][]byte)
	resname := svcname + "-dr"
	res[resname] = out
	return res, nil
}
func createClientServiceEntry(is module.InboundServerIntent, hosts[]string, gwaddr string, gwextport uint32, namespace string, ips newIP, rescount string)(map[string][]byte, error){
	//Create se resource
	ip, err := ips.getIpAddress()
	addresses := []string{ip.String(),}
	pmap := make(map[string]uint32)
	pmap["tcp"] = gwextport
	wle := []WorkloadEntry{{Address: gwaddr,Ports: pmap,},}
	ports := []Port{{Name: "tcp", Number: uint32(is.Spec.Port), Protocol: is.Spec.Protocol,},}
	resname := is.Spec.ServiceName + "-se-client" + rescount
	meta := createGenericMetadata(resname, namespace, "")
	vsspec := createServiceEntrySpec(hosts, addresses, []string{}, wle, ports, "MESH_INTERNAL", "DNS")
	out, err := createServieEntryResource(meta, vsspec)
	if err != nil {
		log.Error("Error creating Servie Entry Resource", log.Fields{
			"error":        err,
			"svc name":     is.Spec.ServiceName,
		})
		return nil, pkgerrors.Wrap(err, "Error creating Servie Entry Resource")
	}

	res := make(map[string][]byte)
	res[resname] = out

	return res, nil
}
func createClientResources(is module.InboundServerIntent, c string, servers []serverData, namespace string,index, ci, cci int, ips newIP)(error) {
	le := len(servers[index].ClusterData)
	hosts := make([]string, le)
	for i, sc := range servers[index].ClusterData {
		pro, clu, err := getProviderAndCluster(sc.ClusterName)
		if err != nil {
			log.Error("Not a valid cluster name", log.Fields{
					"cluster name": sc.ClusterName,
			})
			return pkgerrors.Wrap(err, "Invalid cluster name")
		}
		host := is.Spec.ServiceName + "." + namespace + "." +  pro + "." + clu
		hosts[i] = host
	}
	gwaddr := servers[index].ClusterData[ci].GwAddress
	gwextport := servers[index].ClusterData[ci].GwExternalPort
	res, err := createClientServiceEntry(is, hosts, gwaddr, gwextport, namespace, ips, "0")
	if err != nil {
		log.Error("Error creating client Servie Entry", log.Fields{
			"error":        err,
			"svc name":     is.Spec.ServiceName,
		})
		return pkgerrors.Wrap(err, "Error creating client Servie Entry")
	}
	servers[index].Clients[ci].ClusterData[cci].Reslist = append(servers[index].Clients[ci].ClusterData[cci].Reslist, res)

	hs := make([]string, 1)
	hs[0] = is.Spec.ServiceName
	res, err = createClientServiceEntry(is, hs, gwaddr, gwextport, namespace, ips, "1")
	if err != nil {
		log.Error("Error creating client named Servie Entry", log.Fields{
			"error":        err,
			"svc name":     is.Spec.ServiceName,
		})
		return pkgerrors.Wrap(err, "Error creating client Servie Entry")
	}
	servers[index].Clients[ci].ClusterData[cci].Reslist = append(servers[index].Clients[ci].ClusterData[cci].Reslist, res)

	for i, h := range hosts {
		svcname := is.Spec.ServiceName + "-dr-client" + strconv.Itoa(i)
		res, err = createDestinationRule(svcname, h, namespace)
		if err != nil {
			log.Error("Error creating Destination Rule", log.Fields{
				"error":        err,
				"svc name":     svcname,
				"host name":    h,
			})
			return pkgerrors.Wrap(err, "Error creating Destination Rule")
		}
		servers[index].Clients[ci].ClusterData[cci].Reslist = append(servers[index].Clients[ci].ClusterData[cci].Reslist, res)
	}
	svcname := is.Spec.ServiceName + "-dr-client-svcname"
	res, err = createDestinationRule(svcname, hs[0], namespace)
	servers[index].Clients[ci].ClusterData[cci].Reslist = append(servers[index].Clients[ci].ClusterData[cci].Reslist, res)
	return nil

}

func getClusterKvPair(c, kvkey string)(string, error) {

	parts := strings.Split(c, "+")
	if len(parts) != 2 {
		log.Error("Not a valid cluster name", log.Fields{
                                        "cluster name": c,
		})
		return "", pkgerrors.New("Not a valid cluster name")
	}
	ckv, err := clusterPkg.NewClusterClient().GetAllClusterKvPairs(parts[0], parts[1])
	var val string
	if err == nil {
		for _, kvp := range ckv {
			for _, mkey := range kvp.Spec.Kv {
				if v, ok := mkey[kvkey]; ok {
					val = fmt.Sprintf("%v", v)
					return val, nil
				}
			}
		}
	}

	return "", pkgerrors.New("Cluster kvpair not found")

}
