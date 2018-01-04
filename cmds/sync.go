package cmds

import (
	"github.com/spf13/cobra"
	"log"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"fmt"
	"k8s.io/client-go/tools/clientcmd"
	clientset "k8s.io/client-go/kubernetes"
	"path/filepath"
	"k8s.io/client-go/util/homedir"
)

func NewSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Pulls dependent app manifests",
		Run: func(cmd *cobra.Command, args []string) {
			err := runSync()
			if err != nil {
				log.Fatalln(err)
			}
		},
	}
	return cmd
}

func runSync() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}

	repo, err := getRootDir(path)
	if err != nil {
		return err
	}

	branch, err := repo.Current()
	commitInfo, err := repo.CommitInfo(branch)
	if err != nil {
		return err
	}
	fmt.Println("------------------", commitInfo.Commit)

	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "server",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "kubectl",
					Image: "appscode/kubectl:1.8.0",
					VolumeMounts: []v1.VolumeMount{
						{
							MountPath: "/mypath",
							Name:      "git-volume",
						},
					},
					Command: []string{
						"/bin/sh",
						"-c",
						"while true;do echo hi ; sleep 500;  done",
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "git-volume",
					VolumeSource: v1.VolumeSource{
						GitRepo: &v1.GitRepoVolumeSource{
							Repository: "https://github.com/kubepack/pack.git",
							Revision:   "dbb56d3cda8130a4dc05fbf76f31a4a4bc61dc89",
						},
					},
				},
			},
		},
	}

	fmt.Println("------------------- POD", pod)
	// repo.Remote()
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube/config"))
	if err != nil {
		return err
	}
	kubeClient := clientset.NewForConfigOrDie(kubeConfig)

	_, err = kubeClient.CoreV1().Pods(metav1.NamespaceDefault).Create(&pod)
	if err != nil {
		fmt.Println("hello pod-----", err)
		err = kubeClient.CoreV1().Pods(metav1.NamespaceDefault).Delete(pod.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
		return err
	}

	return nil
}
