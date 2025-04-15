/*
Copyright 2021 The KubeVela Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cli

import (
	"context"
	"github.com/oam-dev/kubevela/apis/core.oam.dev/v1alpha1"
	"github.com/oam-dev/kubevela/apis/types"
	"github.com/oam-dev/kubevela/pkg/utils/common"
	"github.com/oam-dev/kubevela/pkg/utils/filters"
	"github.com/oam-dev/kubevela/pkg/utils/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PackageDescriptionKey = "package.oam.dev/description"
	group                 = "cue.oam.dev"
)

func NewPackageCommandGroup(c common.Args, order string, ioStreams util.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package",
		Short: "Manage packages.",
		Long:  "Manage Packages for CueX implementation.",
		// TODO: fixme
		Example: "bob",
		Annotations: map[string]string{
			types.TagCommandOrder: order,
			types.TagCommandType:  types.TypeExtension,
		},
	}
	cmd.SetOut(ioStreams.Out)
	cmd.AddCommand(
		NewPackageListCommand(c),
	)
	return cmd
}

func NewPackageListCommand(c common.Args) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List packages",
		Long:  "List packages in kubernetes cluster.",
		Example: "# Command below will list all packages in all namespaces\n" +
			"> vela package list\n" +
			"# Command below will list all packages in bob namespace\n" +
			"> vela package list --namespace bob",
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := cmd.Flags().GetString(FlagNamespace)
			if err != nil {
				return errors.Wrapf(err, "failed to get `%s`", Namespace)
			}

			k8sClient, err := c.GetClient()
			if err != nil {
				return errors.Wrapf(err, "failed to get k8s client")
			}
			// TODO: add filters
			packages, err := SearchPackages(k8sClient, namespace)
			if err != nil {
				return err
			}
			if len(packages) == 0 {
				cmd.Println("No packages found.")
				return nil
			}

			table := newUITable()
			table.AddRow("NAME", "TYPE", "NAMESPACE", "DESCRIPTION")
			for _, pkgs := range packages {
				desc := ""
				if annotations := pkgs.GetAnnotations(); annotations != nil {
					desc = annotations[PackageDescriptionKey]
				}
				table.AddRow(pkgs.GetName(), pkgs.GetKind(), pkgs.GetNamespace(), desc)
			}
			cmd.Println(table)
			return nil
		},
	}
	cmd.Flags().StringP(Namespace, "n", "default", "Specify which namespace the definition locates.")
	return cmd
}

func SearchPackages(c client.Client, namespace string, additionalFilters ...filters.Filter) ([]unstructured.Unstructured, error) {
	ctx := context.Background()
	var kind = "Package"

	var listOptions []client.ListOption
	if namespace != "" {
		listOptions = []client.ListOption{client.InNamespace(namespace)}
	}
	var packages []unstructured.Unstructured
	objs := unstructured.UnstructuredList{}
	objs.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   group,
		Version: v1alpha1.Version,
		Kind:    kind + "List",
	})
	if err := c.List(ctx, &objs, listOptions...); err != nil {
		if meta.IsNoMatchError(err) {
			return nil, errors.Wrapf(err, "crd: %s not installed on the cluster", kind)
		}
		return nil, errors.Wrapf(err, "failed to list %s", kind)
	}

	packages = append(packages, objs.Items...)
	return packages, nil
}
