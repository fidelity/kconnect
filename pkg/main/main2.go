package main

import (
	"fmt"
	"github.com/imdario/mergo"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func main() {

	u1 := unstructured.Unstructured{
		Object: map[string]interface{}{"sdfds7fds8fsd8fsd": nil},
	}

	u2 := unstructured.Unstructured{
		Object: map[string]interface{}{"877878vxcvx6v6x": nil},
	}

	cfg1 := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"eks-cluster-2": &clientcmdapi.Cluster{
				Server:                   "https://kubeapiserver1.com",
				Extensions: map[string]runtime.Object{"kconnect": &u1},
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"eks-cluster-1-iam-role-1": {
				Cluster:  "eks-cluster-1",
				AuthInfo: "iam-role-1",
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"iam-role-1": &clientcmdapi.AuthInfo{
				Token: "tokenn string gggggggggg",
			},
		},
		CurrentContext: "eks-cluster-1-iam-role-1",
	}

	cfg2 := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"eks-cluster-2": &clientcmdapi.Cluster{
				CertificateAuthorityData: []byte("Here is a string...."),
				Extensions: map[string]runtime.Object{"kconnect": &u2},
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"eks-cluster-1-iam-role-2": {
				Cluster:  "eks-cluster-2",
				AuthInfo: "iam-role-2",
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"iam-role-2": &clientcmdapi.AuthInfo{
				Token: "tokenn string gggggggggg",
			},
		},
		CurrentContext: "eks-cluster-1-iam-role-2",
	}

	if err := mergo.Merge(&cfg2.Clusters, &cfg1.Clusters); err != nil {
		fmt.Println("ERROR", err)
	}

	for _, cluster := range cfg2.Clusters {
		fmt.Println(cluster.Server)
		fmt.Println(cluster.CertificateAuthorityData)
		for k, v := range cluster.Extensions {
			fmt.Println("extension: %s",k )
			fmt.Println(len(v.(*unstructured.Unstructured).Object))
			for k,_ := range v.(*unstructured.Unstructured).Object{
				fmt.Println(k)
			}
		}
	}
	//fmt.Println(kubeconfig.NewClient().Write(cfg))
}
