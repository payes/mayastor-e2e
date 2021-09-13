package lib

/*

func GenerateYamlFiles() {
        e2eCfg := e2e_config.GetConfig()

        coresDirective := ""
        if e2eCfg.Cores != 0 {
                coresDirective = fmt.Sprintf("%s -c %d", coresDirective, e2eCfg.Cores)
        }

        nodeLocs, err := k8stest.GetNodeLocs()
        Expect(err).ToNot(HaveOccurred(), "GetNodeLocs failed %v", err)
        poolDirectives := ""
        masterNode := ""
        if len(e2eCfg.PoolDevice) != 0 {
                poolDevice := e2eCfg.PoolDevice
                for _, node := range nodeLocs {
                        if node.MasterNode {
                                masterNode = node.NodeName
                        }
                        if !node.MayastorNode {
                                continue
                        }
                        poolDirectives += fmt.Sprintf(" -p '%s,%s'", node.NodeName, poolDevice)
                }
        }

        registryDirective := ""
        if len(e2eCfg.Registry) != 0 {
                registryDirective = fmt.Sprintf(" -r '%s'", e2eCfg.Registry)
        }

        imageTag := e2eCfg.ImageTag

        etcdOptions := "etcd.replicaCount=1,etcd.nodeSelector=kubernetes.io/hostname: " + masterNode + ",etcd.tolerations=- key: node-role.kubernetes.io/master"
        bashCmd := fmt.Sprintf(
                "%s/generate-deploy-yamls.sh -s '%s' -o %s -t '%s' %s %s %s test",
                locations.GetMayastorScriptsDir(),
                etcdOptions,
                locations.GetGeneratedYamlsDir(),
                imageTag, registryDirective, coresDirective, poolDirectives,
        )
        logf.Log.Info("About to execute", "command", bashCmd)
        cmd := exec.Command("bash", "-c", bashCmd)
        out, err := cmd.CombinedOutput()
        Expect(err).ToNot(HaveOccurred(), "%s", out)
}


*/
