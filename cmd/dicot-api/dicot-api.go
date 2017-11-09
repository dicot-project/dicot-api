/*
 * This file is part of the Dicot project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2017 Red Hat, Inc.
 *
 */

package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/spf13/pflag"
	k8s "k8s.io/client-go/kubernetes"
	k8srest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/dicot-project/dicot-api/pkg/api"
	"github.com/dicot-project/dicot-api/pkg/api/identity"
	"github.com/dicot-project/dicot-api/pkg/auth"
	"github.com/dicot-project/dicot-api/pkg/rest"
	computev2_1 "github.com/dicot-project/dicot-api/pkg/rest/compute/v2_1"
	identityv3 "github.com/dicot-project/dicot-api/pkg/rest/identity/v3"
	imagev2 "github.com/dicot-project/dicot-api/pkg/rest/image/v2"
)

func GetClientConfig(kubeconfig string) (*k8srest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return k8srest.InClusterConfig()
}

func GetDicotClient(kubeconfig string) (api.Interface, error) {
	config, err := GetClientConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	return api.NewClientset(config)
}

func GetKubernetesClient(kubeconfig string) (k8s.Interface, error) {
	config, err := GetClientConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	return k8s.NewForConfig(config)
}

func GetTokenManager(cl identity.Interface) (auth.TokenManager, error) {
	key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return auth.NewTokenManager([]interface{}{
		key,
	}, time.Hour, cl), nil
}

func main() {
	var debug bool
	var logRequests bool
	var kubeconfig string
	var imagerepo string

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kube config. Only required if out-of-cluster.")
	pflag.BoolVarP(&debug, "debug", "d", false, "Debug mode")
	pflag.BoolVarP(&logRequests, "log-requests", "l", false, "Log requests")
	pflag.StringVar(&imagerepo, "imagerepo", "/srv/images", "Path to image repository storage.")

	pflag.Parse()

	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(rest.AcceptJSON())
	if logRequests {
		router.Use(gin.Logger())
	}

	client, err := GetDicotClient(kubeconfig)
	if err != nil {
		log.Fatal("Dicot client: %s\n", err)
	}
	k8sClient, err := GetKubernetesClient(kubeconfig)
	if err != nil {
		log.Fatal("Kube client: %s\n", err)
	}

	tm, err := GetTokenManager(client.Identity())
	if err != nil {
		log.Fatal("Token manager: %s\n", err)
	}

	serverID := "e1552b45-f0cb-4d2b-bfb9-ae0877696e39"

	services := &rest.ServiceList{}
	services.AddService(identityv3.NewService(client, k8sClient, tm, services, ""))
	services.AddService(computev2_1.NewService(client, k8sClient, tm, serverID, ""))
	services.AddService(imagev2.NewService(client, tm, imagerepo, serverID, ""))
	services.RegisterRoutes(router)

	srv := &http.Server{
		Addr:    ":8089",
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	glog.V(1).Info("Running, use Ctrl-C to exit...")
	<-quit
	glog.V(1).Info("Shuting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server failed to shutdown:", err)
	}
	glog.V(1).Info("Server exiting")
}
