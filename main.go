package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/astralkn/k8s-logs-extractor/pkg/kube"
	"github.com/sergi/go-diff/diffmatchpatch"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"io/ioutil"
	kubeApiCore "k8s.io/api/core/v1"
	"os"
	"regexp"
	"strings"
	"time"
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
	err := run(opts)
	if err != nil {
		log.Fatal(err)
	}
	if opts.diff {
		log.Println("Diffs extracted to ", opts.outputFile)
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

Flags:
`, name)
		flags.PrintDefaults()
		log.Errorf(`
`)
	}

	flags.StringVar(&opts.kubeConfigPath, "kc", os.Getenv("HOME")+"/.kube/", "set cluster kubeconfig path")
	flags.StringVar(&opts.outputFile, "o", "/cluster-logs/", "set logs output file")
	flags.BoolVar(&opts.version, "version", false, "show version and exit")
	flags.BoolVar(&opts.diff, "diff", false, "create diff file")
	return flags, &opts
}

type options struct {
	kubeConfigPath string
	outputFile     string
	version        bool
	diff           bool
}

func run(opts *options) error {
	configs, err := getConfigs(opts.kubeConfigPath)
	if err != nil {
		return err
	}
	for i, s := range configs {
		err := extractLogs(fmt.Sprintf("%s/cluster-%d", opts.outputFile, i), s, opts.diff)
		if err != nil {
			return err
		}
	}
	return err
}

func extractLogs(path, kc string, diff bool) error {
	acc, err := kube.NewAccessor(kc, "")
	if err != nil {
		return err
	}

	allPods, err := acc.GetAllPods()
	if err != nil {
		return err
	}

	for _, p := range allPods {
		err := extractLogsForPod(path, acc, p, diff)
		if err != nil {
			return err
		}
	}
	return nil
}

func extractLogsForPod(path string, acc *kube.Accessor, p kubeApiCore.Pod, diff bool) error {
	logs, err := acc.Logs(p.Namespace, p.Name, "", false)
	if err != nil {
		var cont = getPodContainers(err)
		for _, p1 := range cont {
			logs, err := acc.Logs(p.Namespace, p.Name, p1, false)
			if err != nil {
				return err
			}
			err = writeLogsToFile(path+"/"+p.Namespace, p.Name+"_"+p1, logs, diff)
			return err
		}
		return nil
	} else {
		err := writeLogsToFile(path+"/"+p.Namespace, p.Name, logs, diff)
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

func writeLogsToFile(path, file string, logs string, diffr bool) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}

	filepath := path + "/" + file
	_, err := os.Stat(filepath + ".txt")
	if os.IsNotExist(err) {
		err = createFile(filepath+".txt", logs)
		if err != nil {
			return err
		}
	} else if diffr {

		content, err := ioutil.ReadFile(filepath + ".txt")
		if err != nil {
			return err
		}
		s := string(content)
		dmp := diffmatchpatch.New()

		diffs := dmp.DiffMain(s, logs, true)

		fs := func(s, sign string) string {
			split := strings.Split(s, "\n")
			for i := range split {
				split[i] = sign + split[i]
			}
			return strings.Join(split, "\n")
		}

		f := func(diffs []diffmatchpatch.Diff) string {
			var text bytes.Buffer

			for _, aDiff := range diffs {
				if aDiff.Type == diffmatchpatch.DiffInsert {

					_, _ = text.WriteString(fs(aDiff.Text, "+"))
				}
				if aDiff.Type == diffmatchpatch.DiffDelete {
					_, _ = text.WriteString(fs(aDiff.Text, "-"))
				}
				if aDiff.Type == diffmatchpatch.DiffEqual {
					_, _ = text.WriteString(aDiff.Text)
				}
			}
			return text.String()
		}

		text := f(diffs)

		err = createFile(filepath+"_diff_"+time.Now().Format(time.RFC3339)+".diff", text)
		if err != nil {
			return err
		}
	} else {
		err = createFile(filepath+"_"+time.Now().Format(time.RFC3339)+".txt", logs)
		if err != nil {
			return err
		}
	}
	return err
}

func createFile(filepath string, logs string) error {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(logs))
	if err != nil {
		return err
	}
	err = f.Close()
	return nil
}
