package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"strings"

	"fyne.io/systray"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
    systray.Run(onReady, onExit)
}

func onReady() {
	
	var selected *systray.MenuItem
	icon, err := getIcon("kubernetes-icon.png")
	if err != nil {
		fmt.Println("error in getting icon: ", err)
		return
	}
    systray.SetIcon(icon)
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfig, _ := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{
		CurrentContext: "",
	}).RawConfig()
	systray.SetTitle("Current Context: " + kubeconfig.CurrentContext)
	for k, v := range kubeconfig.Contexts {
		menuItem := systray.AddMenuItem(k, v.Cluster)
		go func() {
			for range menuItem.ClickedCh {
				if selected != nil {
					selected.Uncheck()
				}
				selected = menuItem
				menuItem.Check()
				arr := strings.Split(menuItem.String(), ",")
				clusterName := strings.Replace(arr[1][:len(arr[1])-1], "\"", "", 2)
				title := ChangeKubeContext(kubeconfig.CurrentContext, clusterName)
				if len(title) > 30 {
					systray.SetTitle("Current Context: " + title[:31])
				} else {
					systray.SetTitle("Current Context: " + title)
				}
				systray.SetTooltip(ChangeKubeContext(kubeconfig.CurrentContext, clusterName))
			}
		}()
	}
    systray.AddSeparator()
	menu := systray.AddMenuItem("Quit", "Quit the program")
	go func() {
		<-menu.ClickedCh
		systray.Quit()
	}()
}

func onExit() {
	os.Exit(0)
}

func ChangeKubeContext(currentContext, clusterName string) string {
	if strings.Contains(currentContext, strings.Trim(clusterName, " ")) {
		return clusterName
	}
	cmd := exec.Command("bash", "-c", fmt.Sprintf("kubectl config use-context %s", clusterName))
	_, err := cmd.Output()
	if err != nil {
		fmt.Printf("error in listing Iptables chains: %v\n", err)
		os.Exit(1)
	}
    return clusterName
}

func getIcon(filePath string) ([]byte, error) {
	var buffer bytes.Buffer
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("error in opening icon file: ", err)
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Println("error in decoding icon file: ", err)
		return nil, err
	}
	err = png.Encode(&buffer, img)
	if err != nil {
		fmt.Println("error in encoding icon file to bytes: ", err)
		return nil, err
	}
    return buffer.Bytes(), nil
}

