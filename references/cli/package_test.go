package cli

import (
	"context"
	"os"
	"strings"
	"testing"

	pkgdef "github.com/oam-dev/kubevela/pkg/definition"
	"github.com/oam-dev/kubevela/pkg/utils"
	common2 "github.com/oam-dev/kubevela/pkg/utils/common"
	"github.com/oam-dev/kubevela/pkg/utils/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)


func TestNewPackageCommandGroup(t *testing.T) {
	cmd := NewPackageCommandGroup(common2.Args{}, "", util.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	initCommand(cmd)
	cmd.SetArgs([]string{"-h"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("failed to execute definition command: %v", err)
	}
}
func TestListPackageCommand(t *testing.T) {
	c := initArgs()
	// normal test	
	cmd := NewPackageListCommand(c)
	initCommand(cmd)
	createPackage(c, "default", "bob", t)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpeced error when executing list command: %v", err)
	}
	// test no package
	cmd = NewPackageListCommand(c)
	initCommand(cmd)
	cmd.SetArgs([]string{"--namespace", "non-existent-namespace"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpeced error when executing list command: %v", err)
	}
	
}

func createPackage(c common2.Args, namespace, name string, t *testing.T) {
	p := `
apiVersion: cue.oam.dev/v1alpha1
kind: Package
metadata:
 name: {{name}} 
 namespace: {{namespace}}
spec:
 path: ext/utils                               
 provider:
   protocol: http
   endpoint: http://httpserverurl:5000       
 templates:
   utils.cue: |
     package utils

     #Sum: {                                   
       #do: "sum"                              
       #provider: {{name}}      
       $params: {
         x: number
         y: number
       }
       $returns: {
         result: number
       }
     }
`
	p = strings.Replace(p, "{{name}}", name, 2)
	p = strings.Replace(p, "{{namespace}}", namespace, 1)
	client, err := c.GetClient()
	if err != nil {
		t.Fatalf("failed to get client: %v", err)
	}
	pkgObject := pkgdef.Definition{Unstructured: unstructured.Unstructured{}}
	pkgObject.FromYAML([]byte(p))
	_, err = utils.CreateOrUpdate(context.Background(), client, &pkgObject)
	if err != nil {
		t.Fatalf("failed to create package: %v", err)
	}
}         