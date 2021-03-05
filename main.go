package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
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
	Authors = []*cli.Author{{Name: "Erik Hollensbe", Email: "erik+github@hollensbe.org"}}
)

func main() {
	app := cli.NewApp()
	app.Version = Version
	app.Authors = Authors

	app.Flags = []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "namespace",
			Aliases: []string{"n"},
			Usage:   "Namespaces you care about; may provide multiple",
		},
		&cli.Int64Flag{
			Name:    "since",
			Aliases: []string{"s"},
			Usage:   "Will only show logs after this time in seconds",
		},
		&cli.TimestampFlag{
			Name:    "after",
			Aliases: []string{"a"},
			Usage:   "Will only show logs that appear after this time",
			Layout:  time.RFC3339,
		},
		&cli.BoolFlag{
			Name:    "timestamp",
			Aliases: []string{"t"},
			Usage:   "Timestamp log messages",
			Value:   true,
		},
	}

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

	if ctx.IsSet("after") && ctx.IsSet("since") {
		return errors.New("after and since may not be set at the same time")
	}

	var (
		after *metav1.Time
		since *int64
	)

	if ctx.Timestamp("after") != nil {
		t := metav1.NewTime(*ctx.Timestamp("after"))
		after = &t
	} else if ctx.Int64("since") != 0 {
		val := ctx.Int64("since")
		since = &val
	}

	if after == nil && since == nil {
		t := metav1.Now()
		after = &t
	}

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

	namespaces := []string{}

	if len(ctx.StringSlice("namespace")) != 0 {
		namespaces = ctx.StringSlice("namespace")
	} else {
		list, err := cs.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		for _, item := range list.Items {
			namespaces = append(namespaces, item.Name)
		}
	}

	for _, ns := range namespaces {
		pods, err := cs.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		for _, pod := range pods.Items {
			go func(pod corev1.Pod) {
				res := cs.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
					Follow:       true,
					SinceSeconds: since,
					SinceTime:    after,
					Timestamps:   ctx.Bool("timestamp"),
				})
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
