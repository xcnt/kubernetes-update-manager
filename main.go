package main

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	_, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

}
