package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Version of the program
const (
	Version = "0.1.0"
)

var (
	// Authors is the authors
	Authors = []cli.Author{{Name: "Erik Hollensbe", Email: "erik+github@hollensbe.org"}}
)

func main() {
	app := cli.NewApp()
	app.Version = Version
	app.Authors = Authors

	app.Flags = []cli.Flag{}

	app.Action = run

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(ctx *cli.Context) error {
	var (
		err   error
		k8s   *rest.Config
		kcEnv = os.Getenv("KUBECONFIG")
	)

	if kcEnv != "" {
		k8s, err = clientcmd.BuildConfigFromFlags("", kcEnv)
		if err != nil {
			return err
		}
	} else {
		k8s, err = rest.InClusterConfig()
		if err != nil {
			return errors.New("please run this with a kubernetes configuration, or within kubernetes")
		}
	}

	cs, err := kubernetes.NewForConfig(k8s)
	if err != nil {
		return err
	}

	namespaces, err := cs.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, item := range namespaces.Items {
		pods, err := cs.CoreV1().Pods(item.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		for _, pod := range pods.Items {
			go func(pod corev1.Pod) {
				res := cs.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Follow: true})
				reader, err := res.Stream(context.Background())
				if err != nil {
					fmt.Fprintln(os.Stderr, color.RedString(err.Error()))
					return
				}

				r := bufio.NewScanner(reader)
				for r.Scan() {
					fmt.Fprintln(os.Stdout, color.GreenString(pod.Namespace), color.YellowString(pod.Name), r.Text())
				}
			}(pod)
		}
	}

	select {}
}
