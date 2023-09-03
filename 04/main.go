package main

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	// create config
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err)
	}

	// create a specific client use the config
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	podClient := clientset.CoreV1().Pods("default")

	// use the client to get data
	pod, err := podClient.Get(context.TODO(), "master-busybox", metav1.GetOptions{})
	if err != nil {
		println(err)
	} else {
		println(pod.Name)
	}

	// RESTClient demo
	// //config
	// config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	// if err != nil {
	// 	panic(err)
	// }
	// config.GroupVersion = &v1.SchemeGroupVersion
	// config.NegotiatedSerializer = scheme.Codecs
	// config.APIPath = "/api"

	// //client
	// restClient, err := rest.RESTClientFor(config)
	// if err != nil {
	// 	panic(err)
	// }

	// // get data
	// pod := v1.Pod{}
	// err = restClient.Get().Namespace("default").Resource("pods").Name("master-busybox").Do(context.TODO()).Into(&pod)
	// if err != nil {
	// 	println(err)
	// } else {
	// 	println(pod.Name)
	// }
}
