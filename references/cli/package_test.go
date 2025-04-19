package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	pkgdef "github.com/oam-dev/kubevela/pkg/definition"
	common2 "github.com/oam-dev/kubevela/pkg/utils/common"
	"github.com/oam-dev/kubevela/pkg/utils/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	mockPackageName      = "foo"
	defaultNamespaceName = "default"
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
	mockPkgObject := generatePackageObject(defaultNamespaceName, mockPackageName, t)
	// normal test list all packages
	cmd := NewPackageListCommand(c)
	b := bytes.NewBufferString("")
	initCommand(cmd)
	cmd.SetOut(b)
	// create a mock package for testing purposes.
	createPackage(c, mockPkgObject, t)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpeced error when executing list command: %v", err)
	}
	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if !strings.Contains(string(out), mockPackageName) {
		t.Fatalf("expected Package name: %s in output, but got %s", mockPackageName, string(out))
	}

	// test namespace with a package
	b = bytes.NewBufferString("")
	cmd = NewPackageListCommand(c)
	initCommand(cmd)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--namespace", defaultNamespaceName})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpeced error when executing list command: %v", err)
	}
	out, err = io.ReadAll(b)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if !strings.Contains(string(out), mockPackageName) {
		t.Fatalf("expected Package name: %s in output, but got %s", mockPackageName, string(out))
	}

	// test namespace with no package
	b = bytes.NewBufferString("")
	cmd = NewPackageListCommand(c)
	initCommand(cmd)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--namespace", "non-existent-namespace"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpeced error when executing list command: %v", err)
	}
	out, err = io.ReadAll(b)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if !strings.Contains(string(out), "No packages found") {
		t.Fatalf("expected No packages found in output, but got %s", string(out))
	}

	// cleanup
	removePackage(c, mockPkgObject, t)
}

func createPackage(c common2.Args, mockPkgObject pkgdef.Definition, t *testing.T) {

	client, err := c.GetClient()
	if err != nil {
		t.Fatalf("failed to get client: %v", err)
	}
	err = client.Create(context.Background(), &mockPkgObject)
	if err != nil {
		t.Fatalf("failed to create package: %v", err)
	}
}

func removePackage(c common2.Args, mockPkgObject pkgdef.Definition, t *testing.T) {
	client, err := c.GetClient()
	if err != nil {
		t.Fatalf("failed to get client: %v", err)
	}
	err = client.Delete(context.Background(), &mockPkgObject)
	if err != nil {
		t.Fatalf("could not dleete mock package: %v", err)
	}
}

func generatePackageObject(namespace, name string, t *testing.T) pkgdef.Definition {
	pkgObj := pkgdef.Definition{Unstructured: unstructured.Unstructured{}}
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

	err := pkgObj.FromYAML([]byte(p))
	if err != nil {
		t.Fatalf("could not generate package object from yaml: %v", err)
	}
	return pkgObj
}
