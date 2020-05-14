package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/astralkn/k8s-logs-extractor/pkg/extractor"
	"github.com/astralkn/k8s-logs-extractor/pkg/kube"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var (
	version = "master"
)

func main() {
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

	if err := run(opts); err != nil {
		log.Error(err.Error())
	} else {
		log.Println("Logs extracted to ", opts.outputFile)
	}
}

func getConfigs(kcPath string) ([]string, error) {
	var configs []string
	dir, err := ioutil.ReadDir(kcPath)
	if err != nil {
		return configs, err
	}
	m := regexp.MustCompile(`.*\.kubeconfig`)
	for i := range dir {
		conf := kcPath + "/" + dir[i].Name()
		fi, err := os.Stat(conf)
		if err != nil {
			return []string{}, err
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			ec, err := getConfigs(conf)
			if err != nil {
				return []string{}, err
			}
			configs = append(configs, ec...)
		case mode.IsRegular():
			if m.MatchString(fi.Name()) {
				configs = append(configs, conf)
				fmt.Println(conf)
			}
		}
	}

	if len(configs) == 0 {
		return configs, errors.New("no kubeconfig found")
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
Default options: all extractions enabled
Flags:
`, name)
		flags.PrintDefaults()
		log.Errorf(`
`)
	}

	flags.StringVar(&opts.kubeConfigPath, "kc", os.Getenv("HOME")+"/.kube/", "set cluster kubeconfig path")
	flags.StringVar(&opts.outputFile, "o", "/cluster-logs/", "set logs output file")
	flags.BoolVar(&opts.version, "version", false, "show version and exit")
	flags.BoolVar(&opts.pod, "no-pod", true, "do not extract pod logs option")
	flags.BoolVar(&opts.cm, "no-cm", true, "do not extract config maps option")
	flags.BoolVar(&opts.svc, "no-svc", true, "do not extract services option")
	flags.BoolVar(&opts.crd, "no-crd", true, "do not extract crds option")

	return flags, &opts
}

type options struct {
	kubeConfigPath string
	outputFile     string
	version        bool
	pod            bool
	cm             bool
	svc            bool
	crd            bool
}

func run(opts *options) error {
	configs, err := getConfigs(opts.kubeConfigPath)
	if err != nil {
		return err
	}
	var errs errs
	var wg sync.WaitGroup
	for _, s := range configs {
		acc, err := kube.NewAccessor(s, "")
		if err != nil {
			return err
		}
		str := strings.Split(s, "/")
		cluster := str[len(str)-1]
		if opts.pod {
			wg.Add(1)
			go func(acc *kube.Accessor, outputFile, clusterName string) {
				err := extractor.PodExtractor{}.Extract(acc, filepath.Join(outputFile, clusterName))
				if err != nil {
					errs = append(errs, err)
				}
				defer wg.Done()
			}(acc, opts.outputFile, cluster)
		}
		if opts.cm {
			wg.Add(1)
			go func(acc *kube.Accessor, outputFile, clusterName string) {
				err := extractor.CMExtractor{}.Extract(acc, filepath.Join(outputFile, clusterName))
				if err != nil {
					errs = append(errs, err)
				}
				defer wg.Done()

			}(acc, opts.outputFile, cluster)
		}
		if opts.svc {
			wg.Add(1)
			go func(acc *kube.Accessor, outputFile, clusterName string) {
				err := extractor.SVCExtractor{}.Extract(acc, filepath.Join(outputFile, clusterName))
				if err != nil {
					errs = append(errs, err)
				}
				defer wg.Done()
			}(acc, opts.outputFile, cluster)
		}
		if opts.crd {
			wg.Add(1)
			go func(acc *kube.Accessor, outputFile, clusterName string) {
				err := extractor.CRDExtractor{}.Extract(acc, filepath.Join(outputFile, clusterName))
				if err != nil {
					errs = append(errs, err)
				}
				defer wg.Done()
			}(acc, opts.outputFile, cluster)
		}
	}
	wg.Wait()
	if len(errs) > 0 {
		return errs
	}
	return nil
}

type errs []error

func (es errs) Error() string {
	buff := bytes.NewBufferString("multiple errors: \n")
	for _, e := range es {
		_, _ = fmt.Fprintf(buff, "\t%s\n", e)
	}
	return buff.String()
}
