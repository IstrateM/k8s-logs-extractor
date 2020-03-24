package main

import (
	"fmt"
	"github.com/astralkn/k8s-logs-extractor/pkg/kube"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"io/ioutil"
	kubeApiCore "k8s.io/api/core/v1"
	"os"
	"regexp"
	"strings"
)

var (
	version = "master"
	f       *os.File
)

func main() {
	defer f.Close()
	name := os.Args[0]
	flags, opts := setupFlags(name)
	switch err := flags.Parse(os.Args[1:]); {
	case err == pflag.ErrHelp:
		os.Exit(0)
	case err != nil:
		log.Errorf("%s", err.Error())
		flags.Usage()
		os.Exit(1)
	}

	if opts.version {
		log.Printf("log-extractor version %s\n", version)
		os.Exit(0)
	}
	err := run(opts)
	if err != nil {
		log.Fatal(err)
	}
}

func getConfigs(opts *options) ([]string, error) {
	var configs []string
	dir, err := ioutil.ReadDir(opts.kubeConfigPath)
	if err != nil {
		return configs, err
	}
	for i := range dir {
		conf := opts.kubeConfigPath + dir[i].Name()
		configs = append(configs, conf)
		fmt.Println(conf)
	}
	return configs, err
}

func setupFlags(name string) (*pflag.FlagSet, *options) {
	opts := options{}
	flags := pflag.NewFlagSet(name, pflag.ContinueOnError)
	flags.SetInterspersed(false)
	flags.Usage = func() {
		log.Errorf(`Usage:
    %s [flags]

Flags:
`, name)
		flags.PrintDefaults()
		log.Errorf(`
`)
	}

	flags.StringVar(&opts.kubeConfigPath, "kc", os.Getenv("HOME")+"/.kube/", "set cluster kubeconfig path")
	flags.StringVar(&opts.outputFile, "o", "/cluster-logs/", "set logs output file")
	flags.BoolVar(&opts.version, "version", false, "show version and exit")
	return flags, &opts
}

type options struct {
	kubeConfigPath string
	outputFile     string
	version        bool
}

func run(opts *options) error {
	configs, err := getConfigs(opts)
	if err != nil {
		return err
	}
	for i, s := range configs {
		err := extractLogs(fmt.Sprintf("%s/cluster-%d", opts.outputFile, i), s)
		if err != nil {
			return err
		}
	}
	return err
}

func extractLogs(path, kc string) error {
	acc, err := kube.NewAccessor(kc, "")
	if err != nil {
		return err
	}

	allPods, err := acc.GetAllPods()
	if err != nil {
		return err
	}

	for _, p := range allPods {
		err := extractLogsForPod(path, acc, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func extractLogsForPod(path string, acc *kube.Accessor, p kubeApiCore.Pod) error {
	logs, err := acc.Logs(p.Namespace, p.Name, "", false)
	if err != nil {
		var cont = getPodContainers(err)
		for _, p1 := range cont {
			logs, err := acc.Logs(p.Namespace, p.Name, p1, false)
			if err != nil {
				return err
			}
			err = writeLogsToFile(path+"/"+p.Namespace, p.Name, logs)
			return err
		}
		return nil
	} else {
		err := writeLogsToFile(path+"/"+p.Namespace, p.Name, logs)
		return err
	}

}

func getPodContainers(err error) []string {
	var cont []string
	m := regexp.MustCompile(`\[[a-zA-Z0-9 \-]*\]`)
	allString := m.FindAllString(err.Error(), 1)
	for _, s := range allString {
		s = strings.TrimPrefix(s, "[")
		s = strings.TrimSuffix(s, "]")
		cont = append(cont, strings.Split(s, " ")...)
	}
	return cont
}

func writeLogsToFile(path, file string, logs string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}

	f, err := os.OpenFile(path+"/"+file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(logs))
	if err != nil {
		return err
	}
	err = f.Close()
	return err
}
