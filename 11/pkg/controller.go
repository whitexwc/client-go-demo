package pkg

import (
	"context"
	"reflect"
	"time"

	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	informer "k8s.io/client-go/informers/core/v1"
	netInformer "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	lister "k8s.io/client-go/listers/core/v1"
	netLister "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	workerNum = 5
	maxRetry = 5 
)

type controller struct {
	client        kubernetes.Interface
	ingressLister netLister.IngressLister
	serviceLister lister.ServiceLister
	queue         workqueue.RateLimitingInterface
}

func (c *controller) Run(stopCh chan struct{}) {
	for i := 0; i < workerNum; i++ {
		go wait.Until(c.worker, time.Minute, stopCh)
	}
	<-stopCh
}

func (c *controller) worker() {
	for c.processNextItem() {

	}
}

func (c *controller) processNextItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(item)

	key := item.(string)

	err := c.syncService(key)
	if err != nil {
		c.handlerError(key, err)
	}
}

func (c *controller) syncService(key string) error {
	namespaceKey, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	// 删除的情况
	service, err := c.serviceLister.Services(namespaceKey).Get(name)
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	// 新增和删除
	_, ok := service.GetAnnotations()["ingress/http"]
	ingress, err := c.ingressLister.Ingresses(namespaceKey).Get(name)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if ok && errors.IsNotFound(err) {
		// create ingress
		ig := c.constructIngress(namespaceKey, name)
		_, err = c.client.NetworkingV1().Ingresses(namespaceKey).Create(context.TODO(), ig, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else if !ok && ingress != nil {
		// delete ingress
		err = c.client.NetworkingV1().Ingresses(namespaceKey).Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

}

func (c *controller) constructIngress(namespaceKey string, name string) *v1.Ingress {
	ingress := v1.Ingress{}
	ingress.Name = name
	ingress.Namespace = namespaceKey
	pathType := v1.PathTypePrefix
	ingress.Spec = v1.IngressSpec{
		Rules: []v1.IngressRule{
			{
				Host: "example.com",
				IngressRuleValue: v1.IngressRuleValue{
					HTTP: &v1.HTTPIngressRuleValue{
						Paths: []v1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &pathType,
								Backend: v1.IngressBackend{
									Service: &v1.IngressServiceBackend{
										Name: name,
										Port: v1.ServiceBackendPort{
											Number: 80,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return &ingress

}

func (c *controller) handlerError(key string, err error) {
	if c.queue.NumRequeues(key) > maxRe
	c.queue.AddRateLimited(key)
}

func NewController(client kubernetes.Interface, serviceInformer informer.ServiceInformer, ingressInformer netInformer.IngressInformer) controller {
	c := controller{
		client:        client,
		ingressLister: ingressInformer.Lister(),
		serviceLister: serviceInformer.Lister(),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingressManager"),
	}

	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addService,
		UpdateFunc: c.updateService,
		//DeleteFunc: c.deleteService,
	})

	ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.deleteIngress,
	})

	return c
}

func (c *controller) addService(obj interface{}) {
	c.enqueue(obj)
}

func (c *controller) updateService(oldObj, newObj interface{}) {
	// todo 比较annotation
	if reflect.DeepEqual(oldObj, newObj) {
		return
	}
	c.enqueue(newObj)
}

// func (c *controller) deleteService(obj interface{}) {

// }

func (c *controller) enqueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
	}
	c.queue.Add(key)
}

// func (c *controller) addIngress(obj interface{}) {

// }

// func (c *controller) updateIngress(oldObj, newObj interface{}) {

// }

func (c *controller) deleteIngress(obj interface{}) {
	ingress := obj.(*v1.Ingress)
	service := metav1.GetControllerOfNoCopy(ingress)
	if service == nil {
		return
	}
	if service.Kind != "Service" {
		return
	}

	c.queue.Add(ingress.Namespace + "/" + ingress.Name)
}
