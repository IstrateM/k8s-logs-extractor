package extractor

import (
	"github.com/astralkn/k8s-logs-extractor/pkg/kube"
	"os"
	"path/filepath"
	"strings"
)

const (
	//NONE = ""
	YAML = ".yaml"
	OUT  = ".out"
)

type Extractor interface {
	Extract(acc *kube.Accessor, outputDir string) error
}

type PodExtractor struct {
}

func (_ PodExtractor) Extract(acc *kube.Accessor, outputDir string) error {
	_, err := acc.DumpInfo(outputDir, "all")
	if err != nil {
		return err
	}
	podList, err := acc.GetPods("", "all")
	if err != nil {
		return err
	}
	err = writeStringToFile(outputDir, "pods", podList, OUT)
	if err != nil {
		return err
	}
	podDescribe, err := acc.DescribePod("", "all")
	if err != nil || podDescribe == "No resources found" || podDescribe == "" {
		return err
	}
	split := strings.Split(podDescribe, "\nName:")
	for i, pod := range split {
		name := getName(pod)
		if i != 0 {
			pod = "Name:" + pod
		}
		//TODO: add namespace as well
		err = writeStringToFile(filepath.Join(outputDir, "pods-describe"), name, pod, YAML)
		if err != nil {
			return err
		}
	}
	return err
}

type CMExtractor struct {
}

func (_ CMExtractor) Extract(acc *kube.Accessor, outputDir string) error {
	s, err := acc.DescribeCM("", "all")
	if err != nil || s == "No resources found" || s == "" {
		return err
	}
	split := strings.Split(s, "\nName:")
	for i, cm := range split {
		name := getName(cm)
		if i != 0 {
			cm = "Name:" + cm
		}
		err = writeStringToFile(filepath.Join(outputDir, "cm"), name, cm, YAML)
		if err != nil {
			return err
		}
	}
	return err
}

type SVCExtractor struct {
}

func (_ SVCExtractor) Extract(acc *kube.Accessor, outputDir string) error {
	s, err := acc.DescribeSVC("", "all")
	if err != nil || s == "No resources found" || s == "" {
		return err
	}
	split := strings.Split(s, "\nName:")
	for i, svc := range split {
		name := getName(svc)
		if i != 0 {
			svc = "Name:" + svc
		}
		err = writeStringToFile(filepath.Join(outputDir, "svc"), name, svc, YAML)
		if err != nil {
			return err
		}
	}
	return err
}

type CRDExtractor struct {
}

func (_ CRDExtractor) Extract(acc *kube.Accessor, outputDir string) error {
	s, err := acc.DescribeCRD("", "all")
	if err != nil || s == "No resources found" || s == "" {
		return err
	}
	split := strings.Split(s, "\nName:")
	for i, crd := range split {
		name := getName(crd)
		if i != 0 {
			crd = "Name:" + crd
		}
		dir := filepath.Join(outputDir, "crd", name)
		err = writeStringToFile(dir, name, crd, YAML)
		if err != nil {
			return err
		}
		err = CRExtractor{}.Extract(acc, dir, name)
		if err != nil {
			return err
		}
	}
	return err
}

type CRExtractor struct {
}

func (_ CRExtractor) Extract(acc *kube.Accessor, outputDir, crd string) error {
	s, err := acc.DescribeCR("", crd, "all")
	if err != nil || s == "No resources found" || s == "" {
		return err
	}
	split := strings.Split(s, "\nName:")
	for i, crd := range split {
		name := getName(crd)
		if i != 0 {
			crd = "Name:" + crd
		}
		err = writeStringToFile(filepath.Join(outputDir, "instances"), name, crd, YAML)
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

func writeStringToFile(path, file, str, fileType string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}

	fpath := filepath.Join(path, file)
	_, err := os.Stat(fpath + fileType)
	if os.IsNotExist(err) {
		err = createFile(fpath+fileType, str)
		if err != nil {
			return err
		}
	}
	return nil
}

func getName(manifest string) string {
	line1 := strings.Split(manifest, "\n")[0]
	n := strings.TrimPrefix(line1, "Name:")
	return strings.TrimSpace(n)
}
